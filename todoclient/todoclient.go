// Package todoclient provides interfaces and types for todo service clients
package todoclient

import (
	"context"
	"time"
)

// ToDoTask represents a task in the to-do list, with a due date and creation time.
type ToDoTask struct {
	ID           string    `json:"id"`            // Unique identifier for the task
	Name         string    `json:"name"`          // Short description of the task
	Description  string    `json:"description"`   // Detailed description of the task
	DueDate      time.Time `json:"due_date"`      // When the task is due
	CreationTime time.Time `json:"creation_time"` // When the task was created
	IsCompleted  bool      `json:"is_completed"`  // Whether the task is completed
}

// ToDoParent represents a parent entity, which can contain multiple tasks.
// It can be thought of as a project or a list that holds related tasks.
type ToDoParent struct {
	ID   string `json:"id"`   // Unique identifier for the parent
	Name string `json:"name"` // Name of the parent (project/list)
}

// ToDoClient defines the interface for interacting with a generic to-do service provider.
// Implementations should provide CRUD operations for tasks and parents (projects/lists),
// as well as retrieval of all tasks and parents. Each method should return an error if the
// operation fails, and all returned slices should be non-nil (empty if no results).
type ToDoClient interface {
	// GetAllTasks retrieves all tasks across all parents (projects/lists).
	GetAllTasks(ctx context.Context) ([]ToDoTask, error)

	// GetChildrenTasks retrieves all tasks under a specific parent (project/list).
	GetChildrenTasks(ctx context.Context, parentID string) ([]ToDoTask, error)

	// CreateTask creates a new task under the specified parent (project/list).
	CreateTask(ctx context.Context, parentID string, task ToDoTask) (ToDoTask, error)

	// UpdateTask updates an existing task under the specified parent (project/list).
	UpdateTask(ctx context.Context, parentID string, task ToDoTask) error

	// DeleteTask deletes a task by its ID under the specified parent (project/list).
	DeleteTask(ctx context.Context, parentID, taskID string) error

	// GetAllParents retrieves all parents (projects/lists).
	GetAllParents(ctx context.Context) ([]ToDoParent, error)

	// CreateParent creates a new parent (project/list) with the given name.
	CreateParent(ctx context.Context, parentName string) (ToDoParent, error)

	// DeleteParent deletes a parent (project/list) by its ID.
	DeleteParent(ctx context.Context, parentID string) error
}

// Validate validates a ToDoTask
func (t *ToDoTask) Validate() error {
	if t.Name == "" {
		return &ValidationError{Field: "name", Message: "task name cannot be empty"}
	}
	if len(t.Name) > 500 {
		return &ValidationError{Field: "name", Message: "task name cannot exceed 500 characters"}
	}
	if len(t.Description) > 2000 {
		return &ValidationError{Field: "description", Message: "task description cannot exceed 2000 characters"}
	}
	return nil
}

// Validate validates a ToDoParent
func (p *ToDoParent) Validate() error {
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "parent name cannot be empty"}
	}
	if len(p.Name) > 200 {
		return &ValidationError{Field: "name", Message: "parent name cannot exceed 200 characters"}
	}
	return nil
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
