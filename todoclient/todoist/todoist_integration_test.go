package todoist

import (
	"log"
	"os"
	"testing"
)

func TestTodoistClient_Integration_GetAllTasks(t *testing.T) {
	client := createClient()
	tasks, err := client.GetAllTasks()

	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}
	if len(tasks) <= 0 {
		t.Error("expected more than 0 tasks")
	}
}

func TestTodoistClient_Integration_GetParents(t *testing.T) {
	client := createClient()
	tasks, err := client.GetAllParents()

	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}
	if len(tasks) <= 0 {
		t.Error("expected more than 0 parent")
	}
}

func createClient() *TodoistClient {
	httpClient := NewTodoistHttpClient(os.Getenv("TODOIST_API_TOKEN"))
	return NewTodoistClient(httpClient)
}

// Skips integration test if requirements are not meet
func TestMain(m *testing.M) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		log.Print("Test will be skipped in Github Context")
		return
	}

	if os.Getenv("TODOIST_API_TOKEN") == "" {
		log.Print("Test will be skipped in Github Context")
		return
	}

	m.Run()
}
