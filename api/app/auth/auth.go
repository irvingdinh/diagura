package auth

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/auth/handler"
	"localhost/app/core/http"
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
	mux.HandleFunc("POST /api/auth/login", m.handler.Login)
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(handler.NewHandler),
		fx.Provide(
			fx.Annotate(moduleImpl, fx.As(new(http.RouteRegistrar)), fx.ResultTags(`group:"routes"`)),
		),
	)
}
