package orm

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sync/atomic"
	"time"
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

// NewID generates a unique, time-sortable 20-character alphanumeric
// identifier. Format: 8 chars timestamp (base32, 40-bit) + 12 chars random.
// The timestamp component wraps every ~34.8 years; see generateID.
func NewID() string {
	return generateID(time.Now())
}

const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

var idCounter atomic.Uint64

// generateID produces a 20-character ID: 8 base32 chars for the timestamp
// (lower 40 bits of millisecond precision) + 12 base32 chars of randomness.
// The 40-bit timestamp wraps every ~34.8 years. Since current Unix
// milliseconds exceed 40 bits, the upper bits are truncated. IDs remain
// time-sortable within any ~34.8-year window (next wrap ~2039).
func generateID(now time.Time) string {
	var buf [20]byte

	ms := uint64(now.UnixMilli())
	for i := 7; i >= 0; i-- {
		buf[i] = alphabet[ms&0x1f]
		ms >>= 5
	}

	var randomBytes [8]byte
	_, _ = rand.Read(randomBytes[:])

	counter := idCounter.Add(1)
	binary.LittleEndian.PutUint64(randomBytes[:], binary.LittleEndian.Uint64(randomBytes[:])+counter)

	for i := 0; i < 12; i++ {
		byteIdx := (i * 5) / 8
		bitOffset := uint((i * 5) % 8)

		var val uint16
		if byteIdx+1 < len(randomBytes) {
			val = uint16(randomBytes[byteIdx]) | uint16(randomBytes[byteIdx+1])<<8
		} else {
			val = uint16(randomBytes[byteIdx])
		}
		buf[8+i] = alphabet[(val>>bitOffset)&0x1f]
	}

	return string(buf[:])
}

// Time format constants for parsing SQLite TEXT timestamps.
const (
	timeFormat   = "2006-01-02 15:04:05"
	timeFormatMs = "2006-01-02 15:04:05.000"
)

// FormatTime formats a time.Time for SQLite TEXT columns.
func FormatTime(t time.Time) string {
	return t.UTC().Format(timeFormatMs)
}
