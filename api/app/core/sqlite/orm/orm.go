package orm

import (
	"crypto/rand"
	"errors"
	"sync"
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

// hextable maps each byte 0x00-0xFF to two lowercase hex characters
// packed into a uint16 (low byte = first char, high byte = second char).
var hextable [256]uint16

func init() {
	const digits = "0123456789abcdef"
	for i := range hextable {
		hextable[i] = uint16(digits[i>>4]) | uint16(digits[i&0x0f])<<8
	}
}

// uuidState holds the monotonic counter state for UUIDv7 generation
// (RFC 9562 Method 1).
var uuidState struct {
	sync.Mutex
	lastMS  int64
	counter uint16
}

// NewID generates a UUIDv7 identifier (RFC 9562) in standard 36-character
// format: xxxxxxxx-xxxx-7xxx-yxxx-xxxxxxxxxxxx.
//
// IDs are time-sortable with monotonic ordering within the same millisecond.
func NewID() string {
	return generateUUIDv7(time.Now())
}

// generateUUIDv7 produces a UUIDv7 string for the given timestamp.
func generateUUIDv7(now time.Time) string {
	ms := now.UnixMilli()

	uuidState.Lock()
	if ms <= uuidState.lastMS {
		// Clock standstill or regression: reuse last timestamp, bump counter.
		ms = uuidState.lastMS
		uuidState.counter++
		if uuidState.counter > 0x0FFF {
			// Counter overflow: advance timestamp by 1 ms to stay monotonic.
			ms++
			uuidState.lastMS = ms
			var seed [2]byte
			_, _ = rand.Read(seed[:])
			uuidState.counter = (uint16(seed[0])<<8 | uint16(seed[1])) & 0x01FF
		}
	} else {
		// New millisecond: seed counter from random with headroom.
		uuidState.lastMS = ms
		var seed [2]byte
		_, _ = rand.Read(seed[:])
		uuidState.counter = (uint16(seed[0])<<8 | uint16(seed[1])) & 0x01FF
	}
	counter := uuidState.counter
	uuidState.Unlock()

	var u [16]byte

	// Bytes 0-5: 48-bit Unix timestamp in milliseconds, big-endian.
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	// Bytes 6-7: version 7 in high nibble, counter in low 12 bits.
	u[6] = 0x70 | byte(counter>>8)&0x0F
	u[7] = byte(counter)

	// Bytes 8-15: random data.
	_, _ = rand.Read(u[8:])

	// Byte 8: set variant bits to 10xxxxxx.
	u[8] = (u[8] & 0x3F) | 0x80

	// Encode to 36-char hex string with dashes: 8-4-4-4-12.
	var buf [36]byte
	putHex := func(p int, b byte) {
		h := hextable[b]
		buf[p] = byte(h)
		buf[p+1] = byte(h >> 8)
	}

	putHex(0, u[0])
	putHex(2, u[1])
	putHex(4, u[2])
	putHex(6, u[3])
	buf[8] = '-'
	putHex(9, u[4])
	putHex(11, u[5])
	buf[13] = '-'
	putHex(14, u[6])
	putHex(16, u[7])
	buf[18] = '-'
	putHex(19, u[8])
	putHex(21, u[9])
	buf[23] = '-'
	putHex(24, u[10])
	putHex(26, u[11])
	putHex(28, u[12])
	putHex(30, u[13])
	putHex(32, u[14])
	putHex(34, u[15])

	return string(buf[:])
}

// FormatTime formats a time.Time for SQLite TEXT columns.
func FormatTime(t time.Time) string {
	return utils.FormatTime(t)
}
