// Package todoist provides Todoist API client implementation
package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jo-hoe/todoapi/internal/common"
	customhttp "github.com/jo-hoe/todoapi/internal/http"
	"github.com/jo-hoe/todoapi/pkg/errors"
	"github.com/jo-hoe/todoapi/pkg/logger"
	"github.com/jo-hoe/todoapi/todoclient"
)

type TodoistClient struct {
	httpClient *http.Client
	logger     *logger.Logger
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
	todoistParentUrl   = todoistParentsUrl + "/%s"
	todoistCommentsUrl = todoistUrl + "comments?task_id=%s"
	timeDueDateLayout  = "2006-01-02"
)

// NewTodoistHTTPClient creates an HTTP client with injected REST API token for each request
func NewTodoistHTTPClient(token string) *http.Client {
	return customhttp.NewHTTPClientWithHeader("Authorization", "Bearer "+token)
}

func NewTodoistClient(httpClient *http.Client) *TodoistClient {
	return &TodoistClient{
		httpClient: httpClient,
		logger:     logger.New(),
	}
}

func (client *TodoistClient) CreateTask(ctx context.Context, parentID string, task todoclient.ToDoTask) (todoclient.ToDoTask, error) {
	var result todoclient.ToDoTask

	if err := task.Validate(); err != nil {
		return result, err
	}

	payload, err := convertTodoistTask(task)
	if err != nil {
		return result, errors.NewAPIError("TODOIST_CONVERT_FAILED", "failed to convert task", err)
	}
	payload.ProjectID = parentID

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return result, errors.NewAPIError("TODOIST_MARSHAL_FAILED", "failed to marshal task", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, todoistTasksUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return result, errors.NewAPIError("TODOIST_REQUEST_FAILED", "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return result, errors.NewAPIError("TODOIST_HTTP_FAILED", "HTTP request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return result, errors.NewAPIError("TODOIST_CREATE_FAILED", fmt.Sprintf("create failed with status %d", resp.StatusCode), nil)
	}

	var responseObject TodoistTask
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&responseObject); err != nil {
		return result, errors.NewAPIError("TODOIST_DECODE_FAILED", "failed to decode response", err)
	}

	convertedTask, err := client.convertToToDoTask(ctx, responseObject)
	if err != nil {
		return result, err
	}

	return *convertedTask, nil
}

func (client *TodoistClient) getComments(ctx context.Context, taskID string) ([]string, error) {
	var comments []TodoistComment
	if err := client.getData(ctx, fmt.Sprintf(todoistCommentsUrl, taskID), &comments); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(comments))
	for _, comment := range comments {
		result = append(result, comment.Content)
	}

	return result, nil
}

