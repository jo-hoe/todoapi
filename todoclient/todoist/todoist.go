package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	customhttp "github.com/jo-hoe/todoapi/http"
	"github.com/jo-hoe/todoapi/todoclient"
)

type TodoistClient struct {
	httpClient *http.Client
}

type TodoistTask struct {
	ID      uint64      `json:"id,omitempty"`
	Content string      `json:"content,omitempty"`
	Created time.Time   `json:"created,omitempty" examples:"2020-04-06T07:31:55Z"`
	Due     *TodoistDue `json:"due,omitempty"`
}

type TodoistProject struct {
	ID   uint64 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type TodoistDue struct {
	Date string `json:"date,omitempty" examples:"2021-11-08"`
}

const (
	todoistURL        = "https://api.todoist.com/rest/v1/"
	todoistTasksUrl   = todoistURL + "tasks"
	todoistParentsUrl = todoistURL + "projects"
	timeDueDateLayout = "2006-01-02"
)

// creates an http client with injects the REST API token for each request
func NewTodoistHttpClient(token string) *http.Client {
	return customhttp.NewHttpClientWithHeader("Authorization", "Bearer "+token)
}

func NewTodoistClient(httpClient *http.Client) *TodoistClient {
	client := &TodoistClient{
		httpClient: httpClient,
	}
	return client
}

func (client *TodoistClient) CreateTask(task todoclient.ToDoTask) (tasks todoclient.ToDoTask, err error) {
	return todoclient.ToDoTask{}, nil
}

func (client *TodoistClient) DeleteTask(task todoclient.ToDoTask) error {
	return nil
}

func (client *TodoistClient) GetAllParents() ([]todoclient.ToDoParent, error) {
	result := make([]todoclient.ToDoParent, 0)
	projects := make([]TodoistProject, 0)

	err := client.getData(todoistParentsUrl, &projects)
	if err != nil {
		return result, err
	}

	for _, project := range projects {
		parent := todoclient.ToDoParent{
			ID:   fmt.Sprint(project.ID),
			Name: project.Name,
		}
		result = append(result, parent)
	}

	return result, nil
}

func (client *TodoistClient) GetAllTasks() (tasks []todoclient.ToDoTask, err error) {
	return client.getTasks(nil)
}

func (client *TodoistClient) GetChildrenTasks(parentId string) (tasks []todoclient.ToDoTask, err error) {
	return client.getTasks(&parentId)
}

func (client *TodoistClient) getTasks(parentId *string) (tasks []todoclient.ToDoTask, err error) {
	todoistTasks := make([]TodoistTask, 0)

	url := todoistTasksUrl
	if parentId != nil {
		url = url + "?project_id=" + *parentId
	}

	err = client.getData(url, &todoistTasks)
	if err != nil {
		return nil, err
	}

	tasks = make([]todoclient.ToDoTask, 0)
	for _, task := range todoistTasks {
		convertedTask, err := convertToToDoTask(task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *convertedTask)
	}

	return tasks, nil
}

func (client *TodoistClient) UpdateTask(task todoclient.ToDoTask) error {
	payload, err := convertTodoistTask(task)
	if err != nil {
		return err
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := client.httpClient.Post(todoistTasksUrl+"/"+task.ID, "application/json", bytes.NewBuffer(jsonPayload))
	if resp.StatusCode != 204 {
		return fmt.Errorf("%+v", resp.Status)
	}

	return err
}

func (client *TodoistClient) getData(url string, data interface{}) error {
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("error after GET call:%s", err.Error())
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("did not get a 200 response but found: %s", resp.Status)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return fmt.Errorf("could not decode data :%v", err)
	}

	return nil
}

func convertToToDoTask(task TodoistTask) (*todoclient.ToDoTask, error) {
	dueDate := time.Time{}
	if task.Due != nil {
		deserializedTime, err := time.Parse(timeDueDateLayout, task.Due.Date)
		if err == nil {
			dueDate = deserializedTime
		} else {
			dueDate = time.Time{}
		}
	}
	result := todoclient.ToDoTask{
		ID:           strconv.FormatUint(task.ID, 10),
		Name:        task.Content,
		DueDate:      dueDate,
		CreationTime: task.Created,
	}
	return &result, nil
}

// converts but does not convert creation date
func convertTodoistTask(task todoclient.ToDoTask) (*TodoistTask, error) {
	id, err := strconv.ParseUint(string(task.ID), 10, 64)
	if err != nil {
		return nil, err
	}
	result := TodoistTask{
		ID:      id,
		Content: task.Name,
	}

	if !task.CreationTime.IsZero() {
		result.Due = &TodoistDue{
			Date: task.DueDate.Format(timeDueDateLayout),
		}
	}

	return &result, nil
}
