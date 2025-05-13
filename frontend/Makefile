# Frontend Makefile
# Commands specific to the frontend service

# Variables
NPM = npm
DOCKER_COMPOSE = COMPOSE_BAKE=true docker compose
TEST_RESULTS_DIR = ../test_results

#----------------------------------------------
# Main targets
#----------------------------------------------
.PHONY: setup clean dev build test test-e2e test-docker test-docker-coverage test-docker-perf

# Setup the frontend
setup:
	@echo "Setting up frontend..."
	@$(NPM) install

# Clean frontend resources
clean:
	@echo "Cleaning frontend resources..."
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml down 2>/dev/null || true
	@sleep 2  # Give Docker time to release resources
	@$(NPM) run clean || true

# Build frontend
build:
	@echo "Building frontend..."
	@$(NPM) run build

# Run frontend in development mode
dev:
	@echo "Starting frontend in development mode..."
	@$(NPM) install && $(NPM) run dev

#----------------------------------------------
# Test targets
#----------------------------------------------

# Run frontend tests
test: ensure-test-dir
	@echo "Running frontend tests..."
	@$(NPM) test -- --coverage

# Run end-to-end tests
test-e2e: ensure-test-dir
	@echo "Running end-to-end tests..."
	@$(NPM) run test:e2e

# Run frontend tests in Docker
test-docker:
	@echo "Running frontend tests in Docker..."
	@$(NPM) run test:compose

# Run frontend coverage tests in Docker
test-docker-coverage:
	@echo "Running frontend coverage tests in Docker..."
	@$(NPM) run test:compose:coverage

# Run frontend performance tests in Docker
test-docker-perf:
	@echo "Running frontend performance tests in Docker..."
	@$(NPM) run test:compose:perf

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
	@echo "Frontend Makefile - Available commands:"
	@echo ""
	@echo "Main commands:"
	@echo "  make setup         - Setup frontend dependencies"
	@echo "  make build         - Build frontend assets"
	@echo "  make clean         - Clean frontend resources"
	@echo "  make dev           - Run frontend in development mode"
	@echo ""
	@echo "Test commands:"
	@echo "  make test          - Run frontend tests"
	@echo "  make test-e2e      - Run end-to-end tests"
	@echo "  make test-docker   - Run frontend tests in Docker"
	@echo "  make test-docker-coverage - Run frontend coverage tests in Docker"
	@echo "  make test-docker-perf - Run frontend performance tests in Docker"
	@echo "  make stop-tests    - Stop test containers"
