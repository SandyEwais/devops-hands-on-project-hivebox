package main

import (
	"log"
)

type Application struct {
}

func main() {
	app := &Application{}
	router := app.Routes()
	if err := router.Listen(":3000"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}