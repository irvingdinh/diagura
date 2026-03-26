# CLAUDE.md

## Temporary Rules

Enforcing these rules when working with Go:

### HTTP imports (`net/http` vs `app/core/http`)

When both the standard library and the project core HTTP package appear in the same file, the **built-in** package is always imported with an alias; the **core** package keeps the short name `http`.

- `net/http` → import as **`nethttp`** (always alias the standard library).
- `localhost/app/core/http` (or `.../app/core/http`) → import as **`http`** (no alias on the core package).

Do not swap these. Never use `http` for `net/http` or alias the core package when both are imported.
