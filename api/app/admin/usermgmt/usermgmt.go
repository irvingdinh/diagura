package usermgmt

import (
	"net/http"

	"go.uber.org/fx"

	"localhost/app/admin/usermgmt/handler"
	"localhost/app/admin/usermgmt/service"
	"localhost/app/auth/middleware"
)

// Module wires the user management submodule routes.
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

// RegisterRoutes registers user management endpoints on the given mux.
func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/users", m.mw.RequireAdmin(m.handler.List))
	mux.HandleFunc("POST /api/admin/users", m.mw.RequireAdmin(m.handler.Create))
	mux.HandleFunc("GET /api/admin/users/{id}", m.mw.RequireAdmin(m.handler.Get))
	mux.HandleFunc("PATCH /api/admin/users/{id}", m.mw.RequireAdmin(m.handler.Update))
	mux.HandleFunc("PUT /api/admin/users/{id}/password", m.mw.RequireAdmin(m.handler.SetPassword))
	mux.HandleFunc("DELETE /api/admin/users/{id}", m.mw.RequireAdmin(m.handler.Delete))
	mux.HandleFunc("POST /api/admin/users/{id}/restore", m.mw.RequireAdmin(m.handler.Restore))
}

// Provide returns fx options for the user management submodule.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(service.NewService),
		fx.Provide(handler.NewHandler),
		fx.Provide(newModule),
	)
}
