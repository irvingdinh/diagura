package handler

import (
	"encoding/json"
	nethttp "net/http"
	"strconv"
	"time"

	"localhost/app/core/sqlite"
	"localhost/app/core/sqlite/orm"
	"localhost/app/user/entity"
)

type Handler struct {
	db *sqlite.DB
}

func NewHandler(db *sqlite.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	b := orm.Select("id", "email", "name", "created_at", "updated_at").
		From("users").
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
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		nethttp.Error(w, "invalid request body", nethttp.StatusBadRequest)
		return
	}
	if input.Email == "" {
		nethttp.Error(w, "email is required", nethttp.StatusBadRequest)
		return
	}

	now := orm.FormatTime(time.Now())
	id := orm.NewID()

	query, args := orm.Insert("users").
		Set("id", id).
		Set("email", input.Email).
		Set("name", input.Name).
		Set("created_at", now).
		Set("updated_at", now).
		Build()

	if _, err := h.db.Exec(query, args...); err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
		return
	}

	u := entity.User{
		BaseModel: orm.BaseModel{ID: id},
		Email:     input.Email,
		Name:      input.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": u})
}
