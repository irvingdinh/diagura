package orm

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// SELECT
// ---------------------------------------------------------------------------

// SelectBuilder builds SELECT queries.
type SelectBuilder struct {
	columns  []string
	table    string
	wheres   []whereClause
	joins    []joinClause
	groupBys []string
	havings  []whereClause
	orderBys []orderByClause
	limit    int
	offset   int
}

type whereClause struct {
	cond string
	args []any
}

type joinClause struct {
	kind  string // "JOIN" or "LEFT JOIN"
	table string
	on    string
}

type orderByClause struct {
	column string
	dir    string
}

// Select starts building a SELECT query with the given columns.
func Select(columns ...string) *SelectBuilder {
	return &SelectBuilder{columns: columns, limit: -1, offset: -1}
}

func (b *SelectBuilder) From(table string) *SelectBuilder {
	b.table = table
	return b
}

func (b *SelectBuilder) Where(cond string, args ...any) *SelectBuilder {
	b.wheres = append(b.wheres, whereClause{cond: cond, args: args})
	return b
}

func (b *SelectBuilder) WhereIn(column string, values ...any) *SelectBuilder {
	if len(values) == 0 {
		b.wheres = append(b.wheres, whereClause{cond: "1 = 0"})
		return b
	}
	placeholders := strings.Repeat("?, ", len(values))
	placeholders = placeholders[:len(placeholders)-2]
	cond := fmt.Sprintf("%s IN (%s)", column, placeholders)
	b.wheres = append(b.wheres, whereClause{cond: cond, args: values})
	return b
}

func (b *SelectBuilder) Join(table, on string) *SelectBuilder {
	b.joins = append(b.joins, joinClause{kind: "JOIN", table: table, on: on})
	return b
}

func (b *SelectBuilder) LeftJoin(table, on string) *SelectBuilder {
	b.joins = append(b.joins, joinClause{kind: "LEFT JOIN", table: table, on: on})
	return b
}

func (b *SelectBuilder) GroupBy(columns ...string) *SelectBuilder {
	b.groupBys = append(b.groupBys, columns...)
	return b
}

func (b *SelectBuilder) Having(cond string, args ...any) *SelectBuilder {
	b.havings = append(b.havings, whereClause{cond: cond, args: args})
	return b
}

func (b *SelectBuilder) OrderBy(column, dir string) *SelectBuilder {
	b.orderBys = append(b.orderBys, orderByClause{column: column, dir: dir})
	return b
}

func (b *SelectBuilder) Limit(n int) *SelectBuilder {
	b.limit = n
	return b
}

func (b *SelectBuilder) Offset(n int) *SelectBuilder {
	b.offset = n
	return b
}

func (b *SelectBuilder) Build() (string, []any) {
	var buf strings.Builder
	var args []any

	buf.WriteString("SELECT ")
	if len(b.columns) == 0 {
		buf.WriteString("*")
	} else {
		for i, col := range b.columns {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(col)
			if strings.Contains(col, ".") && !strings.ContainsAny(col, " (*") {
				buf.WriteString(` AS "`)
				buf.WriteString(col)
				buf.WriteString(`"`)
			}
		}
	}

	buf.WriteString(" FROM ")
	buf.WriteString(b.table)

	for _, j := range b.joins {
		buf.WriteString(" ")
		buf.WriteString(j.kind)
		buf.WriteString(" ")
		buf.WriteString(j.table)
		buf.WriteString(" ON ")
		buf.WriteString(j.on)
	}

	if len(b.wheres) > 0 {
		buf.WriteString(" WHERE ")
		for i, w := range b.wheres {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			buf.WriteString(w.cond)
			args = append(args, w.args...)
		}
	}

	if len(b.groupBys) > 0 {
		buf.WriteString(" GROUP BY ")
		buf.WriteString(strings.Join(b.groupBys, ", "))
	}

	if len(b.havings) > 0 {
		buf.WriteString(" HAVING ")
		for i, h := range b.havings {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			buf.WriteString(h.cond)
			args = append(args, h.args...)
		}
	}

	for i, o := range b.orderBys {
		if i == 0 {
			buf.WriteString(" ORDER BY ")
		} else {
			buf.WriteString(", ")
		}
		buf.WriteString(o.column)
		buf.WriteString(" ")
		buf.WriteString(o.dir)
	}

	if b.limit >= 0 {
		fmt.Fprintf(&buf, " LIMIT %d", b.limit)
	} else if b.offset >= 0 {
		buf.WriteString(" LIMIT -1")
	}
	if b.offset >= 0 {
		fmt.Fprintf(&buf, " OFFSET %d", b.offset)
	}

	return buf.String(), args
}

