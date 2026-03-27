package events

// Event is the interface all domain events must implement.
type Event interface {
	EventName() string
}

// EntityEvent is an optional extension for events that relate to a
// specific entity. When implemented, the store indexes entity_type and
// entity_id for efficient per-entity queries.
type EntityEvent interface {
	Event
	EntityType() string
	EntityID() string
}
