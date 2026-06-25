package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func errorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
	})
}

func createdResponse(c *fiber.Ctx, data interface{}) error {
	return c.Status(201).JSON(data)
}

func parseID(c *fiber.Ctx) (int, error) {
	return strconv.Atoi(c.Params("id"))
}
