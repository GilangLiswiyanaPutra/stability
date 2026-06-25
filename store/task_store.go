package store

import "stability-test-task-api/models"

var Tasks = []models.Task{
	{ID: 1, Title: "Learn Go", Done: false},
	{ID: 2, Title: "Build API", Done: false},
}

var nextID = 3

func NextID() int {
	id := nextID
	nextID++
	return id
}

func GetAllTasks() []models.Task {
	return Tasks
}

func GetTaskByID(id int) *models.Task {
	for i:= range Tasks {
		if Tasks[i].ID == id {
			return &Tasks[i]
		}
	}
	return nil
}

func AddTask(task models.Task) {
	Tasks = append(Tasks, task)
}

func DeleteTask(id int) bool {
	for i, t := range Tasks {
		if t.ID == id {
			Tasks = append(Tasks[:i], Tasks[i+1:]...)
			return true
		}
	}
	return false
}

func UpdateTask(id int, updated models.Task) bool {
	for i := range Tasks {
		if Tasks[i].ID == id {
			Tasks[i].Title = updated.Title
			Tasks[i].Done = updated.Done
			return true
		}
	}
	return false
}

