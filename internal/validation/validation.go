// Package validation provides input validation utilities
package validation

import (
	"strings"
	"time"

	"github.com/jo-hoe/todoapi/pkg/errors"
)

// ValidateTaskName validates a task name
func ValidateTaskName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.NewValidationError("name", "task name cannot be empty")
	}
	if len(name) > 500 {
		return errors.NewValidationError("name", "task name cannot exceed 500 characters")
	}
	return nil
}

// ValidateTaskDescription validates a task description
func ValidateTaskDescription(description string) error {
	if len(description) > 2000 {
		return errors.NewValidationError("description", "task description cannot exceed 2000 characters")
	}
	return nil
}

// ValidateParentName validates a parent (project/list) name
func ValidateParentName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.NewValidationError("name", "parent name cannot be empty")
	}
	if len(name) > 200 {
		return errors.NewValidationError("name", "parent name cannot exceed 200 characters")
	}
	return nil
}

// ValidateID validates an ID field
func ValidateID(id, fieldName string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.NewValidationError(fieldName, fieldName+" cannot be empty")
	}
	return nil
}

// ValidateDueDate validates a due date
func ValidateDueDate(dueDate time.Time) error {
	if !dueDate.IsZero() && dueDate.Before(time.Now().AddDate(0, 0, -1)) {
		return errors.NewValidationError("dueDate", "due date cannot be in the past")
	}
	return nil
}
