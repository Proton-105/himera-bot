# ==============================================================================
# Variables
# ==============================================================================
BIN_DIR := bin
BIN_NAME := himera-bot
BIN_PATH := $(BIN_DIR)/$(BIN_NAME)

# Go-related variables
GO := go
GO_FLAGS := -v
GO_BUILD_FLAGS := -ldflags="-s -w"
GO_TEST_FLAGS := -race -cover

# Docker-related variables
DOCKER_COMPOSE := docker-compose

.PHONY: all build run test lint clean docker-up docker-down docker-logs migrate-up db-reset mod-tidy help

# ==============================================================================
# Main targets
# ==============================================================================

## build: Compile the Go binary
build: $(BIN_PATH)

$(BIN_PATH):
	@echo "Building Go binary..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build $(GO_FLAGS) $(GO_BUILD_FLAGS) -o $(BIN_PATH) ./cmd/bot

## run: Run the bot directly
run: build
	@echo "Running the bot..."
	@./$(BIN_PATH)

## test: Run Go unit tests with coverage
test:
	@echo "Running tests..."
	@$(GO) test $(GO_TEST_FLAGS) -coverprofile=coverage.out ./...

## lint: Run golangci-lint locally
lint:
	@echo "Running golangci-lint locally..."
	@golangci-lint run ./...

# ==============================================================================
# Utility targets
# ==============================================================================

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR) coverage.out

## mod-tidy: Sync Go module dependencies
mod-tidy:
	@echo "Tidying go modules..."
	@$(GO) mod tidy
	@$(GO) mod vendor

## help: Show this help message
help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ==============================================================================
# Docker & Database targets
# ==============================================================================

## docker-up: Start services with Docker Compose
docker-up:
	@echo "Starting Docker services..."
	@$(DOCKER_COMPOSE) up -d

## docker-down: Stop services
docker-down:
	@echo "Stopping Docker services..."
	@$(DOCKER_COMPOSE) down

## docker-logs: Follow service logs
docker-logs:
	@echo "Following logs..."
	@$(DOCKER_COMPOSE) logs -f

## migrate-up: Apply database migrations
migrate-up:
	@echo "This target should be implemented with a migration tool."

## db-reset: Reset database using script
db-reset:
	@echo "This target should be implemented to reset the database."
