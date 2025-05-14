# Backend Go Tests Makefile

# Variables
GOTEST = go test
GOCMD = go
TEST_DIR = ./...
TEST_RESULTS_DIR = test_results
COVERAGE_OUT = $(TEST_RESULTS_DIR)/coverage.out
COVERAGE_HTML = $(TEST_RESULTS_DIR)/coverage.html
COVERAGE_SUMMARY = $(TEST_RESULTS_DIR)/coverage_summary.txt
BENCHMARK_RESULTS = $(TEST_RESULTS_DIR)/benchmark_results.txt
# Exclude packages that shouldn't be included in coverage
PACKAGES = $(shell go list $(TEST_DIR) | grep -v "github.com/jungtechou/valomap/cmd" | grep -v "github.com/jungtechou/valomap/di")
DOCKER_COMPOSE = COMPOSE_BAKE=true docker compose

.PHONY: test test-unit test-coverage test-benchmark test-race test-all clean view-coverage

# Default target: run all tests
test:
	@mkdir -pv $(TEST_RESULTS_DIR)
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up --abort-on-container-exit --remove-orphans --build test

# Run unit tests only
test-unit:
	@echo "==> Running Unit Tests"
	@$(GOTEST) -v $(TEST_DIR)

# Run tests with coverage
test-coverage:
	@echo "==> Running Tests with Coverage"
	@$(GOTEST) -coverprofile=$(COVERAGE_OUT) $(PACKAGES)
	@$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@$(GOCMD) tool cover -func=$(COVERAGE_OUT) | tee $(COVERAGE_SUMMARY)
	@echo "Coverage report generated at $(COVERAGE_HTML)"

# Run benchmark tests
test-benchmark:
	@echo "==> Running Benchmark Tests"
	@$(GOTEST) -run=^$ -bench=. -benchmem $(TEST_DIR) | tee $(BENCHMARK_RESULTS)
	@echo "Benchmark results saved to $(BENCHMARK_RESULTS)"

# Run race condition tests
test-race:
	@echo "==> Running Race Condition Tests"
	@CGO_ENABLED=1 $(GOTEST) -race -short $(TEST_DIR)

# Run all types of tests
test-all:
	@echo "==> Running All Backend Tests"
	@$(GOTEST) -v $(TEST_DIR)
	@mkdir -p $(TEST_RESULTS_DIR)
	@$(GOTEST) -coverprofile=$(COVERAGE_OUT) $(PACKAGES)
	@$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@$(GOCMD) tool cover -func=$(COVERAGE_OUT) | tee $(COVERAGE_SUMMARY)
	@$(GOTEST) -run=^$ -bench=. -benchmem $(TEST_DIR) | tee $(BENCHMARK_RESULTS)
	@CGO_ENABLED=1 $(GOTEST) -race -short $(TEST_DIR)
	@echo "All tests completed!"

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -rfv $(TEST_RESULTS_DIR)
	@echo "Clean complete!"

# Help information
help:
	@echo "Backend Go Tests Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make test          - Run all tests in Docker container"
	@echo "  make test-unit     - Run only unit tests"
	@echo "  make test-coverage - Run tests with coverage reporting"
	@echo "  make test-benchmark - Run benchmark tests"
	@echo "  make test-race     - Run race condition tests"
	@echo "  make test-all      - Run all types of tests"
	@echo "  make clean         - Clean test artifacts"
	@echo "  make help          - Show this help message"
	@echo ""
