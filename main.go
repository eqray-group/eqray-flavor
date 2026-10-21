package main

import (
	"eqray-flavor/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
