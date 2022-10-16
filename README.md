# TodoAPI

[![Test Status](https://github.com/jo-hoe/todoapi/workflows/test/badge.svg)](https://github.com/jo-hoe/todoapi/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/todoapi/workflows/lint/badge.svg)](https://github.com/jo-hoe/todoapi/actions?workflow=lint)

API for todo applications.

[x] Todoist
[x] MS To Do

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
