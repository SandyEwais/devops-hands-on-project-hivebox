package main

import (
	"log"

	"github.com/joho/godotenv"
)

type Application struct {
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}
}

func main() {
	app := &Application{}
	router := app.Routes()
	router.Listen(":3000")
}