// ---------------------------------------------------------------------------
// INSERT
// ---------------------------------------------------------------------------

// InsertBuilder builds INSERT queries.
type InsertBuilder struct {
	table   string
	columns []string
	values  []any
}

// Insert starts building an INSERT query for the given table.
func Insert(table string) *InsertBuilder {
	return &InsertBuilder{table: table}
}

func (b *InsertBuilder) Set(column string, value any) *InsertBuilder {
	b.columns = append(b.columns, column)
	b.values = append(b.values, value)
	return b
}

func (b *InsertBuilder) Build() (string, []any) {
	var buf strings.Builder

	buf.WriteString("INSERT INTO ")
	buf.WriteString(b.table)
	buf.WriteString(" (")
	buf.WriteString(strings.Join(b.columns, ", "))
	buf.WriteString(") VALUES (")

	for i := range b.values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("?")
	}
	buf.WriteString(")")

	return buf.String(), b.values
}

// ---------------------------------------------------------------------------
// UPDATE
// ---------------------------------------------------------------------------

// UpdateBuilder builds UPDATE queries.
type UpdateBuilder struct {
	table  string
	sets   []setClause
	wheres []whereClause
}

type setClause struct {
	column string
	value  any
}

// Update starts building an UPDATE query for the given table.
func Update(table string) *UpdateBuilder {
	return &UpdateBuilder{table: table}
}

func (b *UpdateBuilder) Set(column string, value any) *UpdateBuilder {
	b.sets = append(b.sets, setClause{column: column, value: value})
	return b
}

func (b *UpdateBuilder) Where(cond string, args ...any) *UpdateBuilder {
	b.wheres = append(b.wheres, whereClause{cond: cond, args: args})
	return b
}

func (b *UpdateBuilder) Build() (string, []any) {
	var buf strings.Builder
	var args []any

	buf.WriteString("UPDATE ")
	buf.WriteString(b.table)
	buf.WriteString(" SET ")

	for i, s := range b.sets {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(s.column)
		buf.WriteString(" = ?")
		args = append(args, s.value)
	}

	if len(b.wheres) > 0 {
		buf.WriteString(" WHERE ")
		for i, w := range b.wheres {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			buf.WriteString(w.cond)
			args = append(args, w.args...)
		}
	}

	return buf.String(), args
}

// ---------------------------------------------------------------------------
// DELETE
// ---------------------------------------------------------------------------

// DeleteBuilder builds DELETE queries.
type DeleteBuilder struct {
	table  string
	wheres []whereClause
}

// Delete starts building a DELETE query for the given table.
func Delete(table string) *DeleteBuilder {
	return &DeleteBuilder{table: table}
}

func (b *DeleteBuilder) Where(cond string, args ...any) *DeleteBuilder {
	b.wheres = append(b.wheres, whereClause{cond: cond, args: args})
	return b
}

func (b *DeleteBuilder) Build() (string, []any) {
	var buf strings.Builder
	var args []any

	buf.WriteString("DELETE FROM ")
	buf.WriteString(b.table)

	if len(b.wheres) > 0 {
		buf.WriteString(" WHERE ")
		for i, w := range b.wheres {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			buf.WriteString(w.cond)
			args = append(args, w.args...)
		}
	}

	return buf.String(), args
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// EscapeLike escapes special characters in a LIKE pattern.
// The caller must include ESCAPE '\' in the SQL clause for the
// escaping to take effect. Prefer LikeCondition for a safer API.
func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// LikeCondition returns a WHERE condition with the proper ESCAPE clause
// and the escaped pattern for use as a bind parameter.
//
// Usage:
//
//	cond, pattern := orm.LikeCondition("name", userInput)
//	builder.Where(cond, "%"+pattern+"%")
func LikeCondition(column, pattern string) (cond string, escapedPattern string) {
	return fmt.Sprintf("%s LIKE ? ESCAPE '\\'", column), EscapeLike(pattern)
}
