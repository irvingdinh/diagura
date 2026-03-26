package user

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/core/http"
	"localhost/app/user/handler"
)

type Module struct {
	handler *handler.Handler
}

func moduleImpl(h *handler.Handler) *Module {
	return &Module{
		handler: h,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/users", m.handler.List)
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(moduleImpl, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
