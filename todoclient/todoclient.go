package todoclient

import (
	"time"
)

// ToDoTask represents a task in the to-do list, with a due date and creation time.
type ToDoTask struct {
	ID           string    // Unique identifier for the task
	Name         string    // Short description of the task
	Description  string    // Detailed description of the task
	DueDate      time.Time // When the task is due
	CreationTime time.Time // When the task was created
}

// ToDoParent represents a parent entity, which can contain multiple tasks.
// It can be thought of as a project or a list that holds related tasks.
type ToDoParent struct {
	ID   string // Unique identifier for the parent
	Name string // Name of the parent (project/list)
}

// ToDoClient defines the interface for interacting with a generic to-do service provider.
// Implementations should provide CRUD operations for tasks and parents (projects/lists),
// as well as retrieval of all tasks and parents. Each method should return an error if the
// operation fails, and all returned slices should be non-nil (empty if no results).
type ToDoClient interface {
	// GetAllTasks retrieves all tasks across all parents (projects/lists).
	GetAllTasks() (tasks []ToDoTask, err error)

	// GetChildrenTasks retrieves all tasks under a specific parent (project/list).
	GetChildrenTasks(parentId string) (tasks []ToDoTask, err error)

	// CreateTask creates a new task under the specified parent (project/list).
	CreateTask(parentId string, task ToDoTask) (ToDoTask, error)

	// UpdateTask updates an existing task under the specified parent (project/list).
	UpdateTask(parentId string, task ToDoTask) error

	// DeleteTask deletes a task by its ID under the specified parent (project/list).
	DeleteTask(parentId string, taskId string) error

	// GetAllParents retrieves all parents (projects/lists).
	GetAllParents() ([]ToDoParent, error)

	// CreateParent creates a new parent (project/list) with the given name.
	CreateParent(parentName string) (ToDoParent, error)

	// DeleteParent deletes a parent (project/list) by its ID.
	DeleteParent(parentId string) error
}
