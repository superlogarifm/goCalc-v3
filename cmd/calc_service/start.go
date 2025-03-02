package main

import (
	application "github.com/superlogarifm/goCalc/application"
)

func main() {
	app := application.NewApp()
	app.StartServer()
}
