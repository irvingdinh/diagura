package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Panic helpers
// ---------------------------------------------------------------------------

func notFoundErr(key string) string {
	return fmt.Sprintf("config: key %q not found (set %s env var or add to config.json)", key, envName(key))
}

func coerceErr(key, typeName string, raw any) string {
	return fmt.Sprintf("config: cannot coerce %q to %s (raw value: %v)", key, typeName, raw)
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func GetString(key string) string {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	return toString(raw)
}

func GetStringOr(key string, fallback string) string {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	return toString(raw)
}

// ---------------------------------------------------------------------------
// Bool
// ---------------------------------------------------------------------------

func GetBool(key string) bool {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toBool(raw)
	if !ok {
		panic(coerceErr(key, "bool", raw))
	}
	return v
}

func GetBoolOr(key string, fallback bool) bool {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toBool(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Int
// ---------------------------------------------------------------------------

func GetInt(key string) int {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toInt(raw)
	if !ok {
		panic(coerceErr(key, "int", raw))
	}
	return v
}

func GetIntOr(key string, fallback int) int {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toInt(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Int32
// ---------------------------------------------------------------------------

func GetInt32(key string) int32 {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toInt32(raw)
	if !ok {
		panic(coerceErr(key, "int32", raw))
	}
	return v
}

func GetInt32Or(key string, fallback int32) int32 {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toInt32(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Int64
// ---------------------------------------------------------------------------

func GetInt64(key string) int64 {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toInt64(raw)
	if !ok {
		panic(coerceErr(key, "int64", raw))
	}
	return v
}

func GetInt64Or(key string, fallback int64) int64 {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toInt64(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Uint
// ---------------------------------------------------------------------------

func GetUint(key string) uint {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toUint(raw)
	if !ok {
		panic(coerceErr(key, "uint", raw))
	}
	return v
}

func GetUintOr(key string, fallback uint) uint {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toUint(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Uint8
// ---------------------------------------------------------------------------

func GetUint8(key string) uint8 {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toUint8(raw)
	if !ok {
		panic(coerceErr(key, "uint8", raw))
	}
	return v
}

func GetUint8Or(key string, fallback uint8) uint8 {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toUint8(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Uint16
// ---------------------------------------------------------------------------

func GetUint16(key string) uint16 {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toUint16(raw)
	if !ok {
		panic(coerceErr(key, "uint16", raw))
	}
	return v
}

func GetUint16Or(key string, fallback uint16) uint16 {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toUint16(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Uint32
// ---------------------------------------------------------------------------

func GetUint32(key string) uint32 {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toUint32(raw)
	if !ok {
		panic(coerceErr(key, "uint32", raw))
	}
	return v
}

func GetUint32Or(key string, fallback uint32) uint32 {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toUint32(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Uint64
// ---------------------------------------------------------------------------

func GetUint64(key string) uint64 {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toUint64(raw)
	if !ok {
		panic(coerceErr(key, "uint64", raw))
	}
	return v
}

func GetUint64Or(key string, fallback uint64) uint64 {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toUint64(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Float64
// ---------------------------------------------------------------------------

func GetFloat64(key string) float64 {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toFloat64(raw)
	if !ok {
		panic(coerceErr(key, "float64", raw))
	}
	return v
}

func GetFloat64Or(key string, fallback float64) float64 {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toFloat64(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Time
// ---------------------------------------------------------------------------

func GetTime(key string) time.Time {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toTime(raw)
	if !ok {
		panic(coerceErr(key, "time.Time", raw))
	}
	return v
}

func GetTimeOr(key string, fallback time.Time) time.Time {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toTime(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Duration
// ---------------------------------------------------------------------------

func GetDuration(key string) time.Duration {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toDuration(raw)
	if !ok {
		panic(coerceErr(key, "time.Duration", raw))
	}
	return v
}

func GetDurationOr(key string, fallback time.Duration) time.Duration {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toDuration(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// IntSlice
// ---------------------------------------------------------------------------

func GetIntSlice(key string) []int {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toIntSlice(raw)
	if !ok {
		panic(coerceErr(key, "[]int", raw))
	}
	return v
}

func GetIntSliceOr(key string, fallback []int) []int {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toIntSlice(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// StringSlice
// ---------------------------------------------------------------------------

func GetStringSlice(key string) []string {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toStringSlice(raw)
	if !ok {
		panic(coerceErr(key, "[]string", raw))
	}
	return v
}

func GetStringSliceOr(key string, fallback []string) []string {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toStringSlice(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// StringMap
// ---------------------------------------------------------------------------

func GetStringMap(key string) map[string]any {
	raw, ok := resolve(key)
	if !ok {
		panic(notFoundErr(key))
	}
	v, ok := toStringMap(raw)
	if !ok {
		panic(coerceErr(key, "map[string]any", raw))
	}
	return v
}

func GetStringMapOr(key string, fallback map[string]any) map[string]any {
	raw, ok := resolve(key)
	if !ok {
		return fallback
	}
	v, ok := toStringMap(raw)
	if !ok {
		return fallback
	}
	return v
}

// ---------------------------------------------------------------------------
// Scalar coercion
// ---------------------------------------------------------------------------

func toString(raw any) string {
	switch v := raw.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toBool(raw any) (bool, bool) {
	switch v := raw.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(v) {
		case "true", "1", "t", "yes", "on":
			return true, true
		case "false", "0", "f", "no", "off":
			return false, true
		default:
			return false, false
		}
	case int:
		return v != 0, true
	case float64:
		return v != 0, true
	default:
		return false, false
	}
}

func toInt(raw any) (int, bool) {
	switch v := raw.(type) {
	case int:
		return v, true
	case int64:
		if v > math.MaxInt || v < math.MinInt {
			return 0, false
		}
		return int(v), true
	case float64:
		if v > float64(math.MaxInt) || v < float64(math.MinInt) {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return int(v), true
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		if n > int64(math.MaxInt) || n < int64(math.MinInt) {
			return 0, false
		}
		return int(n), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toInt32(raw any) (int32, bool) {
	switch v := raw.(type) {
	case int32:
		return v, true
	case int:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return 0, false
		}
		return int32(v), true
	case int64:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return 0, false
		}
		return int32(v), true
	case float64:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return int32(v), true
	case string:
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return 0, false
		}
		return int32(n), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toInt64(raw any) (int64, bool) {
	switch v := raw.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case float64:
		if v > float64(math.MaxInt64) || v < float64(math.MinInt64) {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return int64(v), true
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toUint(raw any) (uint, bool) {
	switch v := raw.(type) {
	case uint:
		return v, true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case float64:
		if v < 0 || v > float64(math.MaxUint) {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return uint(v), true
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(n), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toUint8(raw any) (uint8, bool) {
	switch v := raw.(type) {
	case uint8:
		return v, true
	case int:
		if v < 0 || v > math.MaxUint8 {
			return 0, false
		}
		return uint8(v), true
	case int64:
		if v < 0 || v > math.MaxUint8 {
			return 0, false
		}
		return uint8(v), true
	case float64:
		if v < 0 || v > math.MaxUint8 {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return uint8(v), true
	case string:
		n, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return 0, false
		}
		return uint8(n), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toUint16(raw any) (uint16, bool) {
	switch v := raw.(type) {
	case uint16:
		return v, true
	case int:
		if v < 0 || v > math.MaxUint16 {
			return 0, false
		}
		return uint16(v), true
	case int64:
		if v < 0 || v > math.MaxUint16 {
			return 0, false
		}
		return uint16(v), true
	case float64:
		if v < 0 || v > math.MaxUint16 {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return uint16(v), true
	case string:
		n, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return 0, false
		}
		return uint16(n), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toUint32(raw any) (uint32, bool) {
	switch v := raw.(type) {
	case uint32:
		return v, true
	case int:
		if v < 0 {
			return 0, false
		}
		if uint64(v) > math.MaxUint32 {
			return 0, false
		}
		return uint32(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		if uint64(v) > math.MaxUint32 {
			return 0, false
		}
		return uint32(v), true
	case float64:
		if v < 0 || v > math.MaxUint32 {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return uint32(v), true
	case string:
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, false
		}
		return uint32(n), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toUint64(raw any) (uint64, bool) {
	switch v := raw.(type) {
	case uint64:
		return v, true
	case uint:
		return uint64(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		if math.Trunc(v) != v {
			return 0, false
		}
		return uint64(v), true
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toFloat64(raw any) (float64, bool) {
	switch v := raw.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toTime(raw any) (time.Time, bool) {
	switch v := raw.(type) {
	case time.Time:
		return v, true
	case string:
		for _, layout := range []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02",
			"2006-01-02 15:04:05",
		} {
			t, err := time.Parse(layout, v)
			if err == nil {
				return t, true
			}
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

func toDuration(raw any) (time.Duration, bool) {
	switch v := raw.(type) {
	case time.Duration:
		return v, true
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			return 0, false
		}
		return d, true
	default:
		return 0, false
	}
}

// ---------------------------------------------------------------------------
// Collection coercion
// ---------------------------------------------------------------------------

func toIntSlice(raw any) ([]int, bool) {
	switch v := raw.(type) {
	case []int:
		return v, true
	case []any:
		result := make([]int, 0, len(v))
		for _, item := range v {
			n, ok := toInt(item)
			if !ok {
				return nil, false
			}
			result = append(result, n)
		}
		return result, true
	case string:
		if v == "" {
			return nil, false
		}
		parts := strings.Split(v, ",")
		result := make([]int, 0, len(parts))
		for _, p := range parts {
			n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
			if err != nil {
				return nil, false
			}
			result = append(result, int(n))
		}
		return result, true
	default:
		return nil, false
	}
}

func toStringSlice(raw any) ([]string, bool) {
	switch v := raw.(type) {
	case []string:
		return v, true
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result, true
	case string:
		if v == "" {
			return nil, false
		}
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			result = append(result, strings.TrimSpace(p))
		}
		return result, true
	default:
		return nil, false
	}
}

func toStringMap(raw any) (map[string]any, bool) {
	if v, ok := raw.(map[string]any); ok {
		return v, true
	}
	return nil, false
}
