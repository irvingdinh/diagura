package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	nethttp "net/http"

	profileservice "localhost/app/admin/profile/service"
	authevent "localhost/app/auth/event"
	authservice "localhost/app/auth/service"
	"localhost/app/core/events"
	"localhost/app/core/http"
	userevent "localhost/app/user/event"
	userservice "localhost/app/user/service"
)

// Handler provides HTTP handlers for self-profile management.
type Handler struct {
	svc     *profileservice.Service
	userSvc *userservice.Service
	bus     *events.Bus
}

// NewHandler creates a Handler with the given dependencies.
func NewHandler(svc *profileservice.Service, userSvc *userservice.Service, bus *events.Bus) *Handler {
	return &Handler{svc: svc, userSvc: userSvc, bus: bus}
}

// Get returns the authenticated user's profile.
func (h *Handler) Get(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, _ := authservice.UserFromContext(r.Context())

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

// Update modifies the authenticated user's name.
func (h *Handler) Update(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, _ := authservice.UserFromContext(r.Context())

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}
	if input.Name == "" {
		http.WriteError(w, nethttp.StatusBadRequest, "Name is required")
		return
	}

	if err := h.userSvc.UpdateProfile(r.Context(), user.ID, input.Name); err != nil {
		slog.ErrorContext(r.Context(), "failed to update profile", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	h.bus.Emit(r.Context(), userevent.UserUpdated{
		UserID:  user.ID,
		Changes: map[string]any{"name": input.Name},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]any{
			"id":    user.ID,
			"email": user.Email,
			"name":  input.Name,
			"role":  user.RoleSlug,
		},
	})
}

// ChangePassword changes the authenticated user's password after verifying
// the current one, and invalidates all other sessions.
func (h *Handler) ChangePassword(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, _ := authservice.UserFromContext(r.Context())
	session, _ := authservice.SessionFromContext(r.Context())

	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}
	if input.CurrentPassword == "" || input.NewPassword == "" {
		http.WriteError(w, nethttp.StatusBadRequest, "Current password and new password are required")
		return
	}

	err := h.svc.ChangePassword(r.Context(), user.Email, input.CurrentPassword, input.NewPassword, user.ID, session.ID)
	if err != nil {
		if errors.Is(err, profileservice.ErrIncorrectPassword) {
			http.WriteError(w, nethttp.StatusBadRequest, "Current password is incorrect")
			return
		}
		slog.ErrorContext(r.Context(), "failed to change password", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	h.bus.Emit(r.Context(), userevent.UserPasswordChanged{
		UserID: user.ID,
	})

	w.WriteHeader(nethttp.StatusNoContent)
}

// LogoutOtherSessions invalidates all sessions except the current one.
func (h *Handler) LogoutOtherSessions(w nethttp.ResponseWriter, r *nethttp.Request) {
	user, _ := authservice.UserFromContext(r.Context())
	session, _ := authservice.SessionFromContext(r.Context())

	if err := h.svc.LogoutOtherSessions(r.Context(), user.ID, session.ID); err != nil {
		slog.ErrorContext(r.Context(), "failed to logout other sessions", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	h.bus.Emit(r.Context(), authevent.SessionInvalidatedAll{
		UserID: user.ID,
	})

	w.WriteHeader(nethttp.StatusNoContent)
}
