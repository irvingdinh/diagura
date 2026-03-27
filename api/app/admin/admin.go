package admin

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/handler"
	"localhost/app/admin/logmgmt"
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
	logmgmt  *logmgmt.Module
	mw       *middleware.Middleware
}

func moduleImpl(h *handler.Handler, p *profile.Module, um *usermgmt.Module, lm *logmgmt.Module, mw *middleware.Middleware) *Module {
	return &Module{
		handler:  h,
		profile:  p,
		usermgmt: um,
		logmgmt:  lm,
		mw:       mw,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/dashboard", m.mw.RequireAdmin(m.handler.Dashboard))
	m.profile.RegisterRoutes(mux)
	m.usermgmt.RegisterRoutes(mux)
	m.logmgmt.RegisterRoutes(mux)
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(userservice.NewService),
		profile.Provide(),
		usermgmt.Provide(),
		logmgmt.Provide(),
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(moduleImpl, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
