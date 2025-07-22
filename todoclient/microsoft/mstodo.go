// Package microsoft provides Microsoft To Do API client implementation
package microsoft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jo-hoe/todoapi/internal/common"
	"github.com/jo-hoe/todoapi/pkg/errors"
	"github.com/jo-hoe/todoapi/pkg/logger"
	"github.com/jo-hoe/todoapi/todoclient"
)

const (
	todoAPIURL = "https://graph.microsoft.com/v1.0/me/todo/"
	listsURL   = todoAPIURL + "lists/"
	listURL    = listsURL + "%s/"   // %s = list id
	tasksURL   = listURL + "tasks/" // %s = list id
	taskURL    = tasksURL + "%s"    // %s = list id; %s = task id

	timeDueDateLayout = "2006-01-02T15:04:05.9999999" // this weird MS format is not used consistently in JSON object
	defaultTimeZone   = "Etc/GMT"
)

// Client uses REST MS API
// https://learn.microsoft.com/en-us/graph/api/resources/todo-overview?view=graph-rest-1.0
type MSToDo struct {
	client *http.Client
	logger *logger.Logger
}

type msTask struct {
	ID             string              `json:"id"`
	DisplayName    string              `json:"displayName"`
	BodyItem       bodyItem            `json:"bodyItem"`
	DueDate        time.Time           `json:"dueDateTime"`
	CreationDate   time.Time           `json:"createdDateTime"`
	CheckListItems []msDisplayNameItem `json:"checklistItems"`
	ListID         string
}

type bodyItem struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType" example:"text"`
}

type msOdataLists struct {
	OdataContext  string              `json:"@odata.context"`
	OdataNextlink string              `json:"@odata.nextLink,omitempty"`
	Value         []msDisplayNameItem `json:"value"`
}

