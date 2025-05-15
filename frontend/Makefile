# Frontend Tests Makefile

# Variables
TEST_RESULTS_DIR = test_results
DOCKER_COMPOSE = COMPOSE_BAKE=true docker compose

.PHONY: test test-unit

# Default target: run all tests
test:
	@mkdir -pv $(TEST_RESULTS_DIR)
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up --abort-on-container-exit --remove-orphans --build test

# Run all types of tests
test-all:
	@mkdir -pv $(TEST_RESULTS_DIR)
	@echo "Running Jest tests with coverage..."
	@JEST_JUNIT_OUTPUT_DIR="$TEST_RESULTS_DIR" npm test -- --coverage --ci --reporters=default --reporters=jest-junit

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -rfv $(TEST_RESULTS_DIR)
	@echo "Clean complete!"

# Help information
help:
	@echo "Frontend Tests Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make test          - Run all tests in Docker container"
	@echo "  make test-all      - Run all types of tests"
	@echo "  make clean         - Clean test artifacts"
	@echo "  make help          - Show this help message"
	@echo ""
