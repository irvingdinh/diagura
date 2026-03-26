package driver

import "errors"

// SQLite primary result codes.
const (
	CodeConstraint = 19 // SQLITE_CONSTRAINT
)

// SQLite extended result codes for constraint violations.
const (
	CodeConstraintCheck      = 275  // SQLITE_CONSTRAINT_CHECK
	CodeConstraintForeignKey = 787  // SQLITE_CONSTRAINT_FOREIGNKEY
	CodeConstraintNotNull    = 1299 // SQLITE_CONSTRAINT_NOTNULL
	CodeConstraintPrimaryKey = 1555 // SQLITE_CONSTRAINT_PRIMARYKEY
	CodeConstraintUnique     = 2067 // SQLITE_CONSTRAINT_UNIQUE
)

// Error represents a SQLite error with result code information.
type Error struct {
	Code         int
	ExtendedCode int
	Message      string
}

func (e *Error) Error() string {
	return e.Message
}

// IsConstraintError reports whether err is any SQLite constraint violation.
func IsConstraintError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == CodeConstraint
}

// IsUniqueConstraintError reports whether err is a UNIQUE constraint violation.
func IsUniqueConstraintError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.ExtendedCode == CodeConstraintUnique
}

// IsForeignKeyConstraintError reports whether err is a FOREIGN KEY constraint violation.
func IsForeignKeyConstraintError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.ExtendedCode == CodeConstraintForeignKey
}

// IsNotNullConstraintError reports whether err is a NOT NULL constraint violation.
func IsNotNullConstraintError(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.ExtendedCode == CodeConstraintNotNull
}
