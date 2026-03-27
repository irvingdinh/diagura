package admin

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/handler"
	"localhost/app/admin/profile"
	"localhost/app/admin/usermgmt"
	"localhost/app/auth/middleware"
	"localhost/app/core/http"
	userservice "localhost/app/user/service"
)

type Module struct {
	handler  *handler.Handler
	profile  *profile.Module
	usermgmt *usermgmt.Module
	mw       *middleware.Middleware
}

func moduleImpl(h *handler.Handler, p *profile.Module, um *usermgmt.Module, mw *middleware.Middleware) *Module {
	return &Module{
		handler:  h,
		profile:  p,
		usermgmt: um,
		mw:       mw,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/dashboard", m.mw.RequireAdmin(m.handler.Dashboard))
	m.profile.RegisterRoutes(mux)
	m.usermgmt.RegisterRoutes(mux)
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(userservice.NewService),
		profile.Provide(),
		usermgmt.Provide(),
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(moduleImpl, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
