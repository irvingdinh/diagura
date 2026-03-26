package config

import (
	"fmt"
	"sort"
	"strings"
)

// SetRule registers one or more validation rules for a config key. Rules are
// checked when Validate is called. Multiple SetRule calls for the same key
// accumulate — all rules must pass.
//
// Call SetRule during the Register phase (alongside SetDefault). Rules
// registered after Validate has run will not be checked.
//
//	config.SetRule("http.port", rule.Required, rule.Between(1, 65535))
//	config.SetRule("log.level", rule.Required, rule.In("DEBUG", "INFO", "WARN", "ERROR"))
func SetRule(key string, rules ...Rule) {
	global.mu.Lock()
	if global.frozen {
		global.mu.Unlock()
		panic(fmt.Sprintf("config: SetRule(%q) called after config is frozen", key))
	}
	global.rules[key] = append(global.rules[key], rules...)
	global.mu.Unlock()
}

// Validate runs all registered rules and panics with a summary of every
// violation. The panic message lists each failing key with its environment
// variable name and the specific error.
//
// Called by the app framework after all modules have registered defaults
// and rules — but before any module's Boot phase. Config is frozen after
// Validate completes.
func Validate() {
	// Snapshot rules under read lock, then release before calling resolve
	// (which acquires its own read lock). This avoids potential deadlock
	// if a writer is waiting between the two acquisitions.
	global.mu.RLock()
	snapshot := make(map[string][]Rule, len(global.rules))
	for k, v := range global.rules {
		snapshot[k] = v
	}
	global.mu.RUnlock()

	var errs []string
	for key, rules := range snapshot {
		val, exists := resolve(key)
		for _, rule := range rules {
			if err := rule(key, val, exists); err != nil {
				errs = append(errs, fmt.Sprintf("  %s (%s): %s", key, envName(key), err))
			}
		}
	}

	if len(errs) > 0 {
		sort.Strings(errs)
		panic(fmt.Sprintf("config: validation failed:\n%s", strings.Join(errs, "\n")))
	}

	Freeze()
}

// Freeze prevents further SetDefault, SetDefaults, and SetRule calls. Called
// automatically at the end of Validate. Can also be called manually to freeze
// config earlier.
//
// Freeze is idempotent — calling it multiple times is safe.
func Freeze() {
	global.mu.Lock()
	global.frozen = true
	global.mu.Unlock()
}

// IsFrozen reports whether the config has been frozen.
func IsFrozen() bool {
	global.mu.RLock()
	frozen := global.frozen
	global.mu.RUnlock()
	return frozen
}
