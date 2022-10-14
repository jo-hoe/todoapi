package todoclient

import (
	"time"
)

type ToDoTask struct {
	ID           string
	Title        string
	DueDate      time.Time
	CreationTime time.Time
}

type ToDoParent struct {
	ID    string
	Title string
}

type ToDoClient interface {
	GetAllTasks() (tasks []ToDoTask, err error)
	GetChildrenTasks(parentId string) (tasks []ToDoTask, err error)
	GetAllParents() (ToDoParent, error)

	CreateTask(task ToDoTask) (ToDoTask, error)
	UpdateTask(task ToDoTask) error
	DeleteTask(task ToDoTask) error
}
