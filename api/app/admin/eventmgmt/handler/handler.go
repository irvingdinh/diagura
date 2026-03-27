package handler

import (
	"encoding/json"
	"log/slog"
	"math"
	nethttp "net/http"
	"strconv"

	"localhost/app/core/events"
	"localhost/app/core/http"
)

// Handler provides HTTP handlers for the event viewer.
type Handler struct {
	store *events.Store
}

// NewHandler creates a Handler with the given event store.
func NewHandler(store *events.Store) *Handler {
	return &Handler{store: store}
}

// List returns a paginated, filtered list of stored events.
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

	result, err := h.store.List(r.Context(), events.ListFilter{
		Name:       q.Get("name"),
		ActorID:    q.Get("actor_id"),
		EntityType: q.Get("entity_type"),
		EntityID:   q.Get("entity_id"),
		Search:     q.Get("search"),
		DateFrom:   q.Get("date_from"),
		DateTo:     q.Get("date_to"),
		Page:       page,
		PerPage:    perPage,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list events", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(perPage)))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": result.Events,
		"meta": map[string]int{
			"page":        page,
			"per_page":    perPage,
			"total":       result.Total,
			"total_pages": totalPages,
		},
	})
}

// Names returns the distinct event names stored in the database.
func (h *Handler) Names(w nethttp.ResponseWriter, r *nethttp.Request) {
	names, err := h.store.AvailableNames(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list event names", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": names,
	})
}
