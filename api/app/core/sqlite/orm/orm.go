package orm

import (
	"errors"

	"localhost/app/core/utils"
)

// ErrNotFound is returned when a query expects a row but finds none.
var ErrNotFound = errors.New("orm: not found")

// NewID generates a UUIDv7 identifier (RFC 9562) in standard 36-character
// format. Delegates to utils.NewID.
func NewID() string {
	return utils.NewID()
}
