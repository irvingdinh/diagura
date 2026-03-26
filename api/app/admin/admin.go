package admin

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/handler"
	"localhost/app/admin/usermgmt"
	"localhost/app/core/http"
)

type Module struct {
	handler  *handler.Handler
	usermgmt *usermgmt.Module
}

func moduleImpl(h *handler.Handler, um *usermgmt.Module) *Module {
	return &Module{
		handler:  h,
		usermgmt: um,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/dashboard", m.handler.Dashboard)
	m.usermgmt.RegisterRoutes(mux)
}

func Provide() fx.Option {
	return fx.Options(
		usermgmt.Provide(),
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(moduleImpl, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
