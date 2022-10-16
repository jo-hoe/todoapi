package todoist

import (
	"os"
	"testing"
	"time"

	"github.com/jo-hoe/todoapi/todoclient"
)

func TestTodoistClient_Integration_GetAllTasks(t *testing.T) {
	checkPrerequisites(t)

	client := createClient()
	tasks, err := client.GetAllTasks()

	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}
	if len(tasks) <= 0 {
		t.Error("expected more than 0 tasks")
	}
}

func TestTodoistClient_Integration_CRUD(t *testing.T) {
	checkPrerequisites(t)

	client := createClient()
	// test get parents
	parents, err := client.GetAllParents()
	if err != nil {
		t.Errorf("could not get parents '%v'", err)
	}
	if len(parents) <= 0 {
		t.Error("expected more than 0 parent")
	}

	// test create
	task, err := client.CreateTask(parents[0].ID, todoclient.ToDoTask{
		Name:    "test",
		DueDate: time.Now(),
	})
	if err != nil {
		t.Errorf("issue creating task '%v'", err)
	}

	// test update
	task.Name = "testUpdate"
	err = client.UpdateTask(parents[0].ID, task)
	if err != nil {
		t.Errorf("issue updating task '%v'", err)
	}

	// test delete
	err = client.DeleteTask(parents[0].ID, task.ID)
	if err != nil {
		t.Errorf("issue updating task '%v'", err)
	}
}

func createClient() *TodoistClient {
	httpClient := NewTodoistHttpClient(os.Getenv("TODOIST_API_TOKEN"))
	return NewTodoistClient(httpClient)
}

// Skips integration test if requirements are not meet
func checkPrerequisites(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Test will be skipped in Github Context")
	}

	if os.Getenv("TODOIST_API_TOKEN") == "" {
		t.Skip("Test will be skipped in Github Context")
	}
}
