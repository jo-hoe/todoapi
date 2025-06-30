# TodoAPI

[![Test Status](https://github.com/jo-hoe/todoapi/workflows/test/badge.svg)](https://github.com/jo-hoe/todoapi/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/todoapi/workflows/lint/badge.svg)](https://github.com/jo-hoe/todoapi/actions?workflow=lint)

API for todo applications.

- [x] Todoist
- [x] MS To Do

## Dev Setup

This project requires API tokens/secrets for integration tests and running against real services. You can provide these via environment variables or `.env` files as described below.

### Todoist

- **Token:** Obtain your Todoist API token from your [Todoist Integrations Settings](https://todoist.com/prefs/integrations).
- **Usage:**
  - Set the environment variable `TODOIST_API_TOKEN` with your token value.
  - Or, create a `.env` file in `todoclient/todoist/` with:

    ```env
    TODOIST_API_TOKEN=your_token_here
    ```

### Microsoft To Do

- **Client Credentials & Token:**
  - Register an application in the [Azure Portal](https://portal.azure.com/) to obtain your `clientId` and `clientSecret`.
  - Obtain an OAuth token for Microsoft Graph with the required scopes (`openid offline_access tasks.readwrite`).
- **Usage:**
  - Set the environment variables:
    - `MSCLIENTCREDENTIALS` (JSON: `{ "clientId": "...", "clientSecret": "..." }`)
    - `MSTOKEN` (JSON: `{ "token_type": "Bearer", ... }`)
  - Or, create a `.env` file in `todoclient/microsoft/` with:

    ```env
    MSCLIENTCREDENTIALS={"clientId": "your_client_id", "clientSecret": "your_client_secret"}
    MSTOKEN={"token_type": "Bearer", ...}
    ```

## Makefile Usage

The project provides a `Makefile` to simplify common development tasks. You can run `make help` (if your shell supports it) to see a summary of available targets and their descriptions.

## Linting

Project used `golangci-lint` for linting.

### Installation

<https://golangci-lint.run/usage/install/>

### Execution

Run the linting locally by executing

```cli
golangci-lint run ./...
```

in the working directory

## Testing

The project contains both unit and integrations tests.

### Unit Test Execution

The unit test can be excuted using the default golang commands. To run all test execute the following in the parent folder of the repository.

```powershell
go test ./...
```
