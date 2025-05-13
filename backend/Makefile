# Backend Makefile
# Commands specific to the backend service

# Variables
GO = go
DOCKER_COMPOSE = COMPOSE_BAKE=true docker compose
TEST_RESULTS_DIR = ../test_results
SWAG = $(shell command -v swag 2> /dev/null)

#----------------------------------------------
# Main targets
#----------------------------------------------
.PHONY: setup clean dev build test test-docker test-docker-unit test-docker-coverage test-docker-benchmark test-docker-race swagger

# Setup the backend
setup:
	@echo "Setting up backend..."
	@go mod tidy
	@chmod +x test.sh

# Clean backend resources
clean:
	@echo "Cleaning backend resources..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down 2>/dev/null || true
	@sleep 2  # Give Docker time to release resources
	@rm -rf coverage_report || true
	@rm -f coverage.out benchmark_results.txt || true
	@mkdir -p coverage_report
	@$(GO) clean

# Build backend
build:
	@echo "Building backend..."
	@$(GO) build -o bin/server ./cmd/main.go

# Run backend in development mode
dev:
	@echo "Starting backend in development mode..."
	@$(GO) run ./cmd/main.go

#----------------------------------------------
# Test targets
#----------------------------------------------

# Run backend tests in Docker (default target)
test: test-docker

# Run all backend tests in Docker
test-docker:
	@echo "Running backend tests in Docker..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans test
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down

# Run backend unit tests in Docker
test-docker-unit:
	@echo "Running backend unit tests in Docker..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans test-unit
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down

# Run backend coverage tests in Docker
test-docker-coverage:
	@echo "Running backend coverage tests in Docker..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans test-coverage
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down

# Run backend benchmark tests in Docker
test-docker-benchmark:
	@echo "Running backend benchmark tests in Docker..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans test-benchmark
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down

# Run backend race condition tests in Docker
test-docker-race:
	@echo "Running backend race condition tests in Docker..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans test-race
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down

# View test coverage results
test-coverage-view:
	@echo "Opening test coverage reports..."
	@[ -f coverage_report/coverage.html ] && open coverage_report/coverage.html || echo "Error: Coverage report not found. Run test-docker-coverage first."
	@echo "Coverage reports available in coverage_report directory"

#----------------------------------------------
# Documentation targets
#----------------------------------------------

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if [ -z "$(SWAG)" ]; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@echo "Running swag init from $$(pwd)"
	@swag init -g cmd/main.go -o docs
	@echo "Swagger documentation generated successfully!"

#----------------------------------------------
# Helper targets
#----------------------------------------------
.PHONY: help ensure-test-dir stop-tests

# Ensure test directory exists
ensure-test-dir:
	@mkdir -p $(TEST_RESULTS_DIR)

# Stop test containers
stop-tests:
	@echo "Stopping test containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down 2>/dev/null || true
	@sleep 2  # Give Docker time to release resources

# Help command
help:
	@echo "Backend Makefile - Available commands:"
	@echo ""
	@echo "Main commands:"
	@echo "  make setup         - Setup backend dependencies"
	@echo "  make build         - Build backend binaries"
	@echo "  make clean         - Clean backend resources"
	@echo "  make dev           - Run backend in development mode"
	@echo ""
	@echo "Test commands (Docker only):"
	@echo "  make test          - Run backend tests in Docker"
	@echo "  make test-docker   - Run all backend tests in Docker"
	@echo "  make test-docker-unit - Run backend unit tests in Docker"
	@echo "  make test-docker-coverage - Run backend coverage tests in Docker"
	@echo "  make test-docker-benchmark - Run backend benchmark tests in Docker"
	@echo "  make test-docker-race - Run backend race condition tests in Docker"
	@echo "  make test-coverage-view - View test coverage reports"
	@echo "  make stop-tests    - Stop all test containers"
	@echo ""
	@echo "Documentation commands:"
	@echo "  make swagger       - Generate Swagger documentation"
