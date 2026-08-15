package main

import (
	"github.com/gofiber/fiber/v3"
	"hivebox/internal/handlers"
)


func (app *Application) Routes() *fiber.App {
	router := fiber.New()
	router.Get("/version", handlers.VersionHandler)
	router.Get("/temperature", handlers.TemperatureHandler)
	return router
}