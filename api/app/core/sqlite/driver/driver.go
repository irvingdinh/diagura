package driver

/*
#cgo CFLAGS: -DSQLITE_THREADSAFE=2
#cgo CFLAGS: -DSQLITE_ENABLE_FTS5
#cgo CFLAGS: -DSQLITE_ENABLE_JSON1
#cgo CFLAGS: -DSQLITE_DEFAULT_WAL_SYNCHRONOUS=1
#cgo CFLAGS: -DSQLITE_DEFAULT_FOREIGN_KEYS=1
#cgo CFLAGS: -DSQLITE_DQS=0
#cgo CFLAGS: -DSQLITE_LIKE_DOESNT_MATCH_BLOBS
#cgo CFLAGS: -DSQLITE_USE_ALLOCA
#cgo darwin LDFLAGS: -lm
#cgo linux LDFLAGS: -lm -ldl -lpthread

#include "cgo.h"
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

// Column types returned by SQLite.
const (
	TypeInteger = int(C.SQLITE_INTEGER)
	TypeFloat   = int(C.SQLITE_FLOAT)
	TypeText    = int(C.SQLITE3_TEXT)
	TypeBlob    = int(C.SQLITE_BLOB)
	TypeNull    = int(C.SQLITE_NULL)
)

// Open flags.
const (
	OpenReadWrite = int(C.SQLITE_OPEN_READWRITE)
	OpenCreate    = int(C.SQLITE_OPEN_CREATE)
	OpenReadOnly  = int(C.SQLITE_OPEN_READONLY)
	OpenNoMutex   = int(C.SQLITE_OPEN_NOMUTEX)
)

const (
	resultOK   = C.SQLITE_OK
	resultRow  = C.SQLITE_ROW
	resultDone = C.SQLITE_DONE
)

// Conn is a single SQLite3 connection. Not safe for concurrent use.
type Conn struct {
	db *C.sqlite3
}

// Open opens a SQLite database at the given path with the specified flags.
func Open(path string, flags int) (*Conn, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var db *C.sqlite3
	rc := C._open(cpath, &db, C.int(flags))
	if rc != resultOK {
		var msg string
		if db != nil {
			msg = C.GoString(C._errmsg(db))
			C._close(db)
		}
		return nil, fmt.Errorf("sqlite: open %s: %s (code %d)", path, msg, rc)
	}
	return &Conn{db: db}, nil
}

// Close closes the connection.
func (c *Conn) Close() error {
	if c.db == nil {
		return nil
	}
	rc := C._close(c.db)
	c.db = nil
	if rc != resultOK {
		return fmt.Errorf("sqlite: close: code %d", rc)
	}
	return nil
}

// Exec executes a SQL statement without parameters. Use for PRAGMAs
// and DDL statements.
func (c *Conn) Exec(sql string) error {
	csql := C.CString(sql)
	defer C.free(unsafe.Pointer(csql))

	var errmsg *C.char
	rc := C._exec(c.db, csql, &errmsg)
	if rc != resultOK {
		msg := C.GoString(errmsg)
		C.free(unsafe.Pointer(errmsg))
		return &Error{
			Code:         int(C._errcode(c.db)),
			ExtendedCode: int(C._extended_errcode(c.db)),
			Message:      msg,
		}
	}
	return nil
}

// SetBusyTimeout configures the busy timeout in milliseconds.
func (c *Conn) SetBusyTimeout(ms int) error {
	rc := C._busy_timeout(c.db, C.int(ms))
	if rc != resultOK {
		msg := C.GoString(C._errmsg(c.db))
		return fmt.Errorf("sqlite: busy_timeout: %s", msg)
	}
	return nil
}

// Prepare compiles a SQL statement for execution.
func (c *Conn) Prepare(sql string) (*Stmt, error) {
	csql := C.CString(sql)
	defer C.free(unsafe.Pointer(csql))

	var stmt *C.sqlite3_stmt
	rc := C._prepare(c.db, csql, C.int(len(sql)), &stmt, nil)
	if rc != resultOK {
		return nil, c.makeError("sqlite: prepare")
	}
	if stmt == nil {
		return nil, fmt.Errorf("sqlite: prepare: empty statement")
	}
	return &Stmt{stmt: stmt, conn: c}, nil
}

// Changes returns the number of rows modified by the most recent
// INSERT, UPDATE, or DELETE.
func (c *Conn) Changes() int64 {
	return int64(C._changes(c.db))
}

// LastInsertRowID returns the rowid of the most recent INSERT.
func (c *Conn) LastInsertRowID() int64 {
	return int64(C._last_insert_rowid(c.db))
}

func (c *Conn) makeError(prefix string) *Error {
	code := int(C._errcode(c.db))
	extCode := int(C._extended_errcode(c.db))
	msg := C.GoString(C._errmsg(c.db))
	if prefix != "" {
		msg = prefix + ": " + msg
	}
	return &Error{Code: code, ExtendedCode: extCode, Message: msg}
}

// Stmt is a prepared statement.
type Stmt struct {
	stmt *C.sqlite3_stmt
	conn *Conn
}

// Bind binds arguments to the prepared statement. Supported types:
// nil, int, int32, int64, float32, float64, bool, string, []byte, time.Time.
func (s *Stmt) Bind(args ...any) error {
	for i, arg := range args {
		col := C.int(i + 1) // SQLite parameters are 1-indexed
		var rc C.int

		switch v := arg.(type) {
		case nil:
			rc = C._bind_null(s.stmt, col)
		case int:
			rc = C._bind_int64(s.stmt, col, C.longlong(v))
		case int32:
			rc = C._bind_int64(s.stmt, col, C.longlong(v))
		case int64:
			rc = C._bind_int64(s.stmt, col, C.longlong(v))
		case float32:
			rc = C._bind_double(s.stmt, col, C.double(v))
		case float64:
			rc = C._bind_double(s.stmt, col, C.double(v))
		case bool:
			if v {
				rc = C._bind_int64(s.stmt, col, 1)
			} else {
				rc = C._bind_int64(s.stmt, col, 0)
			}
		case string:
			cs := C.CString(v)
			rc = C._bind_text(s.stmt, col, cs, C.int(len(v)))
			C.free(unsafe.Pointer(cs))
		case []byte:
			if len(v) == 0 {
				rc = C._bind_blob(s.stmt, col, nil, 0)
			} else {
				rc = C._bind_blob(s.stmt, col, unsafe.Pointer(&v[0]), C.int(len(v)))
			}
		case time.Time:
			ts := v.UTC().Format("2006-01-02 15:04:05.000")
			cs := C.CString(ts)
			rc = C._bind_text(s.stmt, col, cs, C.int(len(ts)))
			C.free(unsafe.Pointer(cs))
		default:
			return fmt.Errorf("sqlite: bind: unsupported type %T at index %d", arg, i)
		}

		if rc != resultOK {
			return s.conn.makeError(fmt.Sprintf("sqlite: bind %d", i))
		}
	}
	return nil
}

// Step advances the statement. Returns true if a row is available
// (SQLITE_ROW), false if done (SQLITE_DONE).
func (s *Stmt) Step() (bool, error) {
	rc := C._step(s.stmt)
	switch rc {
	case resultRow:
		return true, nil
	case resultDone:
		return false, nil
	default:
		return false, s.conn.makeError("sqlite: step")
	}
}

// Reset resets the statement for re-execution.
func (s *Stmt) Reset() error {
	rc := C._reset(s.stmt)
	if rc != resultOK {
		return s.conn.makeError("sqlite: reset")
	}
	return nil
}

// ClearBindings clears all parameter bindings.
func (s *Stmt) ClearBindings() error {
	rc := C._clear_bindings(s.stmt)
	if rc != resultOK {
		return s.conn.makeError("sqlite: clear bindings")
	}
	return nil
}

// Finalize destroys the prepared statement.
func (s *Stmt) Finalize() error {
	if s.stmt == nil {
		return nil
	}
	rc := C._finalize(s.stmt)
	s.stmt = nil
	if rc != resultOK {
		return s.conn.makeError("sqlite: finalize")
	}
	return nil
}

// ColumnCount returns the number of columns in the result set.
func (s *Stmt) ColumnCount() int {
	return int(C._column_count(s.stmt))
}

// ColumnName returns the name of the i-th column.
func (s *Stmt) ColumnName(i int) string {
	return C.GoString(C._column_name(s.stmt, C.int(i)))
}

// ColumnType returns the SQLite type of the i-th column in the current row.
func (s *Stmt) ColumnType(i int) int {
	return int(C._column_type(s.stmt, C.int(i)))
}

// ColumnInt64 returns the i-th column as an int64.
func (s *Stmt) ColumnInt64(i int) int64 {
	return int64(C._column_int64(s.stmt, C.int(i)))
}

// ColumnFloat64 returns the i-th column as a float64.
func (s *Stmt) ColumnFloat64(i int) float64 {
	return float64(C._column_double(s.stmt, C.int(i)))
}

// ColumnText returns the i-th column as a string.
func (s *Stmt) ColumnText(i int) string {
	return C.GoString(C._column_text(s.stmt, C.int(i)))
}

// ColumnBlob returns the i-th column as a byte slice.
func (s *Stmt) ColumnBlob(i int) []byte {
	n := C._column_bytes(s.stmt, C.int(i))
	if n == 0 {
		return nil
	}
	src := C._column_blob(s.stmt, C.int(i))
	buf := make([]byte, int(n))
	copy(buf, unsafe.Slice((*byte)(src), int(n)))
	return buf
}
