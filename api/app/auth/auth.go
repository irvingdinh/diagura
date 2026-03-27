package auth

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/auth/handler"
	"localhost/app/auth/middleware"
	"localhost/app/auth/service"
	"localhost/app/core/http"
)

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

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", m.handler.Login)
	mux.HandleFunc("GET /api/auth/session", m.mw.RequireAuth(m.handler.Session))
	mux.HandleFunc("POST /api/auth/logout", m.mw.RequireAuth(m.handler.Logout))
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(service.NewService),
		fx.Provide(middleware.NewMiddleware),
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(newModule, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
