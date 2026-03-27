package main

import (
	"localhost/app/admin"
	"localhost/app/auth"
	"localhost/app/core/container"
)

func main() {
	container.Run(
		admin.Provide(),
		auth.Provide(),
	)
}
