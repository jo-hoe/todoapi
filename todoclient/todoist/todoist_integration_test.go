package todoist

import (
	"os"
	"testing"

	"github.com/jo-hoe/todoapi/todoclient/testutil"
)

func TestTodoistClient_Integration_GetAllTasks(t *testing.T) {
	testutil.IntegrationTest_GetAllTasks(t, createClient(t))
}

func TestTodoistClient_Integration_CRUD(t *testing.T) {
	testutil.IntegrationTest_CRUD(t, createClient(t))
}

func createClient(t *testing.T) *TodoistClient {
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		t.Skip("Test will be skipped in Github Context")
	}

	httpClient := NewTodoistHttpClient(token)
	return NewTodoistClient(httpClient)
}
