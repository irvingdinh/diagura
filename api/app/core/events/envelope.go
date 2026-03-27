package events

import (
	"context"
	"time"

	"localhost/app/core/log"
	"localhost/app/core/utils"
)

// Envelope wraps every emitted event with metadata extracted from the
// request context at emit time.
type Envelope struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Time       time.Time `json:"time"`
	ActorID    string    `json:"actor_id"`
	RequestID  string    `json:"request_id"`
	IP         string    `json:"ip"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Data       Event     `json:"-"`
}

// newEnvelope constructs an Envelope by extracting metadata from ctx.
func newEnvelope(ctx context.Context, event Event) Envelope {
	env := Envelope{
		ID:        utils.NewID(),
		Name:      event.EventName(),
		Time:      time.Now(),
		ActorID:   log.UserIDFromCtx(ctx),
		RequestID: log.RequestIDFromCtx(ctx),
		IP:        log.IPFromCtx(ctx),
		Data:      event,
	}

	if ee, ok := event.(EntityEvent); ok {
		env.EntityType = ee.EntityType()
		env.EntityID = ee.EntityID()
	}

	return env
}
