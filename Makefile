# Valorant Map Picker Makefile
# Simplified version with Docker-only operations

# Variables
DOCKER_COMPOSE = COMPOSE_BAKE=true docker compose
DOCKER = docker
NETWORK_NAME = web
TEST_RESULTS_DIR = test_results

#----------------------------------------------
# Main targets
#----------------------------------------------
.PHONY: setup build deploy test test-coverage benchmark-test view-coverage help ensure-network ensure-test-dir

# Setup: Prepare environment and dependencies in Docker
setup: ensure-network ensure-test-dir
	@echo "Setting up project in Docker..."
	@if [ ! -f ".env" ]; then \
		echo "Creating default .env file..."; \
		cp .env.example .env; \
	fi
	@echo "Building setup containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.setup.yml build setup-frontend setup-backend
	@echo "Running setup for frontend..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.setup.yml run --rm setup-frontend
	@echo "Running setup for backend..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.setup.yml run --rm setup-backend
	@echo "Setup complete!"

# Build all components in Docker
build: ensure-network
	@echo "Building all components in Docker..."
	@$(DOCKER_COMPOSE) build --no-cache
	@echo "Build complete!"

# Deploy to production using Docker
deploy: ensure-network
	@echo "Deploying to production using Docker..."
	@echo "Stopping any running services..."
	@$(DOCKER_COMPOSE) down
	@echo "Starting production services..."
	@COMPOSE_PROFILES=production $(DOCKER_COMPOSE) up -d
	@echo "Deployment complete!"

# Run all tests in Docker
test: ensure-network ensure-test-dir
	@echo "Running all tests in Docker..."
	@echo "Running frontend tests..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans frontend-test
	@echo "Running backend tests..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans backend-test
	@echo "Running end-to-end tests..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans e2e-test
	@echo "Cleaning up test containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml down
	@echo "All tests completed!"

# Run tests with detailed coverage reports
test-coverage: ensure-network ensure-test-dir
	@echo "Running tests with coverage reporting in Docker..."
	@echo "Running frontend coverage tests..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans frontend-coverage
	@echo "Running backend coverage tests..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans backend-coverage
	@echo "Cleaning up test containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml down
	@echo "Coverage tests completed! Use 'make view-coverage' to see reports."

# Run benchmark tests
benchmark-test: ensure-network ensure-test-dir
	@echo "Running benchmark tests in Docker..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --remove-orphans backend-benchmark
	@echo "Cleaning up benchmark containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.test.yml down
	@echo "Benchmark tests completed!"

# View coverage reports
view-coverage:
	@echo "Opening coverage reports..."
	@if [ -d "$(TEST_RESULTS_DIR)/frontend-coverage" ]; then \
		echo "Frontend coverage report at $(TEST_RESULTS_DIR)/frontend-coverage/index.html"; \
	else \
		echo "Frontend coverage report not found. Run 'make test-coverage' first."; \
	fi
	@if [ -d "$(TEST_RESULTS_DIR)/backend-coverage" ]; then \
		echo "Backend coverage report at $(TEST_RESULTS_DIR)/backend-coverage/index.html"; \
	else \
		echo "Backend coverage report not found. Run 'make test-coverage' first."; \
	fi

#----------------------------------------------
# Helper targets
#----------------------------------------------

# Ensure docker network exists
ensure-network:
	@docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)

# Ensure test directory exists
ensure-test-dir:
	@mkdir -p $(TEST_RESULTS_DIR)
	@mkdir -p $(TEST_RESULTS_DIR)/frontend-coverage
	@mkdir -p $(TEST_RESULTS_DIR)/backend-coverage
	@mkdir -p $(TEST_RESULTS_DIR)/benchmark

# Help command
help:
	@echo "Valorant Map Picker - Simplified Docker Commands:"
	@echo ""
	@echo "  make setup         - Setup project in Docker containers"
	@echo "  make build         - Build all components in Docker"
	@echo "  make deploy        - Deploy to production using Docker"
	@echo "  make test          - Run all tests in Docker containers"
	@echo "  make test-coverage - Run tests with detailed coverage reports"
	@echo "  make benchmark-test - Run benchmark tests"
	@echo "  make view-coverage - View generated coverage reports"
	@echo ""
