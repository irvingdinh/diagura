package usermgmt

import (
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/admin/usermgmt/handler"
)

type Module struct {
	handler *handler.Handler
}

func NewModule(h *handler.Handler) *Module {
	return &Module{
		handler: h,
	}
}

func (m *Module) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/admin/users", m.handler.List)
}

func Provide() fx.Option {
	return fx.Options(
		fx.Provide(handler.NewHandler),
		fx.Provide(NewModule),
	)
}
