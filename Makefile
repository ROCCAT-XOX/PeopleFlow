# PeopleFlow Makefile

.PHONY: help test test-all test-unit test-integration test-coverage test-models test-handlers test-repos clean build run docker-build docker-run

# Default target
help:
	@echo "PeopleFlow Development Commands"
	@echo "=============================="
	@echo "make test           - Run all tests"
	@echo "make test-all       - Run comprehensive test suite with detailed output"
	@echo "make test-unit      - Run unit tests only"
	@echo "make test-integration - Run integration tests"
	@echo "make test-coverage  - Run tests with coverage report"
	@echo "make test-models    - Run model tests only"
	@echo "make test-handlers  - Run handler tests only"
	@echo "make test-repos     - Run repository tests only"
	@echo "make build          - Build the application"
	@echo "make run            - Run the application"
	@echo "make docker-build   - Build Docker image"
	@echo "make docker-run     - Run with Docker Compose"
	@echo "make clean          - Clean build artifacts"

# Run all tests
test:
	@echo "🧪 Running all tests..."
	@go test -v ./...

# Run comprehensive test suite
test-all:
	@echo "🚀 Running comprehensive test suite..."
	@go run run_all_tests.go

# Run unit tests only (exclude integration tests)
test-unit:
	@echo "🧪 Running unit tests..."
	@go test -v -short ./...

# Run integration tests only
test-integration:
	@echo "🧪 Running integration tests..."
	@go test -v -run Integration ./...

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

# Run model tests
test-models:
	@echo "🧪 Running model tests..."
	@go test -v -cover ./backend/model/...

# Run handler tests
test-handlers:
	@echo "🧪 Running handler tests..."
	@go test -v -cover ./backend/handler/...

# Run repository tests
test-repos:
	@echo "🧪 Running repository tests..."
	@go test -v -cover ./backend/repository/...

# Build the application
build:
	@echo "🔨 Building PeopleFlow..."
	@go build -o peopleflow main.go
	@echo "✅ Build complete: ./peopleflow"

# Run the application
run:
	@echo "🚀 Starting PeopleFlow..."
	@go run main.go

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t peopleflow:latest .

# Run with Docker Compose
docker-run:
	@echo "🐳 Starting with Docker Compose..."
	@docker-compose up -d

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f peopleflow
	@rm -f coverage.out coverage.html
	@rm -rf tmp/
	@echo "✅ Clean complete"

# Development shortcuts
.PHONY: t tc tm th tr

t: test
tc: test-coverage
tm: test-models
th: test-handlers
tr: test-repos