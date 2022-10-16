package todoist

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jo-hoe/todoapi/todoclient"
)

func TestTodoistClient_ImplementationTest(t *testing.T) {
	// tests if interface is implemented
	var _ todoclient.ToDoClient = (*TodoistClient)(nil)
}

func TestTodoistClient_GetAllTasks(t *testing.T) {
	client := NewTodoistClient(createMockClient(demoList))

	tasks, err := client.GetAllTasks()

	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected %d tasks but found %d", 3, len(tasks))
	}
}

func TestTodoistClient_GetChildrenTasks(t *testing.T) {
	mockProjectId := "2180393141"
	client := NewTodoistClient(createMockClient(demoListProject))

	tasks, err := client.GetChildrenTasks(mockProjectId)

	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected %d tasks but found %d", 1, len(tasks))
	}
}

func TestTodoistClient_UpdateTask(t *testing.T) {
	client := NewTodoistClient(createMockClient(demoListProject))
	task := todoclient.ToDoTask{
		ID:           "5196276900",
		Name:         "mockTitle",
		DueDate:      time.Now(),
		CreationTime: time.Now(),
	}

	err := client.UpdateTask("2180393145", task)

	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}
}

func TestTodoistClient_UpdateTask_Without_CreationTime(t *testing.T) {
	client := NewTodoistClient(createMockClient(demoListProject))
	task := todoclient.ToDoTask{
		ID:           "5196276900",
		Name:         "mockTitle",
		DueDate:      time.Time{},
		CreationTime: time.Now(),
	}

	err := client.UpdateTask("2180393145", task)

	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}
}

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

func createMockClient(bodies ...string) *http.Client {
	i := -1
	return NewMockClient(func(_ *http.Request) *http.Response {
		i = i + 1
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewBufferString(bodies[i])),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
}

const demoList = `[
	{
			"id": "5196276900",
			"assigner": "0",
			"project_id": "2180393141",
			"section_id": "0",
			"order": "56",
			"content": "buff",
			"description": "",
			"is_completed": false,
			"label_ids": [],
			"priority": "1",
			"comment_count": "0",
			"creator": 16460291,
			"created": "2021-09-28T23:07:29Z",
			"url": "https://todoist.com/showTask?id=5196276900"
	},
	{
			"id": "5207162814",
			"assigner": "0",
			"project_id": "2180393145",
			"section_id": "0",
			"order": "57",
			"content": "stuff",
			"description": "",
			"is_completed": false,
			"label_ids": [],
			"priority": "1",
			"comment_count": "0",
			"creator": "16460291",
			"created": "2021-10-02T18:57:07Z",
			"due": {
					"recurring": false,
					"string": "Dec 1",
					"date": "2021-12-01"
			},
			"url": "https://todoist.com/showTask?id=5207162814"
	},
	{
			"id": "5210371268",
			"assigner": "0",
			"project_id": "2180393145",
			"section_id": "0",
			"order": "58",
			"content": "things",
			"description": "",
			"is_completed": false,
			"label_ids": [],
			"priority": "1",
			"comment_count": "0",
			"creator": "16460291",
			"created": "2021-10-04T07:42:51Z",
			"url": "https://todoist.com/showTask?id=5210371268"
	}
]`

const demoListProject = `[
	{
			"id": "5196276900",
			"assigner": "0",
			"project_id": "2180393141",
			"section_id": "0",
			"order": "56",
			"content": "buffy",
			"description": "",
			"is_completed": false,
			"label_ids": [],
			"priority": "1",
			"comment_count": "0",
			"creator": "16460291",
			"created": "2021-09-28T23:07:29Z",
			"url": "https://todoist.com/showTask?id=5196276900"
	}
]`

func Test_NewTodoistHttpClient(t *testing.T) {
	var token = "dummyToken"
	client := NewTodoistHttpClient(token)

	if client.Transport == nil {
		t.Error("client.Transport was nil")
	}
}
