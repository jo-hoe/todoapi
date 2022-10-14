package microsoft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jo-hoe/todoapi/todoclient"
)

const listURL = "https://graph.microsoft.com/v1.0/me/todo/lists/"

type MSToDo struct {
	client    *http.Client
	taskCache map[string]msTask
}

type msTask struct {
	ID           string    `json:"id"`
	DisplayName  string    `json:"displayName"`
	DueDate      time.Time `json:"dueDateTime"`
	CreationDate time.Time `json:"createdDateTime"`
	ListID       string
}

type msOdataLists struct {
	OdataContext  string               `json:"@odata.context"`
	OdataNextlink string               `json:"@odata.nextLink,omitempty"`
	Value         []msOdataListDetails `json:"value"`
}

type msOdataListDetails struct {
	ID          string `json:"id"`
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
	CreationDateTime *time.Time       `json:"createdDateTime,omitempty"`
}

type msOdataDateTime struct {
	DateTime string `json:"dateTime,omitempty" examples:"2020-08-25T04:00:00.0000000"`
}

func NewMSToDo(client *http.Client) *MSToDo {
	mstodo := &MSToDo{
		client:    client,
		taskCache: make(map[string]msTask),
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
	msToDo.taskCache = make(map[string]msTask)

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
			Title:        task.DisplayName,
			DueDate:      task.DueDate,
			CreationTime: task.CreationDate,
		})

		msToDo.taskCache[task.ID] = task
	}

	return result
}

func (msToDo *MSToDo) UpdateTask(task todoclient.ToDoTask) error {
	// check if items id is known in cache and the list id can therefore be retrieved
	if _, ok := msToDo.taskCache[task.ID]; !ok {
		_, err := msToDo.GetAllTasks()
		if err != nil {
			return err
		}
	}
	msTask := msToDo.taskCache[task.ID]
	listId := msTask.ListID

	payload := msOdataTask{
		Title: task.Title,
	}
	if !task.DueDate.IsZero() {
		msTask.DueDate = task.DueDate
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, listURL+listId+"/tasks/"+task.ID, bytes.NewBuffer(jsonPayload))
	req.Header.Add("Content-Type", "application/json")
	resp, err := msToDo.client.Do(req)

	if resp.StatusCode != 200 {
		return fmt.Errorf("%+v", resp.Status)
	}

	return err
}

func (client *MSToDo) CreateTask(task todoclient.ToDoTask) (tasks todoclient.ToDoTask, err error) {
	return todoclient.ToDoTask{}, nil
}

func (msToDo *MSToDo) DeleteTask(task todoclient.ToDoTask) error {
	return nil
}

func (msToDo *MSToDo) GetAllParents() (todoclient.ToDoParent, error) {
	return todoclient.ToDoParent{}, nil
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
	timeDueDateLayout := "2006-01-02T15:04:05.9999999" // this weird MS format is not used consistently in JSON object
	url := listURL + parentId + "/tasks"
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
			result = append(result, msTask{
				ID:           task.ID,
				DisplayName:  task.Title,
				DueDate:      dueDate,
				CreationDate: *task.CreationDateTime,
				ListID:       parentId,
			})
		}
		url = tasks.OdataNextlink
	}
	return result, nil
}

func (msToDo *MSToDo) getTaskLists() (*msOdataLists, error) {
	lists := msOdataLists{}
	url := listURL
	for url != "" {
		tmpList := msOdataLists{}
		err := msToDo.getData(listURL, &tmpList)
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
