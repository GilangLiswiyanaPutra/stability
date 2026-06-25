package handlers

import (
	"strings"

	"stability-test-task-api/models"
	"stability-test-task-api/store"

	"github.com/gofiber/fiber/v2"
)

func GetTasks(c *fiber.Ctx) error {
	tasks := store.GetAllTasks()
	return c.JSON(tasks)
}

func GetTask(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, 400, "invalid id, must be a number")
	}

	task := store.GetTaskByID(id)
	if task == nil {
		return errorResponse(c, 404, "task not found")
	}

	return c.JSON(task)
}

func CreateTask(c *fiber.Ctx) error {
	var task models.Task

	if err := c.BodyParser(&task); err != nil {
		return errorResponse(c, 400, "invalid request body")
	}

	task.Title = strings.TrimSpace(task.Title)

	if task.Title == "" {
		return errorResponse(c, 400, "title is required")
	}

	if len(task.Title) > 100 {
		return errorResponse(c, 400, "title must not exceed 100 characters")
	}

	task.ID = store.NextID()
	store.AddTask(task)

	return createdResponse(c, task)
}

func UpdateTask(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, 400, "invalid id, must be a number")
	}

	var input models.Task
	if err := c.BodyParser(&input); err != nil {
		return errorResponse(c, 400, "invalid request body")
	}

	input.Title = strings.TrimSpace(input.Title)

	if input.Title == "" {
		return errorResponse(c, 400, "title is required")
	}

	if len(input.Title) > 100 {
		return errorResponse(c, 400, "title must not exceed 100 characters")
	}

	updated := store.UpdateTask(id, input)
	if !updated {
		return errorResponse(c, 404, "task not found")
	}

	task := store.GetTaskByID(id)
	return c.JSON(task)
}

func DeleteTask(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errorResponse(c, 400, "invalid id, must be a number")
	}

	deleted := store.DeleteTask(id)
	if !deleted {
		return errorResponse(c, 404, "task not found")
	}

	return c.JSON(fiber.Map{
		"message": "task deleted successfully",
	})
}
