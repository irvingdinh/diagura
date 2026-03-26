package container

import (
	"go.uber.org/fx"

	"localhost/app/core/config"
	"localhost/app/core/http"
	"localhost/app/core/sqlite"
	"localhost/database/migrations"
)

func Run(opts ...fx.Option) {
	config.Load()

	core := fx.Options(
		sqlite.Provide(migrations.FS),
		fx.Provide(http.NewServer),
		fx.Invoke(func(_ http.Server, _ *sqlite.DB) {
			config.Validate()
		}),
	)

	fx.New(core, fx.Options(opts...)).Run()
}
