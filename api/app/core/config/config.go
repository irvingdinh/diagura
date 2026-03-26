package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Rule validates a resolved config value. It receives the key name, the
// resolved value (or nil if missing), and whether the key exists in any
// layer. Return a non-nil error to indicate a validation failure.
type Rule func(key string, value any, exists bool) error

// Values is a convenience alias for flat config maps, mainly for SetDefaults
// calls and examples.
type Values = map[string]any

type state struct {
	mu       sync.RWMutex
	frozen   bool
	values   map[string]any    // from JSON file (flattened)
	defaults map[string]any    // from SetDefault calls
	rules    map[string][]Rule // from SetRule calls
}

var global = state{
	values:   make(map[string]any),
	defaults: make(map[string]any),
	rules:    make(map[string][]Rule),
}

// Load resolves the data directory, creates it if needed, loads
// {data_dir}/config.json, and initializes the global config state. It panics
// if the config file exists but contains malformed JSON. A missing config file
// is silently ignored.
//
// The data directory is resolved from the DATA_DIR environment variable,
// falling back to ~/.standalone. The resolved value is stored as a default
// so callers can retrieve it via GetString("data_dir").
//
// Calling Load again resets all state (values, defaults, and rules). This is
// useful in tests to get a clean config between test cases.
func Load() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("config: resolving home directory: %v", err))
		}
		dataDir = filepath.Join(home, ".standalone")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(fmt.Sprintf("config: creating data directory %q: %v", dataDir, err))
	}

	path := filepath.Join(dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		panic(fmt.Sprintf("config: reading config file: %v", err))
	}

	var flat map[string]any
	if data != nil {
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			panic(fmt.Sprintf("config: parsing config file: %v", err))
		}
		flat = make(map[string]any)
		flatten("", raw, flat)
	}

	defaults := make(map[string]any)
	defaults["data_dir"] = dataDir

	global.mu.Lock()
	global.frozen = false
	global.values = flat
	global.defaults = defaults
	global.rules = make(map[string][]Rule)
	global.mu.Unlock()
}

// Reset clears all config state without reading any files. Useful as a test
// helper to get a blank slate without needing a temp directory.
func Reset() {
	global.mu.Lock()
	global.frozen = false
	global.values = nil
	global.defaults = make(map[string]any)
	global.rules = make(map[string][]Rule)
	global.mu.Unlock()
}

// SetDefault registers a default value for a key. Called by modules during
// their registration phase. Defaults have the lowest priority — config file
// values and environment variables take precedence.
//
// When multiple modules call SetDefault for the same key, the last call wins.
// Panics if the config has been frozen (after Validate).
func SetDefault(key string, value any) {
	global.mu.Lock()
	if global.frozen {
		global.mu.Unlock()
		panic(fmt.Sprintf("config: SetDefault(%q) called after config is frozen", key))
	}
	global.defaults[key] = value
	global.mu.Unlock()
}

// SetDefaults registers multiple default values at once from a flat map.
// Equivalent to calling SetDefault for each entry. Useful for modules that
// need to register many defaults during their registration phase.
//
//	config.SetDefaults(Values{
//	    "http.port":         19110,
//	    "http.read_timeout": "15s",
//	    "http.idle_timeout": "60s",
//	})
func SetDefaults(m Values) {
	global.mu.Lock()
	if global.frozen {
		global.mu.Unlock()
		panic("config: SetDefaults called after config is frozen")
	}
	for k, v := range m {
		global.defaults[k] = v
	}
	global.mu.Unlock()
}

// Has reports whether a key can be resolved from any layer (environment
// variable, config file, or registered default). It does not attempt type
// coercion — it only checks for existence.
func Has(key string) bool {
	_, ok := resolve(key)
	return ok
}

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

// resolve returns the value for a key, checking all three layers.
// Resolution order: env var → config file → registered defaults.
func resolve(key string) (any, bool) {
	envKey := envName(key)
	if envVal, ok := os.LookupEnv(envKey); ok {
		return envVal, true
	}

	global.mu.RLock()
	defer global.mu.RUnlock()

	if val, ok := global.values[key]; ok {
		return val, true
	}

	if val, ok := global.defaults[key]; ok {
		return val, true
	}

	return nil, false
}

// envName converts a dot-notation key to an environment variable name.
// "email.from_address" → "EMAIL_FROM_ADDRESS"
func envName(key string) string {
	return strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

// flatten recursively walks a nested map and produces a flat map with
// dot-separated keys.
func flatten(prefix string, m map[string]any, out map[string]any) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			flatten(key, val, out)
		default:
			out[key] = val
		}
	}
}