type msDisplayNameItem struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"displayName"`
}

type msOdataTasks struct {
	OdataNextlink string        `json:"@odata.nextLink,omitempty"`
	Value         []msOdataTask `json:"value"`
}

type msOdataTask struct {
	DueDateTime      *msOdataDateTime `json:"dueDateTime,omitempty"`
	ID               string           `json:"id,omitempty"`
	Title            string           `json:"title,omitempty"`
	Body             *bodyItem        `json:"body,omitempty"`
	CreationDateTime *time.Time       `json:"createdDateTime,omitempty"`
}

type msOdataDateTime struct {
	DateTime string `json:"dateTime,omitempty" examples:"2020-08-25T04:00:00.0000000"`
	TimeZone string `json:"timeZone,omitempty" examples:"Etc/GMT"`
}

func NewMSToDo(client *http.Client) *MSToDo {
	return &MSToDo{
		client: client,
		logger: logger.New(),
	}
}

// GetAllTasks returns all tasks across all lists
func (msToDo *MSToDo) GetAllTasks(ctx context.Context) ([]todoclient.ToDoTask, error) {
	taskLists, err := msToDo.getTaskLists(ctx)
	if err != nil {
		msToDo.logger.WithError(err).Error("failed to get task lists")
		return nil, errors.NewAPIError("MS_GET_LISTS_FAILED", "failed to retrieve task lists", err)
	}

	result := make([]todoclient.ToDoTask, 0)

	for _, taskList := range taskLists.Value {
		tasksInList, err := msToDo.getChildrenMSTasks(ctx, taskList.ID)
		if err != nil {
			msToDo.logger.WithError(err).WithField("listId", taskList.ID).Error("failed to get tasks for list")
			return nil, errors.NewAPIError("MS_GET_TASKS_FAILED", "failed to retrieve tasks for list", err)
		}
		msTasks := msToDo.processChildren(taskList.ID, tasksInList)
		result = append(result, msTasks...)
	}

	return result, nil
}

func concertToMSToDoTask(input todoclient.ToDoTask) msOdataTask {
	// create result
	result := msOdataTask{
		Title: input.Name,
	}
	if !input.DueDate.IsZero() {
		result.DueDateTime = &msOdataDateTime{
			DateTime: input.DueDate.Format(timeDueDateLayout),
			TimeZone: defaultTimeZone,
		}
	}
	if input.Description != "" {
		result.Body = &bodyItem{
			Content:     input.Description,
			ContentType: "text",
		}
	}

	return result
}

func (msToDo *MSToDo) UpdateTask(ctx context.Context, parentID string, task todoclient.ToDoTask) error {
	if err := task.Validate(); err != nil {
		return err
	}

	payload := concertToMSToDoTask(task)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.NewAPIError("MS_MARSHAL_FAILED", "failed to marshal task", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf(taskURL, parentID, task.ID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return errors.NewAPIError("MS_REQUEST_FAILED", "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := msToDo.client.Do(req)
	if err != nil {
		return errors.NewAPIError("MS_HTTP_FAILED", "HTTP request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.NewAPIError("MS_UPDATE_FAILED", fmt.Sprintf("update failed with status %d", resp.StatusCode), fmt.Errorf("%s", string(body)))
	}

	return nil
}

func (msToDo *MSToDo) CreateTask(ctx context.Context, parentID string, task todoclient.ToDoTask) (todoclient.ToDoTask, error) {
	var result todoclient.ToDoTask

	if err := task.Validate(); err != nil {
		return result, err
	}

	payload := concertToMSToDoTask(task)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return result, errors.NewAPIError("MS_MARSHAL_FAILED", "failed to marshal task", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(tasksURL, parentID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return result, errors.NewAPIError("MS_REQUEST_FAILED", "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := msToDo.client.Do(req)
	if err != nil {
		return result, errors.NewAPIError("MS_HTTP_FAILED", "HTTP request failed", err)
	}

	var data msOdataTask
	if err := decodeJSONResponse(resp, http.StatusCreated, &data); err != nil {
		return result, errors.NewAPIError("MS_DECODE_FAILED", "failed to decode response", err)
	}

	result.Name = data.Title
	result.ID = data.ID
	result.Description = ""
	if data.Body != nil {
		result.Description = data.Body.Content
	}

	if data.CreationDateTime != nil {
		result.CreationTime = *data.CreationDateTime
	}

	if data.DueDateTime != nil {
		if dueDate, err := time.Parse(timeDueDateLayout, data.DueDateTime.DateTime); err == nil {
			result.DueDate = dueDate
		}
	}

	return result, nil
}

func (msToDo *MSToDo) DeleteTask(ctx context.Context, parentID, taskID string) error {
	return msToDo.deleteObject(ctx, fmt.Sprintf(taskURL, parentID, taskID))
}

func (msToDo *MSToDo) deleteObject(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return errors.NewAPIError("MS_REQUEST_FAILED", "failed to create delete request", err)
	}

	resp, err := msToDo.client.Do(req)
	if err != nil {
		return errors.NewAPIError("MS_HTTP_FAILED", "HTTP delete request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return errors.NewAPIError("MS_DELETE_FAILED", fmt.Sprintf("delete failed with status %d", resp.StatusCode), fmt.Errorf("%s", string(body)))
	}

	return nil
}

func (msToDo *MSToDo) CreateParent(ctx context.Context, parentName string) (todoclient.ToDoParent, error) {
	var result todoclient.ToDoParent

	parentName = strings.TrimSpace(parentName)
	if parentName == "" {
		return result, errors.NewValidationError("name", "parent name cannot be empty")
	}

	payload := msDisplayNameItem{
		DisplayName: parentName,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return result, errors.NewAPIError("MS_MARSHAL_FAILED", "failed to marshal parent", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, listsURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return result, errors.NewAPIError("MS_REQUEST_FAILED", "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := msToDo.client.Do(req)
	if err != nil {
		return result, errors.NewAPIError("MS_HTTP_FAILED", "HTTP request failed", err)
	}

	var data msDisplayNameItem
	if err := decodeJSONResponse(resp, http.StatusCreated, &data); err != nil {
		return result, errors.NewAPIError("MS_DECODE_FAILED", "failed to decode response", err)
	}

	result.ID = data.ID
	result.Name = data.DisplayName

	return result, nil
}

func (msToDo *MSToDo) DeleteParent(ctx context.Context, parentID string) error {
	return msToDo.deleteObject(ctx, fmt.Sprintf(listURL, parentID))
}

func (msToDo *MSToDo) GetAllParents(ctx context.Context) ([]todoclient.ToDoParent, error) {
	result := make([]todoclient.ToDoParent, 0)

	lists, err := msToDo.getTaskLists(ctx)
	if err != nil {
		msToDo.logger.WithError(err).Error("failed to get task lists")
		return result, errors.NewAPIError("MS_GET_LISTS_FAILED", "failed to retrieve task lists", err)
	}

	for _, list := range lists.Value {
		result = append(result, todoclient.ToDoParent{
			ID:   list.ID,
			Name: list.DisplayName,
		})
	}

	return result, nil
}

func (msToDo *MSToDo) GetChildrenTasks(ctx context.Context, parentID string) ([]todoclient.ToDoTask, error) {
	childrenTasks, err := msToDo.getChildrenMSTasks(ctx, parentID)
	if err != nil {
		msToDo.logger.WithError(err).WithField("parentId", parentID).Error("failed to get children tasks")
		return nil, errors.NewAPIError("MS_GET_CHILDREN_FAILED", "failed to retrieve children tasks", err)
	}
	return msToDo.processChildren(parentID, childrenTasks), nil
}

// Converts items to OData items to generic ToDoTasks and updates the internal cache
func (msToDo *MSToDo) processChildren(listId string, tasksInList []msTask) (tasks []todoclient.ToDoTask) {
	result := make([]todoclient.ToDoTask, 0)

	for _, task := range tasksInList {
		task.ListID = listId

		result = append(result, todoclient.ToDoTask{
			ID:           task.ID,
			Name:         task.DisplayName,
			Description:  task.BodyItem.Content,
			DueDate:      task.DueDate,
			CreationTime: task.CreationDate,
		})
	}

	return result
}

func (msToDo *MSToDo) getChildrenMSTasks(ctx context.Context, parentID string) ([]msTask, error) {
	result := []msTask{}
	url := fmt.Sprintf(tasksURL, parentID)

	for url != "" {
		tasks := msOdataTasks{}
		if err := msToDo.getData(ctx, url, &tasks); err != nil {
			return nil, err
		}

		for _, task := range tasks.Value {
			dueDate := time.Time{}
			if task.DueDateTime != nil {
				if deserializedTime, err := time.Parse(timeDueDateLayout, task.DueDateTime.DateTime); err == nil {
					dueDate = deserializedTime
				}
			}

			item := msTask{
				ID:          task.ID,
				DisplayName: task.Title,
				DueDate:     dueDate,
				ListID:      parentID,
			}

			if task.CreationDateTime != nil {
				item.CreationDate = *task.CreationDateTime
			}

			if task.Body != nil && task.Body.Content != "" {
				item.BodyItem.Content = task.Body.Content
				item.BodyItem.ContentType = task.Body.ContentType
			}

			result = append(result, item)
		}
		url = tasks.OdataNextlink
	}
	return result, nil
}

func (msToDo *MSToDo) getTaskLists(ctx context.Context) (*msOdataLists, error) {
	lists := msOdataLists{}
	url := listsURL
	for url != "" {
		tmpList := msOdataLists{}
		if err := msToDo.getData(ctx, url, &tmpList); err != nil {
			return nil, err
		}
		lists.Value = append(lists.Value, tmpList.Value...)
		url = tmpList.OdataNextlink
	}
	return &lists, nil
}

func (msToDo *MSToDo) getData(ctx context.Context, url string, data interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.NewAPIError("MS_REQUEST_FAILED", "failed to create GET request", err)
	}

	resp, err := msToDo.client.Do(req)
	if err != nil {
		return errors.NewAPIError("MS_HTTP_FAILED", "HTTP GET request failed", err)
	}
	defer common.CloseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.NewAPIError("MS_GET_FAILED", fmt.Sprintf("GET failed with status %d", resp.StatusCode), fmt.Errorf("%s", string(body)))
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(data); err != nil {
		return errors.NewAPIError("MS_DECODE_FAILED", "failed to decode response data", err)
	}

	return nil
}

// decodeJSONResponse checks the response status, decodes JSON, and closes the body.
func decodeJSONResponse(resp *http.Response, expectedStatus int, out interface{}) error {
	defer common.CloseBody(resp.Body)
	if resp.StatusCode != expectedStatus {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("received error: could not read body: %v", err)
		}
		return fmt.Errorf("received error: %+v", string(b))
	}
	if out != nil {
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(out); err != nil {
			return fmt.Errorf("could not decode data :%v", err)
		}
	}
	return nil
}
