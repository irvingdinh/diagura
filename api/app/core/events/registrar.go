package events

// ListenerRegistrar is implemented by modules that want to register
// event listeners. Collected via Fx group tag and called during bus
// initialization — the same pattern as http.RouteRegistrar.
type ListenerRegistrar interface {
	RegisterListeners(bus *Bus)
}
