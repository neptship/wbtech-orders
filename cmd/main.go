package main

import "github.com/neptship/wbtech-orders/internal/application"

func main() {
	app := application.New()
	app.RunServer()
}
