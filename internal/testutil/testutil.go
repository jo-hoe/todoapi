// Package testutil provides testing utilities
package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jo-hoe/todoapi/todoclient"
)

// MockToDoClient is a mock implementation of ToDoClient for testing
type MockToDoClient struct {
	tasks   []todoclient.ToDoTask
	parents []todoclient.ToDoParent
}

// NewMockToDoClient creates a new mock client
func NewMockToDoClient() *MockToDoClient {
	return &MockToDoClient{
		tasks:   make([]todoclient.ToDoTask, 0),
		parents: make([]todoclient.ToDoParent, 0),
	}
}

func (m *MockToDoClient) GetAllTasks(ctx context.Context) ([]todoclient.ToDoTask, error) {
	return m.tasks, nil
}

func (m *MockToDoClient) GetChildrenTasks(ctx context.Context, parentID string) ([]todoclient.ToDoTask, error) {
	var result []todoclient.ToDoTask
	for _, task := range m.tasks {
		// In a real implementation, you'd have a parent ID field in the task
		result = append(result, task)
	}
	return result, nil
}

func (m *MockToDoClient) CreateTask(ctx context.Context, parentID string, task todoclient.ToDoTask) (todoclient.ToDoTask, error) {
	task.ID = "mock-id"
	task.CreationTime = time.Now()
	m.tasks = append(m.tasks, task)
	return task, nil
}

func (m *MockToDoClient) UpdateTask(ctx context.Context, parentID string, task todoclient.ToDoTask) error {
	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks[i] = task
			return nil
		}
	}
	return nil
}

func (m *MockToDoClient) DeleteTask(ctx context.Context, parentID, taskID string) error {
	for i, task := range m.tasks {
		if task.ID == taskID {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockToDoClient) GetAllParents(ctx context.Context) ([]todoclient.ToDoParent, error) {
	return m.parents, nil
}

func (m *MockToDoClient) CreateParent(ctx context.Context, parentName string) (todoclient.ToDoParent, error) {
	parent := todoclient.ToDoParent{
		ID:   "mock-parent-id",
		Name: parentName,
	}
	m.parents = append(m.parents, parent)
	return parent, nil
}

func (m *MockToDoClient) DeleteParent(ctx context.Context, parentID string) error {
	for i, parent := range m.parents {
		if parent.ID == parentID {
			m.parents = append(m.parents[:i], m.parents[i+1:]...)
			return nil
		}
	}
	return nil
}

// CreateTestServer creates a test HTTP server for testing
func CreateTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// AssertNoError asserts that an error is nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// AssertError asserts that an error is not nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
