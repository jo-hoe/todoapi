package microsoft

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jo-hoe/todoapi/todoclient"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewMockClient returns *http.Client with Transport replaced to avoid making real calls
func NewMockClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}


func TestTodoistClient_ImplementationTest(t *testing.T) {
	// tests if interface is implemented
	var _ todoclient.ToDoClient = (*MSToDo)(nil)
}

func TestMSToDo_GetAllTask(t *testing.T) {
	client := createMockClient()

	api := NewMSToDo(client)

	tasks, err := api.GetAllTasks()
	if len(tasks) != 6 {
		t.Errorf("Expected 6 but found %d tasks", len(tasks))
	}
	if err != nil {
		t.Errorf("Found error: '%v'", err)
	}
}

func TestMSToDo_GetChildrenTasks(t *testing.T) {
	client := createMockClient()

	api := NewMSToDo(client)

	tasks, err := api.GetChildrenTasks("xyz")
	if len(tasks) != 3 {
		t.Errorf("Expected 3 but found %d tasks", len(tasks))
	}
	if err != nil {
		t.Errorf("Found error: '%v'", err)
	}
}

func TestMSToDo_GetAllTask_CheckSerialization(t *testing.T) {
	client := createMockClient()
	testTime := time.Date(1990, time.Month(1), 1, 1, 1, 0, 0, time.UTC)
	api := NewMSToDo(client)

	tasks, err := api.GetAllTasks()

	task := tasks[0]

	if len(task.ID) == 0 {
		t.Error("ID is empty")
	}
	if len(task.Title) == 0 {
		t.Error("Title is empty")
	}
	if testTime.After(task.DueDate) {
		t.Errorf("DueDate has unexpected value %s", task.DueDate)
	}
	if testTime.After(task.CreationTime) {
		t.Errorf("CreationTime has unexpected value %s", task.CreationTime)
	}
	if err != nil {
		t.Errorf("Found error: '%v'", err)
	}
}

func TestMSToDo_UpdateTask(t *testing.T) {
	client := createMockClient()

	api := NewMSToDo(client)

	err := api.UpdateTask(todoclient.ToDoTask{
		ID:      "atask",
		Title:   "test",
		DueDate: time.Now(),
	})

	if err != nil {
		t.Errorf("Error: %s", err)
	}
}

var firstTaskCall = true

func createMockClient() *http.Client {
	return NewMockClient(func(req *http.Request) *http.Response {
		// Test request parameters
		body := ""
		if strings.HasSuffix(req.URL.String(), "lists/") {
			body = demoList
		} else if strings.Contains(req.URL.String(), "tasks") {
			if firstTaskCall {
				body = demoTasks1
				firstTaskCall = false
			} else {
				body = demoTasks2
				firstTaskCall = true
			}

		}
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewBufferString(body)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
}

const demoList = `{
	"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users('something.something%40outlook.com')/todo/lists",
	"value": [
		{
			"@odata.etag": "abc",
			"displayName": "Alist",
			"id": "xyz",
			"isOwner": true,
			"isShared": false,
			"wellknownListName": "none"
		},
		{
			"@odata.etag": "def",
			"displayName": "Blist",
			"id": "zyx",
			"isOwner": true,
			"isShared": false,
			"wellknownListName": "none"
		}
	]
}`

const demoTasks1 = `{
    "@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users('someone.someone%40outlook.com')/todo/lists('demo')/tasks",
    "@odata.nextLink": "https://graph.microsoft.com/v1.0/me/todo/lists/demo/tasks?$top=100skip=100",
    "value": [{
            "@odata.etag": "demo",
            "body": {
                "content": "",
                "contentType": "text"
            },
            "completedDateTime": {
                "dateTime": "2021-04-04T00:00:00.0000000",
                "timeZone": "UTC"
            },
            "dueDateTime": {
                "dateTime": "2021-04-04T22:00:00.0000000",
                "timeZone": "UTC"
            },
            "createdDateTime": "2021-04-04T10:27:46.6543589Z",
            "id": "atask",
            "importance": "high",
            "isReminderOn": false,
            "lastModifiedDateTime": "2021-04-04T11:53:53.2660551Z",
            "status": "completed",
            "title": "Banana"
        }, {
            "@odata.etag": "demo2",
            "body": {
                "content": "",
                "contentType": "text"
            },
            "completedDateTime": {
                "dateTime": "2021-04-04T00:00:00.0000000",
                "timeZone": "UTC"
            },
            "createdDateTime": "2021-04-04T10:27:41.5936152Z",
            "id": "btask",
            "importance": "high",
            "isReminderOn": false,
            "lastModifiedDateTime": "2021-04-04T11:53:52.9365041Z",
            "status": "completed",
            "title": "Raspberry"
        }
    ]
}`

const demoTasks2 = `{
    "@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users('someone.someone%40outlook.com')/todo/lists('demo')/tasks",
    "value": [{
            "@odata.etag": "demo",
            "body": {
                "content": "",
                "contentType": "text"
            },
            "completedDateTime": {
                "dateTime": "2021-04-04T00:00:00.0000000",
                "timeZone": "UTC"
            },
            "dueDateTime": {
                "dateTime": "2021-04-04T22:00:00.0000000",
                "timeZone": "UTC"
            },
            "createdDateTime": "2021-04-04T10:27:46.6543589Z",
            "id": "btask",
            "importance": "high",
            "isReminderOn": false,
            "lastModifiedDateTime": "2021-04-04T11:53:53.2660551Z",
            "status": "completed",
            "title": "Tomato"
        }
    ]
}`
