package container

import (
	"context"

	"go.uber.org/fx"

	"localhost/app/core/config"
	"localhost/app/core/http"
	"localhost/app/core/log"
	"localhost/app/core/sqlite"
	"localhost/database/migrations"
)

func Run(opts ...fx.Option) {
	config.Load()
	log.Load()

	core := fx.Options(
		sqlite.Provide(migrations.FS),
		fx.Provide(http.NewServer),
		fx.Invoke(func(_ http.Server, _ *sqlite.DB) {
			config.Validate()
		}),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStop: func(_ context.Context) error {
					return log.Close()
				},
			})
		}),
	)

	fx.New(core, fx.Options(opts...), log.FxPrinter()).Run()
}
