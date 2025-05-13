# Valorant Map Picker Makefile
# Main orchestration for the entire project

# Variables
DOCKER_COMPOSE = COMPOSE_BAKE=true docker compose
DOCKER = docker
NETWORK_NAME = web
TEST_RESULTS_DIR = test_results

# Command line argument parsing
ifeq (test,$(firstword $(MAKECMDGOALS)))
  # If the first goal is "test", we need to capture the second argument (if any)
  TEST_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(TEST_ARGS):;@:)
endif

ifeq (dev,$(firstword $(MAKECMDGOALS)))
  # If the first goal is "dev", we need to capture the second argument (if any)
  DEV_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(DEV_ARGS):;@:)
endif

#----------------------------------------------
# Main targets
#----------------------------------------------
.PHONY: all setup clean test build run stop deploy dev

# Default target: build and run without setup
all: build run

# First time setup
setup: ensure-network ensure-test-dir
	@echo "Setting up project..."
	@if [ ! -f ".env" ]; then \
		echo "Creating default .env file..."; \
		cp .env.example .env; \
	fi
	@$(MAKE) -C frontend setup
	@$(MAKE) -C backend setup
	@echo "Setup complete!"

# Clean up project files
clean:
	@echo "Cleaning up..."
	@$(MAKE) -C backend clean
	@$(MAKE) -C frontend clean
	@rm -rf $(TEST_RESULTS_DIR) || true
	@$(DOCKER) system prune -f

# Build all components
build:
	@echo "Building all components..."
	@$(MAKE) -C backend build
	@$(MAKE) -C frontend build
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) build

# Run the application
run: ensure-network
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d

# Stop the application
stop:
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down

# Deploy to production
deploy: build
	@echo "Deploying to production..."
	@echo "Stopping any running services..."
	@$(DOCKER_COMPOSE) down
	@echo "Starting production services..."
	@COMPOSE_PROFILES=production $(DOCKER_COMPOSE) up -d
	@echo "Deployment complete!"

# Run in development mode with optional component
dev: ensure-network
	@echo "Starting in development mode..."
	@if [ "$(DEV_ARGS)" = "frontend" ]; then \
		$(MAKE) -C frontend dev; \
	elif [ "$(DEV_ARGS)" = "backend" ]; then \
		$(MAKE) -C backend dev; \
	else \
		$(MAKE) -C backend dev & \
		$(MAKE) -C frontend dev; \
	fi

# Run tests with optional component
test: ensure-test-dir
	@echo "Running tests..."
	@if [ "$(TEST_ARGS)" = "frontend" ]; then \
		$(MAKE) -C frontend test; \
	elif [ "$(TEST_ARGS)" = "backend" ]; then \
		$(MAKE) -C backend test; \
	elif [ "$(TEST_ARGS)" = "backend-unit" ]; then \
		$(MAKE) -C backend test-docker-unit; \
	elif [ "$(TEST_ARGS)" = "backend-coverage" ]; then \
		$(MAKE) -C backend test-docker-coverage; \
	elif [ "$(TEST_ARGS)" = "backend-benchmark" ]; then \
		$(MAKE) -C backend test-docker-benchmark; \
	elif [ "$(TEST_ARGS)" = "backend-race" ]; then \
		$(MAKE) -C backend test-docker-race; \
	elif [ "$(TEST_ARGS)" = "e2e" ]; then \
		$(MAKE) -C frontend test-e2e; \
	elif [ "$(TEST_ARGS)" = "coverage" ]; then \
		$(MAKE) -C backend test-docker-coverage; \
	elif [ "$(TEST_ARGS)" = "docker" ]; then \
		$(MAKE) -C frontend test-docker; \
		$(MAKE) -C backend test-docker; \
	else \
		echo "Running all tests..."; \
		$(MAKE) -C frontend test; \
		$(MAKE) -C backend test; \
		$(MAKE) -C frontend test-e2e; \
	fi

#----------------------------------------------
# Docker utility targets
#----------------------------------------------
.PHONY: docker-push docker-prune

# Push Docker images to registry (requires build first)
docker-push: build
	@echo "Pushing Docker images to registry..."
	@$(DOCKER_COMPOSE) push

# Clean up Docker resources
docker-prune:
	@echo "Pruning Docker resources..."
	@$(DOCKER) system prune -f
	@$(DOCKER) volume prune -f

#----------------------------------------------
# Helper targets
#----------------------------------------------
.PHONY: logs restart status help ensure-network ensure-test-dir

# View application logs
logs:
	@$(DOCKER_COMPOSE) logs -f

# Restart the application
restart: stop run

# Check application status
status:
	@$(DOCKER_COMPOSE) ps

# Ensure docker network exists
ensure-network:
	@docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)

# Ensure test directory exists
ensure-test-dir:
	@mkdir -p $(TEST_RESULTS_DIR)

# Help command
help:
	@echo "Valorant Map Picker - Available commands:"
	@echo ""
	@echo "Main commands:"
	@echo "  make               - Build and run the application"
	@echo "  make setup         - Setup project for first time use"
	@echo "  make build         - Build all components and Docker images"
	@echo "  make run           - Start all services"
	@echo "  make stop          - Stop all services"
	@echo "  make restart       - Restart all services"
	@echo "  make deploy        - Deploy to production"
	@echo "  make clean         - Clean up project files"
	@echo ""
	@echo "Development commands:"
	@echo "  make dev           - Run all components in development mode"
	@echo "  make dev frontend  - Run frontend in development mode"
	@echo "  make dev backend   - Run backend in development mode"
	@echo ""
	@echo "Test commands:"
	@echo "  make test          - Run all tests"
	@echo "  make test frontend - Run frontend tests"
	@echo "  make test backend  - Run backend tests in Docker"
	@echo "  make test backend-unit - Run backend unit tests in Docker"
	@echo "  make test backend-coverage - Run backend coverage tests in Docker"
	@echo "  make test backend-benchmark - Run backend benchmark tests in Docker"
	@echo "  make test backend-race - Run backend race tests in Docker"
	@echo "  make test e2e      - Run end-to-end tests"
	@echo "  make test coverage - Generate test coverage reports"
	@echo "  make test docker   - Run all tests in Docker containers"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-push   - Push Docker images to registry"
	@echo "  make docker-prune  - Clean up Docker resources"
	@echo ""
	@echo "Utility commands:"
	@echo "  make logs          - View application logs"
	@echo "  make status        - Check application status"
	@echo ""
	@echo "For more specific commands:"
	@echo "  make -C frontend help  - Show frontend-specific commands"
	@echo "  make -C backend help   - Show backend-specific commands"
