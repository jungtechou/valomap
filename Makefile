# Valorant Map Picker Makefile

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

ifeq (deploy,$(firstword $(MAKECMDGOALS)))
  # If the first goal is "deploy", we need to capture the second argument (if any)
  DEPLOY_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(DEPLOY_ARGS):;@:)
endif

.PHONY: setup deploy test help

setup:
	@echo "Ensuring external network..."
	@docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)

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

deploy: setup
	@if [ -z "$(TEST_ARGS)" ]; then \
		@echo "Deploying all components to production..."; \
		@$(DOCKER_COMPOSE) down --remove-orphans; \
		@$(DOCKER_COMPOSE) -f docker-compose.yml up --remove-orphans; \
	else \
		@echo "Deploying $(TEST_ARGS) to production ..."; \
		@$(DOCKER_COMPOSE) rm $(TEST_ARGS) -f; \
		@$(DOCKER_COMPOSE) -f docker-compose.yml up --remove-orphans --build $(TEST_ARGS); \
	fi

test:
	@if [ -z "$(TEST_ARGS)" ]; then \
		@echo "Running all tests..."; \
		$(MAKE) -C frontend test; \
		$(MAKE) -C backend test; \
	else \
		@echo "Running $(TEST_ARGS) tests..."; \
		$(MAKE) -C $(TEST_ARGS) test; \
	fi

help:
	@echo "Valorant Map Picker - Docker Management Commands:"
	@echo ""
	@echo "  make setup         - Initialize project setup (network, Traefik, credentials)"
	@echo "  make deploy        - Deploy services to production"
	@echo "  make test          - Run tests (all or specific component)"
	@echo ""
	@echo "Examples:"
	@echo "  make deploy backend    - Deploy only backend service"
	@echo "  make deploy frontend   - Deploy only frontend service"
	@echo "  make deploy            - Deploy all services"
	@echo "  make test backend      - Run backend tests only"
	@echo "  make test frontend     - Run frontend tests only"
	@echo "  make test              - Run all tests"
	@echo ""
