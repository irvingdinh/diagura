package migration

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"localhost/app/core/sqlite/driver"
)

// Migration represents a parsed migration file.
type Migration struct {
	Version int
	Name    string
	Up      []string
	Down    []string
}

// MigrationStatus describes the state of a migration.
type MigrationStatus struct {
	Version   int        `json:"version"`
	Name      string     `json:"name"`
	Applied   bool       `json:"applied"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
}

// Engine manages schema migrations. It operates directly on a
// *driver.Conn (the write connection) to avoid import cycles.
type Engine struct {
	conn       *driver.Conn
	migrations []Migration
}

// NewEngine creates a new migration engine that operates on the given
// write connection.
func NewEngine(conn *driver.Conn) *Engine {
	return &Engine{conn: conn}
}

// Collect reads .sql migration files from the given file systems.
// Files must be named NNNNN_descriptive_name.sql.
func (e *Engine) Collect(sources ...fs.FS) error {
	for _, src := range sources {
		err := fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".sql" {
				return nil
			}

			version, name, err := parseFilename(filepath.Base(path))
			if err != nil {
				return fmt.Errorf("migration: %s: %w", path, err)
			}

			f, err := src.Open(path)
			if err != nil {
				return fmt.Errorf("migration: open %s: %w", path, err)
			}
			defer func() { _ = f.Close() }()

			parsed, err := parse(f)
			if err != nil {
				return fmt.Errorf("migration: parse %s: %w", path, err)
			}

			e.migrations = append(e.migrations, Migration{
				Version: version,
				Name:    name,
				Up:      parsed.Up,
				Down:    parsed.Down,
			})
			return nil
		})
		if err != nil {
			return err
		}
	}

	sort.Slice(e.migrations, func(i, j int) bool {
		return e.migrations[i].Version < e.migrations[j].Version
	})

	for i := 1; i < len(e.migrations); i++ {
		if e.migrations[i].Version == e.migrations[i-1].Version {
			return fmt.Errorf("migration: duplicate version %d", e.migrations[i].Version)
		}
	}

	return nil
}

// Up applies all pending migrations. Returns the count of applied migrations.
func (e *Engine) Up(_ context.Context) (int, error) {
	if err := e.ensureTable(); err != nil {
		return 0, err
	}

	applied, err := e.appliedVersions()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range e.migrations {
		if applied[m.Version] {
			continue
		}

		if err := e.applyUp(m); err != nil {
			return count, fmt.Errorf("migration: version %d (%s): %w", m.Version, m.Name, err)
		}
		count++
	}

	return count, nil
}

// Down rolls back the last count applied migrations.
func (e *Engine) Down(_ context.Context, count int) (int, error) {
	if err := e.ensureTable(); err != nil {
		return 0, err
	}

	applied, err := e.appliedOrdered()
	if err != nil {
		return 0, err
	}

	rolled := 0
	for i := len(applied) - 1; i >= 0 && rolled < count; i-- {
		version := applied[i]
		m := e.findMigration(version)
		if m == nil {
			return rolled, fmt.Errorf("migration: version %d not found in collected migrations", version)
		}
		if len(m.Down) == 0 {
			return rolled, fmt.Errorf("migration: version %d (%s) has no down migration", m.Version, m.Name)
		}

		if err := e.applyDown(*m); err != nil {
			return rolled, fmt.Errorf("migration: rollback version %d (%s): %w", m.Version, m.Name, err)
		}
		rolled++
	}

	return rolled, nil
}

// Version returns the highest applied migration version, or 0 if none.
func (e *Engine) Version(_ context.Context) (int, error) {
	if err := e.ensureTable(); err != nil {
		return 0, err
	}

	stmt, err := e.conn.Prepare("SELECT COALESCE(MAX(version), 0) FROM _migrations")
	if err != nil {
		return 0, fmt.Errorf("migration: version: %w", err)
	}
	defer func() { _ = stmt.Finalize() }()

	hasRow, err := stmt.Step()
	if err != nil {
		return 0, fmt.Errorf("migration: version: %w", err)
	}
	if !hasRow {
		return 0, nil
	}
	return int(stmt.ColumnInt64(0)), nil
}

// Status returns the status of all known migrations.
func (e *Engine) Status(_ context.Context) ([]MigrationStatus, error) {
	if err := e.ensureTable(); err != nil {
		return nil, err
	}

	appliedMap := make(map[int]time.Time)
	stmt, err := e.conn.Prepare("SELECT version, applied_at FROM _migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("migration: status: %w", err)
	}
	defer func() { _ = stmt.Finalize() }()

	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}
		version := int(stmt.ColumnInt64(0))
		appliedAt := stmt.ColumnText(1)
		t, _ := time.Parse(time.RFC3339, appliedAt)
		appliedMap[version] = t
	}

	result := make([]MigrationStatus, 0, len(e.migrations))
	for _, m := range e.migrations {
		s := MigrationStatus{Version: m.Version, Name: m.Name}
		if t, ok := appliedMap[m.Version]; ok {
			s.Applied = true
			s.AppliedAt = &t
		}
		result = append(result, s)
	}
	return result, nil
}

func (e *Engine) ensureTable() error {
	return e.conn.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`)
}

