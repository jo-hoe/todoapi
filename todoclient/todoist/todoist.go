package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	customhttp "github.com/jo-hoe/todoapi/http"
	"github.com/jo-hoe/todoapi/todoclient"
)

type TodoistClient struct {
	httpClient *http.Client
}

type TodoistTask struct {
	ID           string      `json:"id,omitempty"`
	ProjectID    string      `json:"project_id,omitempty"`
	Content      string      `json:"content,omitempty"`
	Description  string      `json:"description,omitempty"`
	CommentCount uint        `json:"comment_count,omitempty"`
	Created      time.Time   `json:"created,omitempty" examples:"2022-10-16T11:53:16.720180Z"`
	Due          *TodoistDue `json:"due,omitempty"`
}

type TodoistComment struct {
	Content string `json:"content,omitempty"`
}

type TodoistProject struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type TodoistDue struct {
	Date string `json:"date,omitempty" examples:"2022-10-16T11:53:16.720180Z"`
}

const (
	todoistUrl         = "https://api.todoist.com/rest/v2/"
	todoistTasksUrl    = todoistUrl + "tasks"
	todoistTaskUrl     = todoistUrl + "tasks/%s"
	todoistParentsUrl  = todoistUrl + "projects"
	todoistCommentsUrl = todoistUrl + "comments?task_id=%s"
	timeDueDateLayout  = "2006-01-02"
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

func (client *TodoistClient) CreateTask(parentId string, task todoclient.ToDoTask) (tasks todoclient.ToDoTask, err error) {
	result := todoclient.ToDoTask{}
	payload, err := convertTodoistTask(task)
	if err != nil {
		return result, err
	}
	payload.ProjectID = parentId

	// create task
	jsonPayload, _ := json.Marshal(payload)
	resp, err := client.httpClient.Post(todoistTasksUrl, "application/json", bytes.NewBuffer(jsonPayload))
	if resp.StatusCode != 200 {
		return result, fmt.Errorf("%+v", resp.Status)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	responseObject := TodoistTask{}
	err = decoder.Decode(&responseObject)
	if err != nil {
		return result, fmt.Errorf("could not decode data :%v", err)
	}

	// convert task
	temp, err := client.convertToToDoTask(responseObject)
	if err != nil {
		return result, err
	}

	return *temp, err
}

func (client *TodoistClient) getComments(taskId string) ([]string, error) {
	comments := make([]TodoistComment, 0)
	err := client.getData(fmt.Sprintf(todoistCommentsUrl, taskId), &comments)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, comment := range comments {
		result = append(result, comment.Content)
	}

	return result, nil
}

func (client *TodoistClient) UpdateTask(parentId string, task todoclient.ToDoTask) error {
	payload, err := convertTodoistTask(task)
	if err != nil {
		return err
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := client.httpClient.Post(fmt.Sprintf(todoistTaskUrl, task.ID), "application/json", bytes.NewBuffer(jsonPayload))
	if resp.StatusCode != 200 {
		return fmt.Errorf("%+v", resp.Status)
	}

	return err
}

func (client *TodoistClient) DeleteTask(parentId string, taskId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf(todoistTaskUrl, taskId), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()

	return err
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
		convertedTask, err := client.convertToToDoTask(task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *convertedTask)
	}

	return tasks, nil
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

func (client *TodoistClient) convertToToDoTask(task TodoistTask) (*todoclient.ToDoTask, error) {
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
		ID:           task.ID,
		Name:         task.Content,
		Description:  task.Description,
		DueDate:      dueDate,
		CreationTime: task.Created,
	}

	if task.CommentCount > 0 {
		comments, err := client.getComments(task.ID)
		if err != nil {
			return &result, nil
		}
		for _, comment := range comments {
			if len(result.Description) > 0 {
				result.Description += "\n"
			}
			result.Description += comment
		}
	}

	return &result, nil
}

// converts but does not convert creation date
func convertTodoistTask(task todoclient.ToDoTask) (*TodoistTask, error) {
	result := TodoistTask{
		ID:          task.ID,
		Content:     task.Name,
		Description: task.Description,
	}

	if !task.CreationTime.IsZero() {
		result.Due = &TodoistDue{
			Date: task.DueDate.Format(timeDueDateLayout),
		}
	}

	return &result, nil
}
