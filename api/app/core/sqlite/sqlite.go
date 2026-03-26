package sqlite

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"go.uber.org/fx"

	"localhost/app/core/config"
	"localhost/app/core/config/rule"
	"localhost/app/core/sqlite/driver"
	"localhost/app/core/sqlite/migration"
)

// ErrNoRows is returned when a query returns no rows.
var ErrNoRows = errors.New("sqlite: no rows")

// ErrTxFinished is returned when Rows from a transaction are used after
// Commit or Rollback.
var ErrTxFinished = errors.New("sqlite: transaction already finished")

// Result holds the outcome of an Exec call.
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// cachedConn wraps a driver.Conn with a prepared statement cache.
type cachedConn struct {
	conn  *driver.Conn
	cache map[string]*driver.Stmt
}

func newCachedConn(conn *driver.Conn) *cachedConn {
	return &cachedConn{conn: conn, cache: make(map[string]*driver.Stmt)}
}

// prepare returns a cached statement or prepares a new one.
func (cc *cachedConn) prepare(sql string) (*driver.Stmt, error) {
	if stmt, ok := cc.cache[sql]; ok {
		if err := stmt.Reset(); err == nil {
			if err := stmt.ClearBindings(); err == nil {
				return stmt, nil
			}
		}
		// Reset or ClearBindings failed; discard and re-prepare.
		delete(cc.cache, sql)
		_ = stmt.Finalize()
	}

	stmt, err := cc.conn.Prepare(sql)
	if err != nil {
		return nil, err
	}
	cc.cache[sql] = stmt
	return stmt, nil
}

// close finalizes all cached statements and closes the connection.
func (cc *cachedConn) close() error {
	for _, stmt := range cc.cache {
		_ = stmt.Finalize()
	}
	cc.cache = nil
	return cc.conn.Close()
}

// DB is the primary database handle. Write operations use a single
// mutex-protected connection. Read operations use a pool of connections.
type DB struct {
	mu           sync.Mutex
	write        *cachedConn
	readPool     chan *cachedConn
	poolDone     chan struct{}
	poolWg       sync.WaitGroup
	poolStopping atomic.Bool
	path         string
	busyTimeout  int
	readPoolSize int
	migrationFS  fs.FS
}

// Provide returns the fx.Option that registers *DB into the DI container.
// It reads config, opens connections, runs migrations, and hooks lifecycle.
func Provide(migrationFS fs.FS) fx.Option {
	return fx.Options(
		fx.Provide(func(lc fx.Lifecycle) *DB {
			defaultPool := runtime.GOMAXPROCS(0)

			config.SetDefaults(config.Values{
				"db.busy_timeout": 5000,
				"db.read_pool":    defaultPool,
			})
			config.SetRule("db.busy_timeout", rule.Required, rule.Positive)
			config.SetRule("db.read_pool", rule.Required, rule.Positive)

			dataDir := config.GetStringOr("data_dir", defaultDataDir())
			path := filepath.Join(dataDir, "database.sqlite")

			db := &DB{
				path:         path,
				busyTimeout:  config.GetIntOr("db.busy_timeout", 5000),
				readPoolSize: config.GetIntOr("db.read_pool", defaultPool),
				migrationFS:  migrationFS,
			}

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return db.open(ctx)
				},
				OnStop: func(_ context.Context) error {
					return db.Close()
				},
			})

			return db
		}),
		fx.Invoke(func(_ *DB) {}), // force construction before config.Validate
	)
}

