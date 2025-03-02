package main

import (
	application "github.com/superlogarifm/goCalc-v2/application"
)

func main() {
	app := application.NewApp()
	app.StartServer()
}
