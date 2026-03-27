package events

import (
	"go.uber.org/fx"

	"localhost/app/core/sqlite"
)

// Provide returns the fx.Option that registers the Bus into the DI
// container. It collects all ListenerRegistrar implementations and
// calls RegisterListeners on each.
func Provide() fx.Option {
	return fx.Provide(func(params struct {
		fx.In
		Registrars []ListenerRegistrar `group:"listeners"`
	}) *Bus {
		bus := NewBus()
		for _, r := range params.Registrars {
			r.RegisterListeners(bus)
		}
		return bus
	})
}

// WithStore returns the fx.Option that registers the EventStore as a
// catch-all listener on the Bus. All emitted events are persisted to
// the SQLite events table.
func WithStore() fx.Option {
	return fx.Options(
		fx.Provide(func(db *sqlite.DB) *Store {
			return NewStore(db)
		}),
		fx.Invoke(func(bus *Bus, store *Store) {
			bus.OnAll(store.handleEvent)
		}),
	)
}
