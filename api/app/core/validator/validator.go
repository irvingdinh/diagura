package validator

import (
	"fmt"
	"math"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Rule validates a single value. Return a non-nil error to indicate failure.
type Rule func(value any) error

// Validate runs all rules against value and returns all errors encountered.
// Returns nil if all rules pass.
func Validate(value any, rules ...Rule) []error {
	var errs []error
	for _, rule := range rules {
		if err := rule(value); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// ---------------------------------------------------------------------------
// Presence rules
// ---------------------------------------------------------------------------

// Required ensures the value is not nil.
var Required Rule = func(value any) error {
	if value == nil {
		return fmt.Errorf("required but not set")
	}
	return nil
}

// Filled ensures the value is not nil and its string representation is not
// empty. Useful for secrets and credentials that must have a non-trivial value.
var Filled Rule = func(value any) error {
	if value == nil {
		return fmt.Errorf("required but not set")
	}
	if asString(value) == "" {
		return fmt.Errorf("must not be empty")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Numeric rules
// ---------------------------------------------------------------------------

// Positive ensures a numeric value is strictly greater than zero.
var Positive Rule = func(value any) error {
	n, ok := asInt(value)
	if !ok {
		return fmt.Errorf("must be a positive integer")
	}
	if n <= 0 {
		return fmt.Errorf("must be positive, got %d", n)
	}
	return nil
}

// NonNegative ensures a numeric value is zero or greater.
var NonNegative Rule = func(value any) error {
	n, ok := asInt(value)
	if !ok {
		return fmt.Errorf("must be a non-negative integer")
	}
	if n < 0 {
		return fmt.Errorf("must be non-negative, got %d", n)
	}
	return nil
}

// Min ensures a numeric value is at least min (inclusive).
func Min(min int) Rule {
	return func(value any) error {
		n, ok := asInt(value)
		if !ok {
			return fmt.Errorf("must be an integer")
		}
		if n < min {
			return fmt.Errorf("must be >= %d, got %d", min, n)
		}
		return nil
	}
}

// Max ensures a numeric value is at most max (inclusive).
func Max(max int) Rule {
	return func(value any) error {
		n, ok := asInt(value)
		if !ok {
			return fmt.Errorf("must be an integer")
		}
		if n > max {
			return fmt.Errorf("must be <= %d, got %d", max, n)
		}
		return nil
	}
}

// Between ensures a numeric value is within [min, max] inclusive.
//
//	validator.Between(1, 65535)
func Between(min, max int) Rule {
	return func(value any) error {
		n, ok := asInt(value)
		if !ok {
			return fmt.Errorf("must be an integer")
		}
		if n < min || n > max {
			return fmt.Errorf("must be in [%d, %d], got %d", min, max, n)
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// String length rules
// ---------------------------------------------------------------------------

// MinLength ensures a string value has at least n characters.
//
//	validator.MinLength(32)
func MinLength(n int) Rule {
	return func(value any) error {
		s := asString(value)
		if len(s) < n {
			return fmt.Errorf("must be at least %d characters, got %d", n, len(s))
		}
		return nil
	}
}

// MaxLength ensures a string value has at most n characters.
func MaxLength(n int) Rule {
	return func(value any) error {
		s := asString(value)
		if len(s) > n {
			return fmt.Errorf("must be at most %d characters, got %d", n, len(s))
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Membership rules
// ---------------------------------------------------------------------------

// In ensures the string value is one of the allowed values. Case-sensitive.
//
//	validator.In("DEBUG", "INFO", "WARN", "ERROR")
func In(allowed ...string) Rule {
	return func(value any) error {
		s := asString(value)
		for _, a := range allowed {
			if s == a {
				return nil
			}
		}
		return fmt.Errorf("must be one of [%s], got %q", strings.Join(allowed, ", "), s)
	}
}

// InFold ensures the string value is one of the allowed values. Case-insensitive.
//
//	validator.InFold("debug", "info", "warn", "error")
func InFold(allowed ...string) Rule {
	return func(value any) error {
		s := asString(value)
		for _, a := range allowed {
			if strings.EqualFold(s, a) {
				return nil
			}
		}
		return fmt.Errorf("must be one of [%s], got %q", strings.Join(allowed, ", "), s)
	}
}

// NotIn ensures the string value is none of the disallowed values. Case-sensitive.
func NotIn(disallowed ...string) Rule {
	return func(value any) error {
		s := asString(value)
		for _, d := range disallowed {
			if s == d {
				return fmt.Errorf("must not be one of [%s], got %q", strings.Join(disallowed, ", "), s)
			}
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Type rules
// ---------------------------------------------------------------------------

// Integer ensures the value is parseable as an integer.
var Integer Rule = func(value any) error {
	if _, ok := asInt(value); !ok {
		return fmt.Errorf("must be an integer")
	}
	return nil
}

// Numeric ensures the value is parseable as a number.
var Numeric Rule = func(value any) error {
	if _, ok := asFloat64(value); !ok {
		return fmt.Errorf("must be numeric")
	}
	return nil
}

// Boolean ensures the value is parseable as a boolean.
var Boolean Rule = func(value any) error {
	if _, ok := asBool(value); !ok {
		return fmt.Errorf("must be a boolean")
	}
	return nil
}

// Duration ensures the value is a valid Go duration string (e.g. "15s",
// "5m", "300ms").
var Duration Rule = func(value any) error {
	s := asString(value)
	if _, err := time.ParseDuration(s); err != nil {
		return fmt.Errorf("must be a valid duration (e.g. \"15s\", \"5m\", \"300ms\"), got %q", s)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Format rules
// ---------------------------------------------------------------------------

// Url ensures the value is a valid URL with a scheme.
var Url Rule = func(value any) error {
	s := asString(value)
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("must be a valid URL, got %q", s)
	}
	return nil
}

// Ip ensures the value is a valid IPv4 or IPv6 address.
var Ip Rule = func(value any) error {
	s := asString(value)
	if net.ParseIP(s) == nil {
		return fmt.Errorf("must be a valid IP address, got %q", s)
	}
	return nil
}

// Email ensures the value is a valid email address.
var Email Rule = func(value any) error {
	s := asString(value)
	if _, err := mail.ParseAddress(s); err != nil {
		return fmt.Errorf("must be a valid email address, got %q", s)
	}
	return nil
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Uuid ensures the value matches UUID format.
var Uuid Rule = func(value any) error {
	s := asString(value)
	if !uuidPattern.MatchString(s) {
		return fmt.Errorf("must be a valid UUID, got %q", s)
	}
	return nil
}

// Regex ensures the string value matches the given regular expression pattern.
//
//	validator.Regex(`^[a-z]+$`)
func Regex(pattern string) Rule {
	re := regexp.MustCompile(pattern)
	return func(value any) error {
		s := asString(value)
		if !re.MatchString(s) {
			return fmt.Errorf("must match pattern %q, got %q", pattern, s)
		}
		return nil
	}
}

// StartsWith ensures the string value starts with the given prefix.
func StartsWith(prefix string) Rule {
	return func(value any) error {
		s := asString(value)
		if !strings.HasPrefix(s, prefix) {
			return fmt.Errorf("must start with %q, got %q", prefix, s)
		}
		return nil
	}
}

// EndsWith ensures the string value ends with the given suffix.
func EndsWith(suffix string) Rule {
	return func(value any) error {
		s := asString(value)
		if !strings.HasSuffix(s, suffix) {
			return fmt.Errorf("must end with %q, got %q", suffix, s)
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Case rules
// ---------------------------------------------------------------------------

// Lowercase ensures the string value is entirely lowercase.
var Lowercase Rule = func(value any) error {
	s := asString(value)
	if s != strings.ToLower(s) {
		return fmt.Errorf("must be lowercase, got %q", s)
	}
	return nil
}

// Uppercase ensures the string value is entirely uppercase.
var Uppercase Rule = func(value any) error {
	s := asString(value)
	if s != strings.ToUpper(s) {
		return fmt.Errorf("must be uppercase, got %q", s)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internal coercion helpers
// ---------------------------------------------------------------------------

func asString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

func asInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		if val > math.MaxInt || val < math.MinInt {
			return 0, false
		}
		return int(val), true
	case float64:
		if val > float64(math.MaxInt) || val < float64(math.MinInt) {
			return 0, false
		}
		if math.Trunc(val) != val {
			return 0, false
		}
		return int(val), true
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, false
		}
		if n > int64(math.MaxInt) || n < int64(math.MinInt) {
			return 0, false
		}
		return int(n), true
	case bool:
		if val {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func asFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case bool:
		if val {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func asBool(v any) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case string:
		switch strings.ToLower(val) {
		case "true", "1", "t", "yes", "on":
			return true, true
		case "false", "0", "f", "no", "off":
			return false, true
		default:
			return false, false
		}
	case int:
		return val != 0, true
	case float64:
		return val != 0, true
	default:
		return false, false
	}
}
