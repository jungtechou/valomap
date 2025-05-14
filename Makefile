# Valorant Map Picker Makefile
# Simplified version with Docker-only operations

# Variables
DOCKER_COMPOSE = COMPOSE_BAKE=true docker compose
DOCKER = docker
NETWORK_NAME = web
TRAEFIK_ACME_JSON = traefik/acme.json

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
.PHONY: traefik setup build deploy test test-coverage benchmark-test view-coverage help ensure-network

traefik: ensure-network
	@echo "Setting up Traefik..."
	@touch $(TRAEFIK_ACME_JSON)
	@chmod 600 $(TRAEFIK_ACME_JSON)

	@echo "Creating admin user for Traefik dashboard"
	@echo "Please enter a password for the Traefik dashboard:"
	@stty -echo; read PASSWORD; stty echo

	@echo "Hashing password..."
	HASHED_PASSWORD=$$(docker run --rm httpd:alpine htpasswd -nbB admin "$$PASSWORD" | cut -d ":" -f 2)

	@echo "Update the docker-compose.yml file with the hashed password"
	@sed -i "s|\$apr1\$70hN10X7\$3QbzMaVnA3pagO1OJl1o90|$HASHED_PASSWORD|g" docker-compose.yml

	@echo "Please enter your email address for Let's Encrypt notifications:"
	@read EMAIL
	@sed -i "s|your-email@example.com|$EMAIL|g" traefik/traefik.yml

	@echo "Setup completed successfully!"

# Setup: Prepare environment and dependencies in Docker
setup: ensure-network
	@echo "Setting up project in Docker..."
#	@if [ ! -f ".env" ]; then \
#		echo "Creating default .env file..."; \
#		cp .env.example .env; \
#	fi
	@if [ "$(TEST_ARGS)" = "frontend" ]; then \
		$(MAKE) -C frontend test; \
	elif [ "$(TEST_ARGS)" = "backend" ]; then \
		$(MAKE) -C backend test; \
	else \
		@echo "Building setup containers..." \
		@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.setup.yml build setup-frontend setup-backend \
		@echo "Running setup for frontend..." \
		@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.setup.yml run --rm setup-frontend \
		@echo "Running setup for backend..." \
		@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.setup.yml run --rm setup-backend \
	fi
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

# Run tests with optional component
test:
	@echo "Running tests..."
	@if [ "$(TEST_ARGS)" = "frontend" ]; then \
		$(MAKE) -C frontend test; \
	fi

	@if [ "$(TEST_ARGS)" = "backend" ]; then \
		$(MAKE) -C backend test; \
	fi

#----------------------------------------------
# Helper targets
#----------------------------------------------

# Ensure docker network exists
ensure-network:
	@docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)

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
