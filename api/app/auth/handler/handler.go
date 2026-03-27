package handler

import (
	"encoding/json"
	"log/slog"
	"net"
	nethttp "net/http"
	"strings"
	"time"

	"localhost/app/auth/service"
	"localhost/app/core/config"
	"localhost/app/core/sqlite"
	"localhost/app/core/sqlite/orm"
)

// Handler provides HTTP handlers for the authentication endpoints.
type Handler struct {
	db  *sqlite.DB
	svc *service.Service
}

// NewHandler creates a Handler with the given database and session service.
func NewHandler(db *sqlite.DB, svc *service.Service) *Handler {
	return &Handler{db: db, svc: svc}
}

// Login authenticates a user by email and password, creates a session,
// and returns the session token as a cookie.
func (h *Handler) Login(w nethttp.ResponseWriter, r *nethttp.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}
	if input.Email == "" || input.Password == "" {
		writeError(w, nethttp.StatusBadRequest, "Email and password are required")
		return
	}

	// Look up user by email, join roles for slug.
	query, args := orm.Select("u.id", "u.email", "u.name", "u.password_hash", "r.slug").
		From("users u").
		Join("roles r", "r.id = u.role_id").
		Where("u.email = ?", input.Email).
		Where("u.deleted_at IS NULL").
		Build()

	row := h.db.QueryRow(query, args...)
	var userID, email, name, passwordHash, roleSlug string
	if err := row.Scan(&userID, &email, &name, &passwordHash, &roleSlug); err != nil {
		writeError(w, nethttp.StatusUnauthorized, "Invalid email or password")
		return
	}

	ok, err := service.VerifyPassword(passwordHash, input.Password)
	if err != nil || !ok {
		writeError(w, nethttp.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Transparent parameter upgrade.
	if service.NeedsRehash(passwordHash) {
		if newHash, err := service.HashPassword(input.Password); err == nil {
			q, a := orm.Update("users").
				Set("password_hash", newHash).
				Set("updated_at", orm.FormatTime(time.Now())).
				Where("id = ?", userID).
				Build()
			_, _ = h.db.Exec(q, a...)
		}
	}

	rawToken, err := h.svc.CreateSession(r.Context(), userID, clientIP(r), r.UserAgent())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to create session", "error", err)
		writeError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	nethttp.SetCookie(w, &nethttp.Cookie{
		Name:     "standalone_session",
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
		Secure:   config.GetBool("http.secure"),
		MaxAge:   365 * 24 * 60 * 60,
	})

	slog.InfoContext(r.Context(), "login successful", "user_id", userID, "email", email)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]string{
			"id":    userID,
			"email": email,
			"name":  name,
			"role":  roleSlug,
		},
	})
}

// Session returns the authenticated user's information and extends the
// sliding session window. Protected by RequireAuth middleware.
func (h *Handler) Session(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, _ := service.UserFromContext(r.Context())
	session, _ := service.SessionFromContext(r.Context())

	if err := h.svc.ExtendSession(r.Context(), session.ID); err != nil {
		slog.ErrorContext(r.Context(), "failed to extend session", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]string{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.RoleSlug,
		},
	})
}

// Logout invalidates the current session and clears the cookie.
// Protected by RequireAuth middleware.
func (h *Handler) Logout(w nethttp.ResponseWriter, r *nethttp.Request) {
	session, _ := service.SessionFromContext(r.Context())

	if err := h.svc.DeleteSession(r.Context(), session.ID); err != nil {
		slog.ErrorContext(r.Context(), "failed to delete session", "error", err)
	}

	nethttp.SetCookie(w, &nethttp.Cookie{
		Name:     "standalone_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
		Secure:   config.GetBool("http.secure"),
		MaxAge:   0,
	})

	w.WriteHeader(nethttp.StatusNoContent)
}

func clientIP(r *nethttp.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeError(w nethttp.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
