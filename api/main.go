package main

import (
	"localhost/app/admin"
	"localhost/app/auth"
	"localhost/app/core/container"
	"localhost/app/core/events"
)

func main() {
	container.Run(
		events.Provide(),
		events.WithStore(),
		admin.Provide(),
		auth.Provide(),
	)
}