func (db *DB) open(ctx context.Context) error {
	// Open write connection.
	writeConn, err := db.openConn(driver.OpenReadWrite | driver.OpenCreate | driver.OpenNoMutex)
	if err != nil {
		return err
	}

	// Write-only PRAGMAs.
	if err := writeConn.Exec("PRAGMA wal_autocheckpoint = 1000"); err != nil {
		_ = writeConn.Close()
		return fmt.Errorf("sqlite: wal_autocheckpoint: %w", err)
	}

	// Run migrations on the raw write connection before wrapping.
	if db.migrationFS != nil {
		engine := migration.NewEngine(writeConn)
		if err := engine.Collect(db.migrationFS); err != nil {
			_ = writeConn.Close()
			return fmt.Errorf("sqlite: collect migrations: %w", err)
		}
		if _, err := engine.Up(ctx); err != nil {
			_ = writeConn.Close()
			return err
		}
	}

	db.write = newCachedConn(writeConn)

	// Open read pool.
	pool := make(chan *cachedConn, db.readPoolSize)
	for range db.readPoolSize {
		rc, err := db.openConn(driver.OpenReadOnly | driver.OpenNoMutex)
		if err != nil {
			close(pool)
			for conn := range pool {
				_ = conn.close()
			}
			_ = db.write.close()
			db.write = nil
			return fmt.Errorf("sqlite: open read connection: %w", err)
		}
		pool <- newCachedConn(rc)
	}
	db.readPool = pool
	db.poolDone = make(chan struct{})

	return nil
}

func (db *DB) openConn(flags int) (*driver.Conn, error) {
	conn, err := driver.Open(db.path, flags)
	if err != nil {
		return nil, err
	}

	if err := conn.SetBusyTimeout(db.busyTimeout); err != nil {
		_ = conn.Close()
		return nil, err
	}

	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 268435456",
		"PRAGMA cache_size = -16000",
	}
	for _, p := range pragmas {
		if err := conn.Exec(p); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("sqlite: %s: %w", p, err)
		}
	}

	return conn, nil
}

