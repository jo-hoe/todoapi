include help.mk

.DEFAULT_GOAL := start

ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: lint-docker
lint-docker: ## Run linters
	@echo "Running linters..."
	@docker run --rm -v $(ROOT_DIR):/app -w /app golangci/golangci-lint:latest golangci-lint run


.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run $(ROOT_DIR)/...

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v -cover ./...