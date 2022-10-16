package todoclient

import (
	"time"
)

type ToDoTask struct {
	ID           string
	Name         string
	DueDate      time.Time
	CreationTime time.Time
}

type ToDoParent struct {
	ID   string
	Name string
}

type ToDoClient interface {
	GetAllTasks() (tasks []ToDoTask, err error)
	GetChildrenTasks(parentId string) (tasks []ToDoTask, err error)
	GetAllParents() ([]ToDoParent, error)

	CreateTask(parentId string, task ToDoTask) (ToDoTask, error)
	UpdateTask(parentId string, task ToDoTask) error
	DeleteTask(parentId string, taskId string) error
}
