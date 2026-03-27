package handler

import (
	"encoding/json"
	nethttp "net/http"
	"strconv"
	"time"

	"localhost/app/auth/service"
	"localhost/app/core/sqlite"
	"localhost/app/core/sqlite/orm"
	"localhost/app/user/entity"
)

const userRoleID = "00000000-0000-7000-0000-000000000002"

type Handler struct {
	db *sqlite.DB
}

func NewHandler(db *sqlite.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	b := orm.Select("id", "role_id", "email", "name", "created_at", "updated_at").
		From("users").
		Where("deleted_at IS NULL").
		OrderBy("created_at", "DESC")
	if limit > 0 {
		b = b.Limit(limit)
	}
	if offset > 0 {
		b = b.Offset(offset)
	}
	query, args := b.Build()

	users, err := orm.QueryAll[entity.User](h.db, query, args...)
	if err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": users})
}

func (h *Handler) Create(w nethttp.ResponseWriter, r *nethttp.Request) {
	var input struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		nethttp.Error(w, "invalid request body", nethttp.StatusBadRequest)
		return
	}
	if input.Email == "" || input.Password == "" {
		nethttp.Error(w, "email and password are required", nethttp.StatusBadRequest)
		return
	}

	passwordHash, err := service.HashPassword(input.Password)
	if err != nil {
		nethttp.Error(w, "internal server error", nethttp.StatusInternalServerError)
		return
	}

	now := orm.FormatTime(time.Now())
	id := orm.NewID()

	query, args := orm.Insert("users").
		Set("id", id).
		Set("role_id", userRoleID).
		Set("email", input.Email).
		Set("name", input.Name).
		Set("password_hash", passwordHash).
		Set("created_at", now).
		Set("updated_at", now).
		Build()

	if _, err := h.db.Exec(query, args...); err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]string{
			"id":    id,
			"email": input.Email,
			"name":  input.Name,
		},
	})
}
