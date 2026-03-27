# CLAUDE.md

## Temporary Rules

Enforcing these rules when working with Go:

### HTTP imports (`net/http` vs `app/core/http`)

When both the standard library and the project core HTTP package appear in the same file, the **built-in** package is always imported with an alias; the **core** package keeps the short name `http`.

- `net/http` → import as **`nethttp`** (always alias the standard library).
- `localhost/app/core/http` (or `.../app/core/http`) → import as **`http`** (no alias on the core package).

Do not swap these. Never use `http` for `net/http` or alias the core package when both are imported.

## Verification

After implementing a feature that touches both the API and admin frontend, **always verify end-to-end before considering the work done**. Do not wait for the user to ask.

1. **API smoke test** — `make run`, then `curl` the new endpoints to confirm they return expected shapes and status codes.
2. **Browser test** — Use `playwright-cli` to walk through the full UI flow: navigate, interact with controls, confirm data renders, test edge cases (filters, pagination, empty states, expansion/collapse).
3. **Cleanup** — `make kill` and remove any test artifacts after verification.

If any step fails, fix the issue and re-verify. Only report completion once everything passes.

Always save playwright-cli snapshots and screenshots into the gitignored `.playwright-cli/` directory (use `--filename=.playwright-cli/name.png` for screenshots). Do not leave artifacts in the repo root.
