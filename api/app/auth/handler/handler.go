package handler

import (
	"encoding/json"
	"log/slog"
	nethttp "net/http"

	authevent "localhost/app/auth/event"
	"localhost/app/auth/service"
	"localhost/app/core/config"
	"localhost/app/core/events"
	"localhost/app/core/http"
)

// Handler provides HTTP handlers for the authentication endpoints.
type Handler struct {
	svc *service.Service
	bus *events.Bus
}

// NewHandler creates a Handler with the given session service.
func NewHandler(svc *service.Service, bus *events.Bus) *Handler {
	return &Handler{svc: svc, bus: bus}
}

// Login authenticates a user by email and password, creates a session,
// and returns the session token as a cookie.
func (h *Handler) Login(w nethttp.ResponseWriter, r *nethttp.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}
	if input.Email == "" || input.Password == "" {
		http.WriteError(w, nethttp.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := h.svc.AuthenticateByEmail(r.Context(), input.Email, input.Password)
	if err != nil {
		http.WriteError(w, nethttp.StatusUnauthorized, "Invalid email or password")
		return
	}

	rawToken, err := h.svc.CreateSession(r.Context(), user.ID, http.ClientIP(r), r.UserAgent())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to create session", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
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

	slog.InfoContext(r.Context(), "login successful", "user_id", user.ID, "email", user.Email)

	h.bus.Emit(r.Context(), authevent.AuthLogin{
		UserID: user.ID,
		Email:  user.Email,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"id":                    user.ID,
			"email":                 user.Email,
			"name":                  user.Name,
			"role":                  user.RoleSlug,
			"force_password_change": user.ForcePasswordChange,
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
		"data": map[string]any{
			"id":                    user.ID,
			"email":                 user.Email,
			"name":                  user.Name,
			"role":                  user.RoleSlug,
			"force_password_change": user.ForcePasswordChange,
		},
	})
}

// Logout invalidates the current session and clears the cookie.
// Protected by RequireAuth middleware.
func (h *Handler) Logout(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, _ := service.UserFromContext(r.Context())
	session, _ := service.SessionFromContext(r.Context())

	if err := h.svc.DeleteSession(r.Context(), session.ID); err != nil {
		slog.ErrorContext(r.Context(), "failed to delete session", "error", err)
	}

	h.bus.Emit(r.Context(), authevent.AuthLogout{
		UserID:    user.ID,
		SessionID: session.ID,
	})

	nethttp.SetCookie(w, &nethttp.Cookie{
		Name:     "standalone_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
		Secure:   config.GetBool("http.secure"),
		MaxAge:   -1,
	})

	w.WriteHeader(nethttp.StatusNoContent)
}
