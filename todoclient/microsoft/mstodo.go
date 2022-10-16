package microsoft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
}

type msTask struct {
	ID           string    `json:"id"`
	DisplayName  string    `json:"displayName"`
	BodyItem     bodyItem  `json:"bodyItem"`
	DueDate      time.Time `json:"dueDateTime"`
	CreationDate time.Time `json:"createdDateTime"`
	ListID       string
}

type bodyItem struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType" example:"text"`
}

type msOdataLists struct {
	OdataContext  string               `json:"@odata.context"`
	OdataNextlink string               `json:"@odata.nextLink,omitempty"`
	Value         []msOdataListDetails `json:"value"`
}

type msOdataListDetails struct {
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
	mstodo := &MSToDo{
		client: client,
	}
	return mstodo
}

// GetAllTask returns all tasks.
// Items will be kepted in a volatile RAM
func (msToDo *MSToDo) GetAllTasks() (tasks []todoclient.ToDoTask, err error) {
	taskLists, err := msToDo.getTaskLists()
	if err != nil {
		return nil, err
	}

	result := make([]todoclient.ToDoTask, 0)

	for _, taskList := range taskLists.Value {
		tasksInList, err := msToDo.getChildrenMSTasks(taskList.ID)
		if err != nil {
			return nil, err
		}
		msTasks := msToDo.processChildren(taskList.ID, tasksInList)
		result = append(result, msTasks...)
	}

	return result, nil
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

func (msToDo *MSToDo) UpdateTask(parentId string, task todoclient.ToDoTask) error {
	payload := concertToMSToDoTask(task)

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf(taskURL, parentId, task.ID), bytes.NewBuffer(jsonPayload))
	req.Header.Add("Content-Type", "application/json")
	resp, err := msToDo.client.Do(req)

	if resp.StatusCode != 200 {
		return fmt.Errorf("%+v", resp.Status)
	}

	return err
}

func (msToDo *MSToDo) CreateTask(parentId string, task todoclient.ToDoTask) (result todoclient.ToDoTask, err error) {
	payload := concertToMSToDoTask(task)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}

	// send request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(tasksURL, parentId), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return result, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := msToDo.client.Do(req)
	if resp.StatusCode != 201 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, err
		}
		return result, fmt.Errorf("received error: %+v", string(b))
	}

	// decode request
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	data := msOdataTask{}
	err = decoder.Decode(&data)
	if err != nil {
		return result, fmt.Errorf("could not decode data :%v", err)
	}
	result.DueDate, err = time.Parse(timeDueDateLayout, data.DueDateTime.DateTime)
	if err != nil {
		return result, fmt.Errorf("could not decode data :%v", err)
	}
	result.Name = data.Title
	result.ID = data.ID
	result.CreationTime = *data.CreationDateTime

	return result, err
}

func (msToDo *MSToDo) DeleteTask(parentId string, taskId string) error {
	return msToDo.deleteObject(fmt.Sprintf(taskURL, parentId, taskId))
}

func (msToDo *MSToDo) deleteObject(url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := msToDo.client.Do(req)
	if resp.StatusCode != 204 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("received error: %+v", string(b))
	}

	return err
}

func (msToDo *MSToDo) CreateParent(parentName string) (todoclient.ToDoParent, error) {
	result := todoclient.ToDoParent{}
	payload := msOdataListDetails{
		DisplayName: parentName,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}

	// send request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(listsURL), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return result, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := msToDo.client.Do(req)
	if resp.StatusCode != 201 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, err
		}
		return result, fmt.Errorf("received error: %+v", string(b))
	}

	// decode request
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	data := msOdataListDetails{}
	err = decoder.Decode(&data)
	if err != nil {
		return result, fmt.Errorf("could not decode data :%v", err)
	}

	result.ID = data.ID
	result.Name = data.DisplayName

	return result, err
}

func (client *MSToDo) DeleteParent(parentId string) error {
	return client.deleteObject(fmt.Sprintf(listURL, parentId))
}

func (msToDo *MSToDo) GetAllParents() ([]todoclient.ToDoParent, error) {
	result := make([]todoclient.ToDoParent, 0)

	lists, err := msToDo.getTaskLists()
	if err != nil {
		return result, err
	}

	for _, list := range lists.Value {
		result = append(result, todoclient.ToDoParent{
			ID:   list.ID,
			Name: list.DisplayName,
		})
	}

	return result, err
}

func (msToDo *MSToDo) GetChildrenTasks(parentId string) (tasks []todoclient.ToDoTask, err error) {
	childerenTasks, err := msToDo.getChildrenMSTasks(parentId)
	if err != nil {
		return nil, err
	}
	return msToDo.processChildren(parentId, childerenTasks), nil
}

func (msToDo *MSToDo) getChildrenMSTasks(parentId string) ([]msTask, error) {
	result := []msTask{}
	url := fmt.Sprintf(tasksURL, parentId)
	for url != "" {
		tasks := msOdataTasks{}
		err := msToDo.getData(url, &tasks)
		if err != nil {
			return nil, err
		}

		for _, task := range tasks.Value {
			dueDate := time.Time{}
			if task.DueDateTime != nil {
				deserializedTime, err := time.Parse(timeDueDateLayout, task.DueDateTime.DateTime)
				if err == nil {
					dueDate = deserializedTime
				} else {
					dueDate = time.Time{}
				}
			}

			item := msTask{
				ID:           task.ID,
				DisplayName:  task.Title,
				DueDate:      dueDate,
				CreationDate: *task.CreationDateTime,
				ListID:       parentId,
			}
			if task.Body.Content != "" {
				item.BodyItem.Content = task.Body.Content
				item.BodyItem.ContentType = task.Body.ContentType
			}

			result = append(result, item)
		}
		url = tasks.OdataNextlink
	}
	return result, nil
}

func (msToDo *MSToDo) getTaskLists() (*msOdataLists, error) {
	lists := msOdataLists{}
	url := listsURL
	for url != "" {
		tmpList := msOdataLists{}
		err := msToDo.getData(url, &tmpList)
		if err != nil {
			return nil, err
		}
		lists.Value = append(lists.Value, tmpList.Value...)
		url = lists.OdataNextlink
	}
	return &lists, nil
}

func (msToDo *MSToDo) getData(url string, data interface{}) error {
	resp, err := msToDo.client.Get(url)
	if err != nil {
		return fmt.Errorf("error after GET call:%s", err.Error())
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return fmt.Errorf("could not decode data :%v", err)
	}

	return nil
}
