package logmgmt

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/logmgmt/handler"
	"localhost/app/admin/logmgmt/service"
	"localhost/app/auth/middleware"
)

// Module wires the log management submodule routes.
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

// RegisterRoutes registers log viewer endpoints on the given mux.
func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/logs", m.mw.RequireAdmin(m.handler.List))
	mux.HandleFunc("GET /api/admin/logs/dates", m.mw.RequireAdmin(m.handler.Dates))
}

// Provide returns fx options for the log management submodule.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(service.NewService),
		fx.Provide(handler.NewHandler),
		fx.Provide(newModule),
	)
}
