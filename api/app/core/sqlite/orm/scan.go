package orm

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"localhost/app/core/sqlite"
)

// Querier is satisfied by *sqlite.DB and *sqlite.Tx.
type Querier interface {
	Query(sql string, args ...any) (*sqlite.Rows, error)
}

// QueryAll executes a query and scans all rows into []T using db struct tags.
// Always returns a non-nil slice (empty when no rows match).
func QueryAll[T any](q Querier, sql string, args ...any) ([]T, error) {
	rows, err := q.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	columns := rows.Columns()
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	m := getMapping(t)

	dests := make([]any, len(columns))
	for i := range dests {
		dests[i] = new(any)
	}

	var result []T
	for rows.Next() {
		item, err := scanRow[T](rows, columns, m, dests)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]T, 0)
	}
	return result, nil
}

// QueryOne executes a query and scans exactly one row into T.
// Returns ErrNotFound if no rows match.
func QueryOne[T any](q Querier, sql string, args ...any) (T, error) {
	var zero T

	rows, err := q.Query(sql, args...)
	if err != nil {
		return zero, err
	}
	defer func() { _ = rows.Close() }()

	columns := rows.Columns()
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	m := getMapping(t)

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return zero, err
		}
		return zero, ErrNotFound
	}

	dests := make([]any, len(columns))
	for i := range dests {
		dests[i] = new(any)
	}
	return scanRow[T](rows, columns, m, dests)
}

// QueryVal executes a query and scans a single scalar value from the first
// row. Returns ErrNotFound if no rows match.
func QueryVal[T any](q Querier, sql string, args ...any) (T, error) {
	var zero T

	rows, err := q.Query(sql, args...)
	if err != nil {
		return zero, err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return zero, err
		}
		return zero, ErrNotFound
	}

	var val T
	if err := rows.Scan(&val); err != nil {
		return zero, err
	}
	return val, nil
}

// ---------------------------------------------------------------------------
// Struct-tag mapping cache
// ---------------------------------------------------------------------------

type fieldMapping struct {
	colToIndex map[string][]int
}

var mappingCache sync.Map

func getMapping(t reflect.Type) *fieldMapping {
	if v, ok := mappingCache.Load(t); ok {
		return v.(*fieldMapping)
	}
	m := &fieldMapping{colToIndex: make(map[string][]int)}
	buildMapping(t, nil, m)
	mappingCache.Store(t, m)
	return m
}

func buildMapping(t reflect.Type, prefix []int, m *fieldMapping) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		idx := append(append([]int{}, prefix...), i)

		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				buildMapping(ft, idx, m)
				continue
			}
		}

		tag := f.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		m.colToIndex[tag] = idx
	}
}

// ---------------------------------------------------------------------------
// Row scanning
// ---------------------------------------------------------------------------

func scanRow[T any](rows *sqlite.Rows, columns []string, m *fieldMapping, dests []any) (T, error) {
	var item T
	v := reflect.ValueOf(&item).Elem()

	if err := rows.Scan(dests...); err != nil {
		return item, fmt.Errorf("orm: scan: %w", err)
	}

	for i, colName := range columns {
		idx, ok := m.colToIndex[colName]
		if !ok {
			continue
		}
		rawVal := *(dests[i].(*any))
		field := v.FieldByIndex(idx)
		if err := setField(field, rawVal); err != nil {
			return item, fmt.Errorf("orm: set field %q: %w", colName, err)
		}
	}

	return item, nil
}

func setField(field reflect.Value, rawVal any) error {
	if rawVal == nil {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
		}
		return nil
	}

	fieldType := field.Type()

	// Pointer fields: allocate and set inner value.
	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()
		ptr := reflect.New(elemType)
		if err := setField(ptr.Elem(), rawVal); err != nil {
			return err
		}
		field.Set(ptr)
		return nil
	}

	// Bool fields: SQLite stores booleans as INTEGER 0/1.
	if fieldType.Kind() == reflect.Bool {
		switch v := rawVal.(type) {
		case int64:
			field.SetBool(v != 0)
			return nil
		case bool:
			field.SetBool(v)
			return nil
		}
	}

	// time.Time fields: parse from TEXT string.
	if fieldType == reflect.TypeOf(time.Time{}) {
		switch v := rawVal.(type) {
		case string:
			t, err := time.Parse(timeFormat, v)
			if err != nil {
				t, err = time.Parse(timeFormatMs, v)
				if err != nil {
					t, err = time.Parse(time.RFC3339, v)
					if err != nil {
						return fmt.Errorf("parse time %q: %w", v, err)
					}
				}
			}
			field.Set(reflect.ValueOf(t))
			return nil
		case time.Time:
			field.Set(reflect.ValueOf(v))
			return nil
		}
	}

	rv := reflect.ValueOf(rawVal)
	if rv.Type().AssignableTo(fieldType) {
		field.Set(rv)
		return nil
	}
	if rv.Type().ConvertibleTo(fieldType) {
		field.Set(rv.Convert(fieldType))
		return nil
	}

	return fmt.Errorf("cannot convert %T to %s", rawVal, fieldType)
}
