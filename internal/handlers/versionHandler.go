package handlers

import (
	"github.com/gofiber/fiber/v3"
)

func VersionHandler(c fiber.Ctx) error {
	return c.SendString("version: 1.0.0")
}