// Close closes all database connections.
func (db *DB) Close() error {
	if db.readPool != nil {
		close(db.poolDone)
		db.poolStopping.Store(true)
		db.poolWg.Wait()

		close(db.readPool)
		for conn := range db.readPool {
			_ = conn.close()
		}
		db.readPool = nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.write == nil {
		return nil
	}

	err := db.write.close()
	db.write = nil
	return err
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// Exec executes a SQL statement that does not return rows. Uses the
// write connection.
func (db *DB) Exec(sql string, args ...any) (Result, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.write == nil {
		return Result{}, fmt.Errorf("sqlite: database not open")
	}

	stmt, err := db.write.prepare(sql)
	if err != nil {
		return Result{}, err
	}

	if err := stmt.Bind(args...); err != nil {
		return Result{}, err
	}

	if _, err := stmt.Step(); err != nil {
		return Result{}, err
	}

	return Result{
		LastInsertID: db.write.conn.LastInsertRowID(),
		RowsAffected: db.write.conn.Changes(),
	}, nil
}

// Query executes a SQL statement that returns rows. Uses a connection
// from the read pool. The caller must call Rows.Close when done.
func (db *DB) Query(sql string, args ...any) (*Rows, error) {
	if db.readPool != nil {
		return db.queryPool(sql, args)
	}
	return db.queryMu(sql, args)
}

func (db *DB) queryPool(sql string, args []any) (*Rows, error) {
	db.poolWg.Add(1)

	var conn *cachedConn
	select {
	case conn = <-db.readPool:
	default:
		select {
		case conn = <-db.readPool:
		case <-db.poolDone:
			db.poolWg.Done()
			return nil, fmt.Errorf("sqlite: database is shutting down")
		}
	}

	stmt, err := conn.prepare(sql)
	if err != nil {
		db.returnToPool(conn)
		return nil, err
	}

	if err := stmt.Bind(args...); err != nil {
		db.returnToPool(conn)
		return nil, err
	}

	return &Rows{
		db:   db,
		conn: conn,
		pool: db.readPool,
		stmt: stmt,
	}, nil
}

func (db *DB) queryMu(sql string, args []any) (*Rows, error) {
	db.mu.Lock()

	if db.write == nil {
		db.mu.Unlock()
		return nil, fmt.Errorf("sqlite: database not open")
	}

	stmt, err := db.write.prepare(sql)
	if err != nil {
		db.mu.Unlock()
		return nil, err
	}

	if err := stmt.Bind(args...); err != nil {
		db.mu.Unlock()
		return nil, err
	}

	return &Rows{
		db:   db,
		conn: nil,
		mu:   &db.mu,
		stmt: stmt,
	}, nil
}

func (db *DB) returnToPool(conn *cachedConn) {
	if db.poolStopping.Load() {
		_ = conn.close()
	} else {
		select {
		case db.readPool <- conn:
		case <-db.poolDone:
			_ = conn.close()
		}
	}
	db.poolWg.Done()
}

// QueryRow executes a SQL statement that returns at most one row.
func (db *DB) QueryRow(sql string, args ...any) *Row {
	rows, err := db.Query(sql, args...)
	if err != nil {
		return &Row{err: err}
	}
	return &Row{rows: rows}
}

// Begin starts a new transaction. The write mutex is held until
// Commit or Rollback is called.
func (db *DB) Begin() (*Tx, error) {
	db.mu.Lock()

	if db.write == nil {
		db.mu.Unlock()
		return nil, fmt.Errorf("sqlite: database not open")
	}

	stmt, err := db.write.prepare("BEGIN IMMEDIATE")
	if err != nil {
		db.mu.Unlock()
		return nil, fmt.Errorf("sqlite: begin: %w", err)
	}
	if _, err := stmt.Step(); err != nil {
		db.mu.Unlock()
		return nil, fmt.Errorf("sqlite: begin: %w", err)
	}

	return &Tx{db: db}, nil
}

// InTx runs fn inside a transaction. If fn returns nil, the
// transaction is committed. If fn returns an error or panics,
// the transaction is rolled back.
func (db *DB) InTx(fn func(*Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Rows iterates over query results.
type Rows struct {
	db         *DB
	conn       *cachedConn      // read pool connection (nil if mutex-held)
	pool       chan *cachedConn // pool to return conn to (nil if mutex-held)
	mu         *sync.Mutex      // mutex to unlock on Close (nil if pooled)
	stmt       *driver.Stmt
	txFinished *bool // non-nil for Tx-originated Rows; points to Tx.finished
	cols       []string
	err        error
	closed     bool
	hasRow     bool
}

// Next advances to the next row.
func (r *Rows) Next() bool {
	if r.closed {
		return false
	}
	if r.txFinished != nil && *r.txFinished {
		r.err = ErrTxFinished
		return false
	}
	hasRow, err := r.stmt.Step()
	if err != nil {
		r.err = err
		return false
	}
	r.hasRow = hasRow
	return hasRow
}

// Scan reads the current row's columns into dest pointers.
// Supported types: *int, *int32, *int64, *float32, *float64,
// *string, *[]byte, *bool, *any.
func (r *Rows) Scan(dest ...any) error {
	if r.txFinished != nil && *r.txFinished {
		return ErrTxFinished
	}
	if !r.hasRow {
		return fmt.Errorf("sqlite: scan: no row")
	}

	n := r.stmt.ColumnCount()
	if len(dest) != n {
		return fmt.Errorf("sqlite: scan: expected %d columns, got %d destinations", n, len(dest))
	}

	for i, d := range dest {
		colType := r.stmt.ColumnType(i)

		switch ptr := d.(type) {
		case *int:
			*ptr = int(r.stmt.ColumnInt64(i))
		case *int32:
			*ptr = int32(r.stmt.ColumnInt64(i))
		case *int64:
			*ptr = int64(r.stmt.ColumnInt64(i))
		case *float32:
			*ptr = float32(r.stmt.ColumnFloat64(i))
		case *float64:
			*ptr = r.stmt.ColumnFloat64(i)
		case *string:
			*ptr = r.stmt.ColumnText(i)
		case *[]byte:
			if colType == driver.TypeNull {
				*ptr = nil
			} else {
				*ptr = r.stmt.ColumnBlob(i)
			}
		case *bool:
			*ptr = r.stmt.ColumnInt64(i) != 0
		case *any:
			switch colType {
			case driver.TypeInteger:
				*ptr = r.stmt.ColumnInt64(i)
			case driver.TypeFloat:
				*ptr = r.stmt.ColumnFloat64(i)
			case driver.TypeText:
				*ptr = r.stmt.ColumnText(i)
			case driver.TypeBlob:
				*ptr = r.stmt.ColumnBlob(i)
			case driver.TypeNull:
				*ptr = nil
			}
		default:
			return fmt.Errorf("sqlite: scan: unsupported type %T at index %d", d, i)
		}
	}
	return nil
}

// Columns returns the column names for the result set.
func (r *Rows) Columns() []string {
	if r.cols != nil {
		return r.cols
	}
	n := r.stmt.ColumnCount()
	r.cols = make([]string, n)
	for i := range n {
		r.cols[i] = r.stmt.ColumnName(i)
	}
	return r.cols
}

// Err returns the error encountered during iteration.
func (r *Rows) Err() error {
	return r.err
}

// Close releases the prepared statement and returns the connection.
func (r *Rows) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	_ = r.stmt.Reset()
	if r.pool != nil {
		r.db.returnToPool(r.conn)
	} else if r.mu != nil {
		r.mu.Unlock()
	}
	return nil
}

// Row is the result of QueryRow.
type Row struct {
	rows *Rows
	err  error
}

// Scan reads the first row into dest and closes the underlying Rows.
// Returns ErrNoRows if the query returned no rows.
func (r *Row) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	defer func() { _ = r.rows.Close() }()

	if !r.rows.Next() {
		return ErrNoRows
	}
	return r.rows.Scan(dest...)
}

// Tx is a database transaction. Holds the write mutex until Commit
// or Rollback.
type Tx struct {
	db       *DB
	finished bool
}

// Commit commits the transaction and releases the write mutex.
func (tx *Tx) Commit() error {
	if tx.finished {
		return fmt.Errorf("sqlite: transaction already finished")
	}
	tx.finished = true
	err := tx.execLocked("COMMIT")
	tx.db.mu.Unlock()
	if err != nil {
		return fmt.Errorf("sqlite: commit: %w", err)
	}
	return nil
}

// Rollback aborts the transaction and releases the write mutex.
// Safe to call after Commit (no-op).
func (tx *Tx) Rollback() error {
	if tx.finished {
		return nil
	}
	tx.finished = true
	err := tx.execLocked("ROLLBACK")
	tx.db.mu.Unlock()
	if err != nil {
		return fmt.Errorf("sqlite: rollback: %w", err)
	}
	return nil
}

// Exec executes a SQL statement within the transaction.
func (tx *Tx) Exec(sql string, args ...any) (Result, error) {
	if tx.finished {
		return Result{}, fmt.Errorf("sqlite: transaction already finished")
	}

	stmt, err := tx.db.write.prepare(sql)
	if err != nil {
		return Result{}, err
	}

	if err := stmt.Bind(args...); err != nil {
		return Result{}, err
	}

	if _, err := stmt.Step(); err != nil {
		return Result{}, err
	}

	return Result{
		LastInsertID: tx.db.write.conn.LastInsertRowID(),
		RowsAffected: tx.db.write.conn.Changes(),
	}, nil
}

// Query executes a query within the transaction.
func (tx *Tx) Query(sql string, args ...any) (*Rows, error) {
	if tx.finished {
		return nil, fmt.Errorf("sqlite: transaction already finished")
	}

	stmt, err := tx.db.write.prepare(sql)
	if err != nil {
		return nil, err
	}

	if err := stmt.Bind(args...); err != nil {
		return nil, err
	}

	return &Rows{
		db:         tx.db,
		stmt:       stmt,
		txFinished: &tx.finished,
	}, nil
}

// QueryRow executes a query that returns at most one row within the
// transaction.
func (tx *Tx) QueryRow(sql string, args ...any) *Row {
	rows, err := tx.Query(sql, args...)
	if err != nil {
		return &Row{err: err}
	}
	return &Row{rows: rows}
}

func (tx *Tx) execLocked(sql string) error {
	stmt, err := tx.db.write.prepare(sql)
	if err != nil {
		return err
	}
	_, err = stmt.Step()
	return err
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".standalone"
	}
	return filepath.Join(home, ".standalone")
}
