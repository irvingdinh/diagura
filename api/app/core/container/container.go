package container

import (
	"go.uber.org/fx"

	"localhost/app/core/config"
)

func Run() {
	config.Load()

	fx.New(
		fx.Invoke(func() {
			config.Validate()
		}),
	).Run()
}
