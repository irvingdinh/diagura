---
title: Module Service Layer Convention
impact: MEDIUM
impactDescription: Inconsistent module structure makes it harder for AI agents to understand patterns and build new modules
tags: module, service, handler, dependency-injection, architecture
---

## Module Service Layer Convention

**Impact: MEDIUM (inconsistent module structure makes it harder for AI agents to understand patterns and build new modules)**

Every module that contains business logic should follow the service layer pattern. The handler is responsible for HTTP concerns only; the service encapsulates business logic and data access. This keeps handlers thin and makes business logic reusable across handlers, middleware, and external modules.

---

### 1. When to Create a Service

Create a `service/` sub-package when a module has:
- Business logic beyond simple CRUD (password hashing, validation rules, multi-step operations)
- Data access that needs to be shared across handler and middleware
- Types or functions consumed by external modules

Do **not** create a service for stub modules or modules whose handler is a trivial pass-through with no business logic.

---

### 2. Module Structure

```
app/{module}/
├── {module}.go           # Module struct + RegisterRoutes() + Provide()
├── service/
│   └── service.go        # Business logic + data access
├── handler/
│   └── handler.go        # HTTP parsing + response writing
├── [middleware/]          # Optional: auth/validation middleware
└── [entity/]             # Optional: domain types
```

---

### 3. Dependency Graph

Service sits at the **bottom** of the module dependency graph. No circular dependencies.

```
handler → service
middleware → service
external modules → middleware (for route protection)
external modules → service (for shared types and accessors)
```

**Incorrect (handler doing data access directly):**

```go
type Handler struct {
    db *sqlite.DB
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    hash, _ := service.HashPassword(input.Password)
    query, args := orm.Insert("users").Set("password_hash", hash).Build()
    h.db.Exec(query, args...)
}
```

**Correct (handler delegates to service):**

```go
type Handler struct {
    svc *service.Service
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    result, err := h.svc.Create(r.Context(), service.CreateInput{...})
    // write HTTP response
}
```

---

### 4. Service Constructor and FX Wiring

The service takes `*sqlite.DB` (and any other dependencies) via constructor injection. Register it in the module's `Provide()`.

```go
// service/service.go
type Service struct {
    db *sqlite.DB
}

func NewService(db *sqlite.DB) *Service {
    return &Service{db: db}
}

// module.go
func Provide() fx.Option {
    return fx.Options(
        fx.Provide(service.NewService),
        fx.Provide(handler.NewHandler),
        fx.Provide(
            fx.Annotate(moduleImpl, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
        ),
    )
}
```

The handler constructor takes `*service.Service` instead of `*sqlite.DB`:

```go
// handler/handler.go
func NewHandler(svc *service.Service) *Handler {
    return &Handler{svc: svc}
}
```

---

### 5. Cross-Module Imports

External modules may import:
- `middleware/` — for route protection (`RequireAuth`, `RequireAdmin`, etc.)
- `service/` — for shared types (`AuthUser`, context accessors) and business operations

External modules should **never** import `handler/` from another module.

---

### 6. Reference Implementation

See `app/auth/` for the canonical example:
- `service/service.go` — session management, authentication, context types
- `service/password.go` — Argon2id hashing (package-level functions)
- `handler/handler.go` — HTTP-only, delegates all logic to service
- `middleware/middleware.go` — imports service for session validation and context types
