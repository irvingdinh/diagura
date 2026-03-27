package rule

import (
	"fmt"

	"localhost/app/core/config"
	"localhost/app/core/validator"
)

// wrap converts a validator.Rule into a config.Rule that skips absent keys.
func wrap(vr validator.Rule) config.Rule {
	return func(_ string, value any, exists bool) error {
		if !exists {
			return nil
		}
		return vr(value)
	}
}

// ---------------------------------------------------------------------------
// Presence rules
// ---------------------------------------------------------------------------

// Required ensures the key exists in at least one config layer. Zero values
// (empty string, 0, false) are considered valid — use Filled for stricter checks.
var Required config.Rule = func(_ string, _ any, exists bool) error {
	if !exists {
		return fmt.Errorf("required but not set")
	}
	return nil
}

// Filled ensures the key exists and its string representation is not empty.
var Filled config.Rule = func(_ string, value any, exists bool) error {
	if !exists {
		return fmt.Errorf("required but not set")
	}
	return validator.Filled(value)
}

// ---------------------------------------------------------------------------
// Numeric rules
// ---------------------------------------------------------------------------

var Positive = wrap(validator.Positive)
var NonNegative = wrap(validator.NonNegative)
var Integer = wrap(validator.Integer)
var Numeric = wrap(validator.Numeric)

func Min(n int) config.Rule            { return wrap(validator.Min(n)) }
func Max(n int) config.Rule            { return wrap(validator.Max(n)) }
func Between(min, max int) config.Rule { return wrap(validator.Between(min, max)) }

// ---------------------------------------------------------------------------
// String rules
// ---------------------------------------------------------------------------

var Boolean = wrap(validator.Boolean)
var Lowercase = wrap(validator.Lowercase)
var Uppercase = wrap(validator.Uppercase)

func MinLength(n int) config.Rule            { return wrap(validator.MinLength(n)) }
func MaxLength(n int) config.Rule            { return wrap(validator.MaxLength(n)) }
func In(allowed ...string) config.Rule       { return wrap(validator.In(allowed...)) }
func InFold(allowed ...string) config.Rule   { return wrap(validator.InFold(allowed...)) }
func NotIn(disallowed ...string) config.Rule { return wrap(validator.NotIn(disallowed...)) }
func Regex(pattern string) config.Rule       { return wrap(validator.Regex(pattern)) }
func StartsWith(prefix string) config.Rule   { return wrap(validator.StartsWith(prefix)) }
func EndsWith(suffix string) config.Rule     { return wrap(validator.EndsWith(suffix)) }

// ---------------------------------------------------------------------------
// Format rules
// ---------------------------------------------------------------------------

var Duration = wrap(validator.Duration)
var Url = wrap(validator.Url)
var Ip = wrap(validator.Ip)
var Email = wrap(validator.Email)
var Uuid = wrap(validator.Uuid)
