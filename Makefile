.PHONY: build run test clean install deps

# Variables
APP_NAME=financli
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
all: build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the application
build: deps
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/main.go

# Run the application
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Run demo without TUI
demo:
	go run cmd/demo/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install the application globally
install: build
	cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

# Development: run with auto-reload (requires entr)
dev:
	find . -name "*.go" | entr -r make run

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Vet the code
vet:
	go vet ./...

# Check for security issues
security:
	gosec ./...

# Run all checks
check: fmt vet lint test

# Docker build
docker-build:
	docker build -t $(APP_NAME) .

# Docker run
docker-run:
	docker run -it --rm \
		-e MONGODB_URI=mongodb://host.docker.internal:27017 \
		-e MONGODB_DATABASE=financli \
		$(APP_NAME)

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application (requires TTY)"
	@echo "  demo          - Run demonstration without TUI"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install application globally"
	@echo "  dev           - Run with auto-reload (requires entr)"
	@echo "  lint          - Lint the code"
	@echo "  fmt           - Format the code"
	@echo "  vet           - Vet the code"
	@echo "  security      - Check for security issues"
	@echo "  check         - Run all checks (fmt, vet, lint, test)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  help          - Show this help message"