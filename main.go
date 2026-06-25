package main

import (
	"stability-test-task-api/handlers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/tasks", handlers.GetTasks)
	app.Get("/tasks/:id", handlers.GetTask)
	app.Post("/tasks", handlers.CreateTask)
	app.Put("/tasks/:id", handlers.UpdateTask)
	app.Delete("/tasks/:id", handlers.DeleteTask)

	app.Listen(":3000")
}
