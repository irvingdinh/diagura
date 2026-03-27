package orm

import (
	"errors"
	"time"

	"localhost/app/core/utils"
)

// ErrNotFound is returned when a query expects a row but finds none.
var ErrNotFound = errors.New("orm: not found")

// BaseModel provides standard fields for domain models.
type BaseModel struct {
	ID        string    `db:"id"         json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// NewBaseModel creates a BaseModel with a generated ID and timestamps
// set to the current time.
func NewBaseModel() BaseModel {
	now := time.Now()
	return BaseModel{
		ID:        NewID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewID generates a UUIDv7 identifier (RFC 9562) in standard 36-character
// format. Delegates to utils.NewID.
func NewID() string {
	return utils.NewID()
}

// FormatTime formats a time.Time for SQLite TEXT columns.
func FormatTime(t time.Time) string {
	return utils.FormatTime(t)
}
