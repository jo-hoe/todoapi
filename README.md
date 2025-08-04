# TodoAPI

[![Test](https://github.com/jo-hoe/todoapi/workflows/test/badge.svg)](https://github.com/jo-hoe/todoapi/actions/workflows/test.yml)
[![Lint](https://github.com/jo-hoe/todoapi/workflows/lint/badge.svg)](https://github.com/jo-hoe/todoapi/actions/workflows/lint.yml)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/todoapi/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/todoapi?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/todoapi)](https://goreportcard.com/report/github.com/jo-hoe/todoapi)

A unified API for todo applications with support for multiple providers including Todoist and Microsoft To Do.

## Features

- ✅ **Multi-provider support**: Todoist and Microsoft To Do
- ✅ **Unified interface**: Consistent API across different todo services

## Quick Start

### Prerequisites

- Go 1.24 or later
- Docker (optional)
- API tokens for the services you want to use

### Configuration

The application uses environment variables for configuration:

#### Server Configuration

- `PORT`: Server port (default: 8080)
- `READ_TIMEOUT`: HTTP read timeout (default: 30s)
- `WRITE_TIMEOUT`: HTTP write timeout (default: 30s)
- `IDLE_TIMEOUT`: HTTP idle timeout (default: 60s)

#### Todoist Configuration

- `TODOIST_API_TOKEN`: Your Todoist API token
- `TODOIST_BASE_URL`: Todoist API base URL (default: <https://api.todoist.com/rest/v2/>)

#### Microsoft To Do Configuration

- `MS_CLIENT_ID`: Microsoft application client ID
- `MS_CLIENT_SECRET`: Microsoft application client secret
- `MS_TENANT_ID`: Microsoft tenant ID
- `MS_BASE_URL`: Microsoft Graph API base URL (default: <https://graph.microsoft.com/v1.0/me/todo/>)

#### Logging Configuration

- `LOG_LEVEL`: Logging level (default: info)
- `LOG_FORMAT`: Log format - json or text (default: text)
- `ENV`: Environment - production for JSON logging (default: development)

## API Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/jo-hoe/todoapi/todoclient"
    "github.com/jo-hoe/todoapi/todoclient/todoist"
    customhttp "github.com/jo-hoe/todoapi/internal/http"
)

func main() {
    // Create HTTP client with authentication
    httpClient := todoist.NewTodoistHTTPClient("your-api-token")
    
    // Create Todoist client
    client := todoist.NewTodoistClient(httpClient)
    
    ctx := context.Background()
    
    // Get all tasks
    tasks, err := client.GetAllTasks(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d tasks\n", len(tasks))
    
    // Create a new task
    newTask := todoclient.ToDoTask{
        Name:        "Learn Go",
        Description: "Study Go programming language",
    }
    
    // Get first project/parent
    parents, err := client.GetAllParents(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    if len(parents) > 0 {
        createdTask, err := client.CreateTask(ctx, parents[0].ID, newTask)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Created task: %s\n", createdTask.Name)
    }
}
```

### Microsoft To Do Example

```go
package main

import (
    "context"
    "log"
    
    "github.com/jo-hoe/todoapi/todoclient/microsoft"
    "golang.org/x/oauth2"
)

func main() {
    // Configure OAuth2
    config := &oauth2.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
            TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
        },
        Scopes: []string{"https://graph.microsoft.com/Tasks.ReadWrite"},
    }
    
    // Get token (implement OAuth2 flow)
    token := &oauth2.Token{AccessToken: "your-access-token"}
    httpClient := config.Client(context.Background(), token)
    
    // Create Microsoft To Do client
    client := microsoft.NewMSToDo(httpClient)
    
    ctx := context.Background()
    
    // Get all tasks
    tasks, err := client.GetAllTasks(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d tasks", len(tasks))
}
```

### API Credentials Setup

#### Todoist

1. Go to [Todoist Integrations Settings](https://todoist.com/prefs/integrations)
2. Create a new app or use an existing one
3. Copy the API token
4. Set the environment variable: `export TODOIST_API_TOKEN=your_token_here`

#### Microsoft To Do

1. Register an application in the [Azure Portal](https://portal.azure.com/)
2. Configure the required permissions for Microsoft Graph API
3. Obtain client credentials and implement OAuth2 flow
4. Set the environment variables:

   ```bash
   export MS_CLIENT_ID=your_client_id
   export MS_CLIENT_SECRET=your_client_secret
   export MS_TENANT_ID=your_tenant_id
   ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
