package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type untypedHandler func(ctx context.Context, env Envelope)
type catchAllHandler func(ctx context.Context, env Envelope)

// Bus is the central event dispatch mechanism. It dispatches events
// synchronously and sequentially. Listener errors are recovered and
// logged — they never propagate to the emitter.
type Bus struct {
	mu       sync.RWMutex
	typed    map[string][]untypedHandler
	catchAll []catchAllHandler
}

// NewBus creates an empty event bus.
func NewBus() *Bus {
	return &Bus{
		typed: make(map[string][]untypedHandler),
	}
}

// Emit constructs an Envelope from the context and event, then
// dispatches to all matching typed listeners followed by catch-all
// listeners. Panics in listeners are recovered and logged.
func (b *Bus) Emit(ctx context.Context, event Event) {
	env := newEnvelope(ctx, event)

	b.mu.RLock()
	typed := b.typed[env.Name]
	all := b.catchAll
	b.mu.RUnlock()

	for _, fn := range typed {
		b.safeCall(ctx, env, fn)
	}
	for _, fn := range all {
		b.safeCall(ctx, env, fn)
	}
}

// OnAll registers a catch-all listener that receives every event.
func (b *Bus) OnAll(fn catchAllHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.catchAll = append(b.catchAll, fn)
}

func (b *Bus) register(name string, fn untypedHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.typed[name] = append(b.typed[name], fn)
}

func (b *Bus) safeCall(ctx context.Context, env Envelope, fn func(context.Context, Envelope)) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "event listener panicked",
				"event", env.Name,
				"panic", fmt.Sprintf("%v", r),
			)
		}
	}()
	fn(ctx, env)
}

// On registers a typed listener for events of type T. The listener is
// only called when an event matching T's EventName is emitted.
func On[T Event](bus *Bus, fn func(ctx context.Context, env Envelope, event T)) {
	var zero T
	name := zero.EventName()

	bus.register(name, func(ctx context.Context, env Envelope) {
		if event, ok := env.Data.(T); ok {
			fn(ctx, env, event)
		}
	})
}
