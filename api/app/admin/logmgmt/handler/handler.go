package handler

import (
	"encoding/json"
	"log/slog"
	"math"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"localhost/app/admin/logmgmt/service"
	"localhost/app/core/http"
)

// Handler provides HTTP handlers for the log viewer.
type Handler struct {
	svc *service.Service
}

// NewHandler creates a Handler with the given service.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

var allowedLevels = map[string]bool{
	"DEBUG": true,
	"INFO":  true,
	"WARN":  true,
	"ERROR": true,
}

// List returns a paginated, filtered list of log entries for a given date.
func (h *Handler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	q := r.URL.Query()

	date := q.Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		http.WriteError(w, nethttp.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD")
		return
	}

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

	level := strings.ToUpper(strings.TrimSpace(q.Get("level")))
	if level != "" && !allowedLevels[level] {
		http.WriteError(w, nethttp.StatusBadRequest, "Level must be one of: DEBUG, INFO, WARN, ERROR")
		return
	}

	result, err := h.svc.ListEntries(r.Context(), service.ListFilter{
		Date:    date,
		Level:   level,
		Search:  q.Get("search"),
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list log entries", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(perPage)))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": result.Entries,
		"meta": map[string]int{
			"page":        page,
			"per_page":    perPage,
			"total":       result.Total,
			"total_pages": totalPages,
		},
	})
}

// Dates returns all dates that have log files, sorted newest first.
func (h *Handler) Dates(w nethttp.ResponseWriter, r *nethttp.Request) {
	dates, err := h.svc.AvailableDates()
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list log dates", "error", err)
		http.WriteError(w, nethttp.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": dates,
	})
}
