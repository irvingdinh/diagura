package usermgmt

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/usermgmt/handler"
	"localhost/app/auth/middleware"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func NewModule(h *handler.Handler, mw *middleware.Middleware) *Module {
	return &Module{
		handler: h,
		mw:      mw,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/users", m.mw.RequireAdmin(m.handler.List))
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(handler.NewHandler),
		fx.Provide(NewModule),
	)
}
