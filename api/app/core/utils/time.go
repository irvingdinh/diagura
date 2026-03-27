package utils

import (
	"fmt"
	"time"
)

// Time format constants for SQLite TEXT timestamps.
const (
	TimeFormat   = "2006-01-02 15:04:05"
	TimeFormatMs = "2006-01-02 15:04:05.000"
)

// FormatTime formats a time.Time for SQLite TEXT columns (UTC, millisecond
// precision).
func FormatTime(t time.Time) string {
	return t.UTC().Format(TimeFormatMs)
}

// ParseTime parses a SQLite TEXT timestamp, trying millisecond format,
// second format, and RFC 3339 in order.
func ParseTime(s string) (time.Time, error) {
	if t, err := time.Parse(TimeFormatMs, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(TimeFormat, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("utils: parse time %q: unrecognized format", s)
}
