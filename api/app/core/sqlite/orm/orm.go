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

// NewID generates a UUIDv7 identifier (RFC 9562) in standard 36-character
// format. Delegates to utils.NewID.
func NewID() string {
	return utils.NewID()
}
