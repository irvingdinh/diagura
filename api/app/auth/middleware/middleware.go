package middleware

import (
	"log/slog"
	nethttp "net/http"

	"localhost/app/auth/service"
	"localhost/app/core/http"
	"localhost/app/core/log"
)

// Middleware provides HTTP middleware for authentication and role-based
// access control.
type Middleware struct {
	svc *service.Service
}

// NewMiddleware creates a Middleware with the given session service.
func NewMiddleware(svc *service.Service) *Middleware {
	return &Middleware{svc: svc}
}

// RequireAuth returns middleware that validates the session cookie.
// Any authenticated user is allowed regardless of role.
func (m *Middleware) RequireAuth(next nethttp.HandlerFunc) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		r, _, ok := m.authenticate(w, r)
		if !ok {
			return
		}
		next(w, r)
	}
}

// RequireRoles returns middleware that validates the session cookie and
// checks the user's role against the allowed list.
func (m *Middleware) RequireRoles(next nethttp.HandlerFunc, roleSlugs ...string) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		r, user, ok := m.authenticate(w, r)
		if !ok {
			return
		}

		allowed := false
		for _, slug := range roleSlugs {
			if user.RoleSlug == slug {
				allowed = true
				break
			}
		}
		if !allowed {
			slog.WarnContext(r.Context(), "forbidden: insufficient role",
				"role", user.RoleSlug,
				"required", roleSlugs,
			)
			http.WriteError(w, nethttp.StatusForbidden, "Forbidden")
			return
		}

		next(w, r)
	}
}

// RequireAdmin returns middleware that allows super_admin and admin users.
func (m *Middleware) RequireAdmin(next nethttp.HandlerFunc) nethttp.HandlerFunc {
	return m.RequireRoles(next, "super_admin", "admin")
}

// authenticate reads the session cookie, validates the session, and sets
// user/session context values. Returns the enriched request, the user, and
// whether authentication succeeded. On failure it writes the 401 response.
func (m *Middleware) authenticate(w nethttp.ResponseWriter, r *nethttp.Request) (*nethttp.Request, *service.AuthUser, bool) {
	cookie, err := r.Cookie("standalone_session")
	if err != nil {
		http.WriteError(w, nethttp.StatusUnauthorized, "Unauthorized")
		return r, nil, false
	}

	tokenHash := service.HashToken(cookie.Value)
	user, session, err := m.svc.ValidateSession(r.Context(), tokenHash)
	if err != nil {
		slog.DebugContext(r.Context(), "session validation failed", "error", err)
		http.WriteError(w, nethttp.StatusUnauthorized, "Unauthorized")
		return r, nil, false
	}

	ctx := r.Context()
	ctx = service.WithUser(ctx, user)
	ctx = service.WithSession(ctx, session)
	ctx = log.WithUserID(ctx, user.ID)

	return r.WithContext(ctx), user, true
}