func (client *TodoistClient) UpdateTask(ctx context.Context, parentID string, task todoclient.ToDoTask) error {
	if err := task.Validate(); err != nil {
		return err
	}

	payload, err := convertTodoistTask(task)
	if err != nil {
		return errors.NewAPIError("TODOIST_CONVERT_FAILED", "failed to convert task", err)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.NewAPIError("TODOIST_MARSHAL_FAILED", "failed to marshal task", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(todoistTaskUrl, task.ID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return errors.NewAPIError("TODOIST_REQUEST_FAILED", "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return errors.NewAPIError("TODOIST_HTTP_FAILED", "HTTP request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.NewAPIError("TODOIST_UPDATE_FAILED", fmt.Sprintf("update failed with status %d", resp.StatusCode), nil)
	}

	return nil
}

func (client *TodoistClient) DeleteTask(ctx context.Context, parentID, taskID string) error {
	return client.deleteObject(ctx, fmt.Sprintf(todoistTaskUrl, taskID))
}

func (client *TodoistClient) deleteObject(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return errors.NewAPIError("TODOIST_REQUEST_FAILED", "failed to create delete request", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return errors.NewAPIError("TODOIST_HTTP_FAILED", "HTTP delete request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return errors.NewAPIError("TODOIST_DELETE_FAILED", fmt.Sprintf("delete failed with status %d", resp.StatusCode), nil)
	}

	return nil
}

func (client *TodoistClient) CreateParent(ctx context.Context, parentName string) (todoclient.ToDoParent, error) {
	var result todoclient.ToDoParent

	parentName = strings.TrimSpace(parentName)
	if parentName == "" {
		return result, errors.NewValidationError("name", "parent name cannot be empty")
	}

	payload := TodoistProject{
		Name: parentName,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return result, errors.NewAPIError("TODOIST_MARSHAL_FAILED", "failed to marshal parent", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, todoistParentsUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return result, errors.NewAPIError("TODOIST_REQUEST_FAILED", "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return result, errors.NewAPIError("TODOIST_HTTP_FAILED", "HTTP request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return result, errors.NewAPIError("TODOIST_CREATE_PARENT_FAILED", fmt.Sprintf("create parent failed with status %d", resp.StatusCode), nil)
	}

	var responseObject TodoistProject
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&responseObject); err != nil {
		return result, errors.NewAPIError("TODOIST_DECODE_FAILED", "failed to decode response", err)
	}

	result.ID = responseObject.ID
	result.Name = responseObject.Name

	return result, nil
}

func (client *TodoistClient) DeleteParent(ctx context.Context, parentID string) error {
	return client.deleteObject(ctx, fmt.Sprintf(todoistParentUrl, parentID))
}

func (client *TodoistClient) GetAllParents(ctx context.Context) ([]todoclient.ToDoParent, error) {
	result := make([]todoclient.ToDoParent, 0)
	var projects []TodoistProject

	if err := client.getData(ctx, todoistParentsUrl, &projects); err != nil {
		client.logger.WithError(err).Error("failed to get all parents")
		return result, errors.NewAPIError("TODOIST_GET_PARENTS_FAILED", "failed to retrieve parents", err)
	}

	for _, project := range projects {
		parent := todoclient.ToDoParent{
			ID:   project.ID,
			Name: project.Name,
		}
		result = append(result, parent)
	}

	return result, nil
}

func (client *TodoistClient) GetAllTasks(ctx context.Context) ([]todoclient.ToDoTask, error) {
	return client.getTasks(ctx, nil)
}

func (client *TodoistClient) GetChildrenTasks(ctx context.Context, parentID string) ([]todoclient.ToDoTask, error) {
	return client.getTasks(ctx, &parentID)
}

func (client *TodoistClient) getTasks(ctx context.Context, parentID *string) ([]todoclient.ToDoTask, error) {
	var todoistTasks []TodoistTask

	url := todoistTasksUrl
	if parentID != nil {
		url = url + "?project_id=" + *parentID
	}

	if err := client.getData(ctx, url, &todoistTasks); err != nil {
		client.logger.WithError(err).Error("failed to get tasks")
		return nil, errors.NewAPIError("TODOIST_GET_TASKS_FAILED", "failed to retrieve tasks", err)
	}

	tasks := make([]todoclient.ToDoTask, 0, len(todoistTasks))
	for _, task := range todoistTasks {
		convertedTask, err := client.convertToToDoTask(ctx, task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *convertedTask)
	}

	return tasks, nil
}

func (client *TodoistClient) getData(ctx context.Context, url string, data interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.NewAPIError("TODOIST_REQUEST_FAILED", "failed to create GET request", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return errors.NewAPIError("TODOIST_HTTP_FAILED", "HTTP GET request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.NewAPIError("TODOIST_GET_FAILED", fmt.Sprintf("GET failed with status %d", resp.StatusCode), nil)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(data); err != nil {
		return errors.NewAPIError("TODOIST_DECODE_FAILED", "failed to decode response data", err)
	}

	return nil
}

func (client *TodoistClient) convertToToDoTask(ctx context.Context, task TodoistTask) (*todoclient.ToDoTask, error) {
	dueDate := time.Time{}
	if task.Due != nil {
		if deserializedTime, err := time.Parse(timeDueDateLayout, task.Due.Date); err == nil {
			dueDate = deserializedTime
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
		comments, err := client.getComments(ctx, task.ID)
		if err != nil {
			// Log error but don't fail the conversion
			client.logger.WithError(err).WithField("taskId", task.ID).Warn("failed to get comments for task")
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

// convertTodoistTask converts a ToDoTask to TodoistTask format
func convertTodoistTask(task todoclient.ToDoTask) (*TodoistTask, error) {
	result := TodoistTask{
		ID:          task.ID,
		Content:     task.Name,
		Description: task.Description,
	}

	if !task.DueDate.IsZero() {
		result.Due = &TodoistDue{
			Date: task.DueDate.Format(timeDueDateLayout),
		}
	}

	return &result, nil
}
