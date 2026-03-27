package admin

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/eventmgmt"
	"localhost/app/admin/handler"
	"localhost/app/admin/logmgmt"
	"localhost/app/admin/profile"
	"localhost/app/admin/usermgmt"
	"localhost/app/auth/middleware"
	"localhost/app/core/http"
	userservice "localhost/app/user/service"
)

type Module struct {
	handler   *handler.Handler
	profile   *profile.Module
	usermgmt  *usermgmt.Module
	logmgmt   *logmgmt.Module
	eventmgmt *eventmgmt.Module
	mw        *middleware.Middleware
}

func newModule(h *handler.Handler, p *profile.Module, um *usermgmt.Module, lm *logmgmt.Module, em *eventmgmt.Module, mw *middleware.Middleware) *Module {
	return &Module{
		handler:   h,
		profile:   p,
		usermgmt:  um,
		logmgmt:   lm,
		eventmgmt: em,
		mw:        mw,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/dashboard", m.mw.RequireAdmin(m.handler.Dashboard))
	m.profile.RegisterRoutes(mux)
	m.usermgmt.RegisterRoutes(mux)
	m.logmgmt.RegisterRoutes(mux)
	m.eventmgmt.RegisterRoutes(mux)
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(userservice.NewService),
		profile.Provide(),
		usermgmt.Provide(),
		logmgmt.Provide(),
		eventmgmt.Provide(),
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(newModule, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
