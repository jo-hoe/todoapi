package microsoft

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jo-hoe/todoapi/todoclient/testutil"
)

func TestMSToDo_Integration_GetAllTasks(t *testing.T) {
	testutil.IntegrationTest_GetAllTasks(t, createClient(t))
}

func TestMSToDo_Integration_CRUD(t *testing.T) {
	testutil.IntegrationTest_CRUD(t, createClient(t))
}

func createClient(t *testing.T) *MSToDo {
	clientCredentials := os.Getenv("MSCLIENTCREDENTIALS")
	if clientCredentials == "" {
		t.Skip("Test will be skipped without client credentials")
	}

	token := os.Getenv("MSTOKEN")
	if token == "" {
		t.Skip("Test will be skipped without token")
	}

	msClientConfig := MSClientConfig{}
	err := json.Unmarshal([]byte(token), &msClientConfig.Token)
	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}

	err = json.Unmarshal([]byte(clientCredentials), &msClientConfig.ClientCredentials)
	if err != nil {
		t.Errorf("error was not nil but '%v'", err)
	}

	httpClient := NewClient(msClientConfig, nil)
	return NewMSToDo(httpClient)
}
