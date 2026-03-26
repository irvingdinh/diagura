package container

import (
	"go.uber.org/fx"

	"localhost/app/core/config"
	"localhost/app/core/http"
)

func Run(opts ...fx.Option) {
	config.Load()

	core := fx.Options(
		fx.Provide(http.NewServer),
		fx.Invoke(func(_ http.Server) {
			config.Validate()
		}),
	)

	fx.New(core, fx.Options(opts...)).Run()
}
