package profile

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/profile/handler"
	"localhost/app/admin/profile/service"
	"localhost/app/auth/middleware"
)

// Module wires the profile submodule routes.
type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func newModule(h *handler.Handler, mw *middleware.Middleware) *Module {
	return &Module{handler: h, mw: mw}
}

// RegisterRoutes registers profile endpoints on the given mux.
func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/profile", m.mw.RequireAdmin(m.handler.Get))
	mux.HandleFunc("PATCH /api/admin/profile", m.mw.RequireAdmin(m.handler.Update))
	mux.HandleFunc("PUT /api/admin/profile/password", m.mw.RequireAdmin(m.handler.ChangePassword))
	mux.HandleFunc("POST /api/admin/profile/sessions/logout", m.mw.RequireAdmin(m.handler.LogoutOtherSessions))
}

// Provide returns fx options for the profile submodule.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(service.NewService),
		fx.Provide(handler.NewHandler),
		fx.Provide(newModule),
	)
}
