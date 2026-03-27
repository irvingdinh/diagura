package eventmgmt

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/eventmgmt/handler"
	"localhost/app/auth/middleware"
)

// Module wires the event viewer submodule routes.
type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func newModule(h *handler.Handler, mw *middleware.Middleware) *Module {
	return &Module{
		handler: h,
		mw:      mw,
	}
}

// RegisterRoutes registers event viewer endpoints on the given mux.
func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/events", m.mw.RequireAdmin(m.handler.List))
	mux.HandleFunc("GET /api/admin/events/names", m.mw.RequireAdmin(m.handler.Names))
}

// Provide returns fx options for the event viewer submodule.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(handler.NewHandler),
		fx.Provide(newModule),
	)
}
