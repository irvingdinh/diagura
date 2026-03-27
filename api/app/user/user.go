package user

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/auth/middleware"
	"localhost/app/core/http"
	"localhost/app/user/handler"
)

type Module struct {
	handler *handler.Handler
	mw      *middleware.Middleware
}

func moduleImpl(h *handler.Handler, mw *middleware.Middleware) *Module {
	return &Module{
		handler: h,
		mw:      mw,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/users", m.mw.RequireAuth(m.handler.List))
	mux.HandleFunc("POST /api/users", m.mw.RequireAuth(m.handler.Create))
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(moduleImpl, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
