package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"localhost/app/core/sqlite"
	"localhost/app/core/sqlite/orm"
	"localhost/app/core/utils"
)

// Store is a built-in catch-all subscriber that persists every event
// envelope to the SQLite events table.
type Store struct {
	db *sqlite.DB
}

// NewStore creates an event store backed by the given database.
func NewStore(db *sqlite.DB) *Store {
	return &Store{db: db}
}

func (s *Store) handleEvent(ctx context.Context, env Envelope) {
	data, err := json.Marshal(env.Data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal event data",
			"event", env.Name,
			"error", err,
		)
		return
	}

	query, args := orm.Insert("events").
		Set("id", env.ID).
		Set("name", env.Name).
		Set("actor_id", env.ActorID).
		Set("request_id", env.RequestID).
		Set("ip", env.IP).
		Set("entity_type", env.EntityType).
		Set("entity_id", env.EntityID).
		Set("data", string(data)).
		Set("created_at", utils.FormatTime(env.Time)).
		Build()

	if _, err := s.db.Exec(query, args...); err != nil {
		slog.ErrorContext(ctx, "failed to persist event",
			"event", env.Name,
			"error", err,
		)
	}
}

// StoredEvent represents an event row read from the database.
type StoredEvent struct {
	ID         string `db:"id"          json:"id"`
	Name       string `db:"name"        json:"name"`
	ActorID    string `db:"actor_id"    json:"actor_id"`
	RequestID  string `db:"request_id"  json:"request_id"`
	IP         string `db:"ip"          json:"ip"`
	EntityType string `db:"entity_type" json:"entity_type"`
	EntityID   string `db:"entity_id"   json:"entity_id"`
	Data       string `db:"data"        json:"data"`
	CreatedAt  string `db:"created_at"  json:"created_at"`
}

// ListFilter defines the query parameters for listing events.
type ListFilter struct {
	Name       string
	ActorID    string
	EntityType string
	EntityID   string
	Search     string
	DateFrom   string
	DateTo     string
	Page       int
	PerPage    int
}

// ListResult holds a page of events and the total count.
type ListResult struct {
	Events []StoredEvent
	Total  int
}

// List returns a paginated, filtered list of stored events.
func (s *Store) List(ctx context.Context, f ListFilter) (*ListResult, error) {
	baseSelect := orm.Select(
		"id", "name", "actor_id", "request_id", "ip",
		"entity_type", "entity_id", "data", "created_at",
	).From("events")

	countSelect := orm.Select("COUNT(*)").From("events")

	if f.Name != "" {
		baseSelect = baseSelect.Where("name = ?", f.Name)
		countSelect = countSelect.Where("name = ?", f.Name)
	}
	if f.ActorID != "" {
		baseSelect = baseSelect.Where("actor_id = ?", f.ActorID)
		countSelect = countSelect.Where("actor_id = ?", f.ActorID)
	}
	if f.EntityType != "" {
		baseSelect = baseSelect.Where("entity_type = ?", f.EntityType)
		countSelect = countSelect.Where("entity_type = ?", f.EntityType)
	}
	if f.EntityID != "" {
		baseSelect = baseSelect.Where("entity_id = ?", f.EntityID)
		countSelect = countSelect.Where("entity_id = ?", f.EntityID)
	}
	if f.Search != "" {
		baseSelect = baseSelect.Where("data LIKE ?", "%"+f.Search+"%")
		countSelect = countSelect.Where("data LIKE ?", "%"+f.Search+"%")
	}
	if f.DateFrom != "" {
		baseSelect = baseSelect.Where("created_at >= ?", f.DateFrom+" 00:00:00.000")
		countSelect = countSelect.Where("created_at >= ?", f.DateFrom+" 00:00:00.000")
	}
	if f.DateTo != "" {
		baseSelect = baseSelect.Where("created_at < ?", f.DateTo+" 23:59:59.999")
		countSelect = countSelect.Where("created_at < ?", f.DateTo+" 23:59:59.999")
	}

	countQuery, countArgs := countSelect.Build()
	total, err := orm.QueryVal[int64](s.db, countQuery, countArgs...)
	if err != nil {
		return nil, fmt.Errorf("count events: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	dataQuery, dataArgs := baseSelect.
		OrderBy("created_at", "DESC").
		Limit(f.PerPage).
		Offset(offset).
		Build()

	events, err := orm.QueryAll[StoredEvent](s.db, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	return &ListResult{Events: events, Total: int(total)}, nil
}

type nameRow struct {
	Name string `db:"name"`
}

// AvailableNames returns the distinct event names stored in the database.
func (s *Store) AvailableNames(ctx context.Context) ([]string, error) {
	query, args := orm.Select("DISTINCT name").From("events").OrderBy("name", "ASC").Build()
	rows, err := orm.QueryAll[nameRow](s.db, query, args...)
	if err != nil {
		return nil, fmt.Errorf("available names: %w", err)
	}
	names := make([]string, len(rows))
	for i, r := range rows {
		names[i] = r.Name
	}
	return names, nil
}
