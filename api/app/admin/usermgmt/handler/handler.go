package handler

import (
	"encoding/json"
	"log/slog"
	"math"
	nethttp "net/http"
	"strconv"

	usermgmtservice "localhost/app/admin/usermgmt/service"
	authevent "localhost/app/auth/event"
	authservice "localhost/app/auth/service"
	"localhost/app/core/events"
	"localhost/app/core/http"
	userevent "localhost/app/user/event"
	userservice "localhost/app/user/service"
)

// Handler provides HTTP handlers for admin user management.
type Handler struct {
	svc     *usermgmtservice.Service
	userSvc *userservice.Service
	bus     *events.Bus
}

// NewHandler creates a Handler with the given dependencies.
func NewHandler(svc *usermgmtservice.Service, userSvc *userservice.Service, bus *events.Bus) *Handler {
	return &Handler{svc: svc, userSvc: userSvc, bus: bus}
}

// canManageRole checks whether the acting user can manage users with the given role slug.
func canManageRole(actorRoleSlug, targetRoleSlug string) bool {
	if actorRoleSlug == "super_admin" {
		return true
	}
	return actorRoleSlug == "admin" && targetRoleSlug == "user"
}

// List returns a paginated, filtered list of users.
func (h *Handler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	status := q.Get("status")
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "deleted" {
		http.WriteError(w, nethttp.StatusBadRequest, "Status must be 'active' or 'deleted'")
		return
	}

	result, err := h.userSvc.ListPaginated(r.Context(), userservice.ListFilter{
		Search:  q.Get("search"),
		Role:    q.Get("role"),
		Status:  status,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list users", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(perPage)))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": result.Users,
		"meta": map[string]int{
			"page":        page,
			"per_page":    perPage,
			"total":       result.Total,
			"total_pages": totalPages,
		},
	})
}

// Create registers a new user with the specified role.
func (h *Handler) Create(w nethttp.ResponseWriter, r *nethttp.Request) {
	actor, _ := authservice.UserFromContext(r.Context())

	var input struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}
	if input.Email == "" || input.Name == "" || input.Password == "" || input.Role == "" {
		http.WriteError(w, nethttp.StatusBadRequest, "Email, name, password, and role are required")
		return
	}

	if !canManageRole(actor.RoleSlug, input.Role) {
		http.WriteError(w, nethttp.StatusForbidden, "You do not have permission to assign this role")
		return
	}

	role, err := h.userSvc.GetRoleBySlug(r.Context(), input.Role)
	if err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid role")
		return
	}

	exists, err := h.userSvc.EmailExists(r.Context(), input.Email, "")
	if err != nil {
		slog.ErrorContext(r.Context(), "email check failed", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}
	if exists {
		http.WriteError(w, nethttp.StatusConflict, "Email already in use")
		return
	}

	result, err := h.userSvc.Create(r.Context(), userservice.CreateInput{
		Email:    input.Email,
		Name:     input.Name,
		Password: input.Password,
		RoleID:   role.ID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to create user", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	h.bus.Emit(r.Context(), userevent.UserCreated{
		UserID: result.ID,
		Email:  result.Email,
		Name:   result.Name,
		Role:   input.Role,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]string{
			"id":    result.ID,
			"email": result.Email,
			"name":  result.Name,
		},
	})
}

