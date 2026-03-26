package main

import (
	"localhost/app/core/container"
	"localhost/app/user"
)

func main() {
	container.Run(
		user.Provide(),
	)
}
