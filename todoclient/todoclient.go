package todoclient

import (
	"time"
)

type ToDoTask struct {
	ID           string
	Name         string
	Description  string
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

	CreateTask(parentId string, task ToDoTask) (ToDoTask, error)
	UpdateTask(parentId string, task ToDoTask) error
	DeleteTask(parentId string, taskId string) error

	GetAllParents() ([]ToDoParent, error)

	CreateParent(parentName string) (ToDoParent, error)
	DeleteParent(parent ToDoParent) error
}