func (e *Engine) appliedVersions() (map[int]bool, error) {
	stmt, err := e.conn.Prepare("SELECT version FROM _migrations")
	if err != nil {
		return nil, fmt.Errorf("migration: list applied: %w", err)
	}
	defer func() { _ = stmt.Finalize() }()

	applied := make(map[int]bool)
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}
		applied[int(stmt.ColumnInt64(0))] = true
	}
	return applied, nil
}

func (e *Engine) appliedOrdered() ([]int, error) {
	stmt, err := e.conn.Prepare("SELECT version FROM _migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("migration: list applied: %w", err)
	}
	defer func() { _ = stmt.Finalize() }()

	var versions []int
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}
		versions = append(versions, int(stmt.ColumnInt64(0)))
	}
	return versions, nil
}

func (e *Engine) applyUp(m Migration) error {
	if err := e.conn.Exec("BEGIN IMMEDIATE"); err != nil {
		return err
	}

	for _, s := range m.Up {
		if err := e.conn.Exec(s); err != nil {
			_ = e.conn.Exec("ROLLBACK")
			return err
		}
	}

	if err := e.execWithArgs(
		"INSERT INTO _migrations (version, name, applied_at) VALUES (?, ?, ?)",
		m.Version, m.Name, time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		_ = e.conn.Exec("ROLLBACK")
		return err
	}

	return e.conn.Exec("COMMIT")
}

func (e *Engine) applyDown(m Migration) error {
	if err := e.conn.Exec("BEGIN IMMEDIATE"); err != nil {
		return err
	}

	for _, s := range m.Down {
		if err := e.conn.Exec(s); err != nil {
			_ = e.conn.Exec("ROLLBACK")
			return err
		}
	}

	if err := e.execWithArgs(
		"DELETE FROM _migrations WHERE version = ?",
		m.Version,
	); err != nil {
		_ = e.conn.Exec("ROLLBACK")
		return err
	}

	return e.conn.Exec("COMMIT")
}

// execWithArgs prepares, binds, and steps a statement with arguments.
func (e *Engine) execWithArgs(sql string, args ...any) error {
	stmt, err := e.conn.Prepare(sql)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Finalize() }()

	if err := stmt.Bind(args...); err != nil {
		return err
	}
	_, err = stmt.Step()
	return err
}

func (e *Engine) findMigration(version int) *Migration {
	for i := range e.migrations {
		if e.migrations[i].Version == version {
			return &e.migrations[i]
		}
	}
	return nil
}

func parseFilename(name string) (int, string, error) {
	base := strings.TrimSuffix(name, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration filename: %s (expected NNNNN_name.sql)", name)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid migration version in %s: %w", name, err)
	}

	return version, parts[1], nil
}
