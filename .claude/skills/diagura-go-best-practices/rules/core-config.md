---
title: Config Package Usage Patterns
impact: HIGH
impactDescription: Incorrect config usage causes panics at startup or silent misconfigurations at runtime
tags: config, lifecycle, defaults, validation, getters, testing
---

## Config Package Usage Patterns

**Impact: HIGH (incorrect config usage causes panics at startup or silent misconfigurations at runtime)**

The config package provides a three-layer configuration system with typed getters and validation. This rule covers the canonical way to use it in diagura.

---

### 1. Lifecycle

`config.Load()` runs **before** `fx.New`. `config.Validate()` runs in the **first** `fx.Invoke`. This ordering ensures all modules have registered their defaults and rules before validation fires.

```go
package container

import (
	"go.uber.org/fx"

	"localhost/app/core/config"
)

func Run() {
	config.Load()

	fx.New(
		// All fx.Provide constructors register defaults and rules here.

		fx.Invoke(func() {
			config.Validate() // Must be the first invoke.
		}),
	).Run()
}
```

**Incorrect (loading inside fx.Invoke):**

```go
func Run() {
	fx.New(
		fx.Invoke(func() {
			config.Load()     // Too late — providers may have already run.
			config.Validate()
		}),
	).Run()
}
```

After `config.Validate()` completes, config is **frozen**. No more `SetDefault` or `SetRule` calls are allowed.

---

### 2. Registering Defaults

Always register defaults in your module's constructor (the function passed to `fx.Provide`). Use `SetDefaults` for multiple keys.

```go
func NewHTTPServer() *HTTPServer {
	config.SetDefaults(config.Values{
		"http.port":         8080,
		"http.read_timeout": "15s",
		"http.idle_timeout": "60s",
	})

	config.SetRule("http.port", rule.Required, rule.Between(1, 65535))
	config.SetRule("http.read_timeout", rule.Required, rule.Duration)
	config.SetRule("http.idle_timeout", rule.Required, rule.Duration)

	return &HTTPServer{}
}
```

**Incorrect (using GetXOr fallbacks instead of SetDefault):**

```go
func (s *HTTPServer) Start() {
	port := config.GetIntOr("http.port", 8080) // Fallback hidden in runtime code.
	timeout := config.GetDurationOr("http.read_timeout", 15*time.Second)
}
```

`GetXOr` fallbacks are invisible to validation and introspection. Defaults belong in `SetDefault` so they participate in the resolution chain and can be overridden by config.json or environment variables.

**When to use `GetXOr`:** Only when the value is truly optional and has no meaningful default to register — e.g., a feature flag that should silently default to off.

---

### 3. Validation Rules

Import `localhost/app/core/config/rule` for all built-in rules. Register rules alongside defaults in the same constructor.

```go
import "localhost/app/core/config/rule"

config.SetRule("jwt.secret", rule.Filled, rule.MinLength(32))
config.SetRule("log.level", rule.Required, rule.In("DEBUG", "INFO", "WARN", "ERROR"))
config.SetRule("app.port", rule.Required, rule.Between(1, 65535))
config.SetRule("smtp.host", rule.Required, rule.Url)
```

Multiple `SetRule` calls for the same key **accumulate** — all rules must pass. This allows different modules to independently add constraints to the same key.

**Incorrect (validating manually in runtime code):**

```go
func (s *Server) Start() {
	port := config.GetInt("app.port")
	if port < 1 || port > 65535 {
		panic("invalid port") // Validation should happen in config.Validate(), not here.
	}
}
```

---

### 4. Getters

Use `GetX` (panics on missing/coercion failure) for **required** values that have been validated. Use `GetXOr` (returns fallback, never panics) for **optional** values.

```go
// Required values — validated by rules, safe to panic if missing.
port := config.GetInt("http.port")
secret := config.GetString("jwt.secret")
timeout := config.GetDuration("http.read_timeout")

// Optional values — may not exist, fallback is acceptable.
debug := config.GetBoolOr("app.debug", false)
workers := config.GetIntOr("worker.count", 4)
```

**Incorrect (using GetXOr for validated keys):**

```go
config.SetRule("http.port", rule.Required, rule.Between(1, 65535))

// Later...
port := config.GetIntOr("http.port", 8080) // Misleading — this key is required and validated.
```

If a key has `rule.Required`, use `GetX` — the fallback in `GetXOr` will never trigger and obscures the contract.

Available getters: `GetString`, `GetBool`, `GetInt`, `GetInt32`, `GetInt64`, `GetUint`, `GetUint8`, `GetUint16`, `GetUint32`, `GetUint64`, `GetFloat64`, `GetTime`, `GetDuration`, `GetIntSlice`, `GetStringSlice`, `GetStringMap`. Each has a corresponding `Or` variant.

---

### 5. Key Naming

Use **dot.notation** with lowercase segments. Keys map to environment variables automatically: dots become underscores, everything uppercased.

| Config key | Environment variable |
|------------|---------------------|
| `http.port` | `HTTP_PORT` |
| `db.host` | `DB_HOST` |
| `jwt.secret` | `JWT_SECRET` |
| `email.from_address` | `EMAIL_FROM_ADDRESS` |

Group related keys under a common prefix:

```go
config.SetDefaults(config.Values{
	"db.host":     "localhost",
	"db.port":     5432,
	"db.name":     "diagura",
	"db.user":     "postgres",
	"db.password": "",
})
```

**Incorrect (flat keys without grouping):**

```go
config.SetDefault("database_host", "localhost")
config.SetDefault("databasePort", 5432) // Inconsistent casing, no grouping.
```

---

### 6. Three-Layer Resolution

Values are resolved in priority order: **environment variable > config.json > registered default**. First match wins.

```
ENV: HTTP_PORT=3000          ← highest priority (always wins)
config.json: {"http": {"port": 8080}}  ← medium priority
SetDefault("http.port", 19110)         ← lowest priority (fallback)

config.GetInt("http.port") → 3000
```

The config.json file lives at `{DATA_DIR}/config.json`. `DATA_DIR` defaults to `~/.standalone`. Nested JSON is automatically flattened:

```json
{
  "http": {
    "port": 8080,
    "read_timeout": "15s"
  }
}
```

Becomes flat keys: `http.port`, `http.read_timeout`.

Use `config.Has("key")` to check existence across all layers without coercion.

---

### 7. Testing

Use `config.Reset()` to clear all state in tests. Use `t.Setenv` to set environment variables that will be cleaned up automatically.

```go
func TestMyFeature(t *testing.T) {
	config.Reset()

	config.SetDefault("feature.enabled", true)
	config.SetDefault("feature.limit", 100)

	t.Setenv("FEATURE_LIMIT", "50") // Overrides default via env var.

	got := config.GetInt("feature.limit")
	if got != 50 {
		t.Errorf("expected 50, got %d", got)
	}
}
```

**Incorrect (calling Load in unit tests):**

```go
func TestMyFeature(t *testing.T) {
	config.Load() // Reads from disk, creates ~/.standalone — side effects in tests.
}
```

`config.Reset()` gives you a blank slate without filesystem side effects. Only use `config.Load()` in integration tests that need the full config.json flow, and always pair it with `t.Setenv("DATA_DIR", t.TempDir())`.

```go
func TestWithConfigFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"http":{"port":9999}}`), 0o644)
	t.Setenv("DATA_DIR", dir)

	config.Load()

	got := config.GetInt("http.port")
	if got != 9999 {
		t.Errorf("expected 9999, got %d", got)
	}
}
```
