include help.mk

.DEFAULT_GOAL := start

ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: lint-docker
lint-docker: ## Run linters in docker container (no installation required)
	@echo "Running linters..."
	@docker run --rm -v $(ROOT_DIR):/app -w /app golangci/golangci-lint:latest golangci-lint run


.PHONY: lint
lint: ## Run linters using golangci-lint
	@echo "Running linters..."
	@golangci-lint run $(ROOT_DIR)...

.PHONY: test
test: ## Run tests via golang
	@echo "Running tests..."
	@go test -v -cover $(ROOT_DIR)...

.PHONY: update
update: ## Update dependencies
	@git pull
	@go get -u ./...
	@go mod tidy