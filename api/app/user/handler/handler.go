package handler

import (
	"encoding/json"
	nethttp "net/http"
	"strconv"

	"localhost/app/user/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	users, err := h.svc.List(r.Context(), limit, offset)
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

	result, err := h.svc.Create(r.Context(), service.CreateInput{
		Email:    input.Email,
		Name:     input.Name,
		Password: input.Password,
	})
	if err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
		return
	}

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