// Get returns a single user by ID.
func (h *Handler) Get(w nethttp.ResponseWriter, r *nethttp.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.WriteError(w, nethttp.StatusBadRequest, "User ID is required")
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		http.WriteError(w, nethttp.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": user})
}

// Update modifies user fields (name, email, role). All fields are optional.
func (h *Handler) Update(w nethttp.ResponseWriter, r *nethttp.Request) {
	actor, _ := authservice.UserFromContext(r.Context())
	id := r.PathValue("id")

	if id == actor.ID {
		http.WriteError(w, nethttp.StatusBadRequest, "Use the profile endpoint to edit yourself")
		return
	}

	target, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		http.WriteError(w, nethttp.StatusNotFound, "User not found")
		return
	}

	if !canManageRole(actor.RoleSlug, target.RoleSlug) {
		http.WriteError(w, nethttp.StatusForbidden, "You do not have permission to edit this user")
		return
	}

	var input struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
		Role  *string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}

	updateInput := userservice.UpdateUserInput{}

	if input.Name != nil {
		if *input.Name == "" {
			http.WriteError(w, nethttp.StatusBadRequest, "Name cannot be empty")
			return
		}
		updateInput.Name = input.Name
	}

	if input.Email != nil {
		if *input.Email == "" {
			http.WriteError(w, nethttp.StatusBadRequest, "Email cannot be empty")
			return
		}
		exists, err := h.userSvc.EmailExists(r.Context(), *input.Email, id)
		if err != nil {
			slog.ErrorContext(r.Context(), "email check failed", "error", err)
			http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
			return
		}
		if exists {
			http.WriteError(w, nethttp.StatusConflict, "Email already in use")
			return
		}
		updateInput.Email = input.Email
	}

	if input.Role != nil {
		if !canManageRole(actor.RoleSlug, *input.Role) {
			http.WriteError(w, nethttp.StatusForbidden, "You do not have permission to assign this role")
			return
		}
		role, err := h.userSvc.GetRoleBySlug(r.Context(), *input.Role)
		if err != nil {
			http.WriteError(w, nethttp.StatusBadRequest, "Invalid role")
			return
		}
		updateInput.RoleID = &role.ID
	}

	if updateInput.Name == nil && updateInput.Email == nil && updateInput.RoleID == nil {
		http.WriteError(w, nethttp.StatusBadRequest, "No fields to update")
		return
	}

	if err := h.userSvc.UpdateUser(r.Context(), id, updateInput); err != nil {
		slog.ErrorContext(r.Context(), "failed to update user", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	updated, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch updated user", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	changes := make(map[string]any)
	if input.Name != nil {
		changes["name"] = *input.Name
	}
	if input.Email != nil {
		changes["email"] = *input.Email
	}
	if input.Role != nil {
		changes["role"] = *input.Role
	}
	h.bus.Emit(r.Context(), userevent.UserUpdated{
		UserID:  id,
		Changes: changes,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": updated})
}

// SetPassword sets a temporary password for a user and invalidates all sessions.
func (h *Handler) SetPassword(w nethttp.ResponseWriter, r *nethttp.Request) {
	actor, _ := authservice.UserFromContext(r.Context())
	id := r.PathValue("id")

	if id == actor.ID {
		http.WriteError(w, nethttp.StatusBadRequest, "Use the profile endpoint to change your own password")
		return
	}

	target, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		http.WriteError(w, nethttp.StatusNotFound, "User not found")
		return
	}

	if !canManageRole(actor.RoleSlug, target.RoleSlug) {
		http.WriteError(w, nethttp.StatusForbidden, "You do not have permission to change this user's password")
		return
	}

	var input struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid request body")
		return
	}
	if input.Password == "" {
		http.WriteError(w, nethttp.StatusBadRequest, "Password is required")
		return
	}

	hash, err := authservice.HashPassword(input.Password)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to hash password", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	if err := h.userSvc.SetPassword(r.Context(), id, hash, true); err != nil {
		slog.ErrorContext(r.Context(), "failed to set password", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	if err := h.svc.InvalidateAllSessions(r.Context(), target.ID); err != nil {
		slog.ErrorContext(r.Context(), "failed to invalidate sessions", "error", err)
	} else {
		h.bus.Emit(r.Context(), authevent.SessionInvalidatedAll{
			UserID: id,
		})
	}

	h.bus.Emit(r.Context(), userevent.UserPasswordSet{
		UserID: id,
		SetBy:  actor.ID,
	})

	w.WriteHeader(nethttp.StatusNoContent)
}

// Delete soft-deletes a user and invalidates all their sessions.
func (h *Handler) Delete(w nethttp.ResponseWriter, r *nethttp.Request) {
	actor, _ := authservice.UserFromContext(r.Context())
	id := r.PathValue("id")

	if id == actor.ID {
		http.WriteError(w, nethttp.StatusBadRequest, "Cannot delete yourself")
		return
	}

	target, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		http.WriteError(w, nethttp.StatusNotFound, "User not found")
		return
	}

	if target.DeletedAt != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "User is already deleted")
		return
	}

	if !canManageRole(actor.RoleSlug, target.RoleSlug) {
		http.WriteError(w, nethttp.StatusForbidden, "You do not have permission to delete this user")
		return
	}

	if err := h.userSvc.SoftDelete(r.Context(), id); err != nil {
		slog.ErrorContext(r.Context(), "failed to delete user", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	if err := h.svc.InvalidateAllSessions(r.Context(), id); err != nil {
		slog.ErrorContext(r.Context(), "failed to invalidate sessions", "error", err)
	} else {
		h.bus.Emit(r.Context(), authevent.SessionInvalidatedAll{
			UserID: id,
		})
	}

	h.bus.Emit(r.Context(), userevent.UserDeleted{
		UserID: id,
		Email:  target.Email,
	})

	w.WriteHeader(nethttp.StatusNoContent)
}

// Restore restores a soft-deleted user and sets force_password_change.
func (h *Handler) Restore(w nethttp.ResponseWriter, r *nethttp.Request) {
	actor, _ := authservice.UserFromContext(r.Context())
	id := r.PathValue("id")

	target, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		http.WriteError(w, nethttp.StatusNotFound, "User not found")
		return
	}

	if target.DeletedAt == nil {
		http.WriteError(w, nethttp.StatusBadRequest, "User is not deleted")
		return
	}

	if !canManageRole(actor.RoleSlug, target.RoleSlug) {
		http.WriteError(w, nethttp.StatusForbidden, "You do not have permission to restore this user")
		return
	}

	if err := h.userSvc.Restore(r.Context(), id); err != nil {
		slog.ErrorContext(r.Context(), "failed to restore user", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	h.bus.Emit(r.Context(), userevent.UserRestored{
		UserID: id,
		Email:  target.Email,
	})

	w.WriteHeader(nethttp.StatusNoContent)
}
