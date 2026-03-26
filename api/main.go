package main

import (
	"localhost/app/admin"
	"localhost/app/auth"
	"localhost/app/core/container"
	"localhost/app/user"
)

func main() {
	container.Run(
		admin.Provide(),
		auth.Provide(),
		user.Provide(),
	)
}
