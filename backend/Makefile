# Backend Go Tests Makefile

# Variables
GOTEST = go test
GOCMD = go
TEST_DIR = ./...
COVERAGE_OUT = coverage.out
COVERAGE_HTML = coverage_report/coverage.html
COVERAGE_SUMMARY = coverage_report/coverage_summary.txt
BENCHMARK_RESULTS = benchmark_results.txt
# Exclude packages that shouldn't be included in coverage
PACKAGES = $(shell go list $(TEST_DIR) | grep -v "github.com/jungtechou/valomap/cmd" | grep -v "github.com/jungtechou/valomap/di")

.PHONY: test test-unit test-coverage test-benchmark test-race test-all clean view-coverage

# Default target: run all tests
test: test-all

# Run unit tests only
test-unit:
	@echo "==> Running Unit Tests"
	@$(GOTEST) -v $(TEST_DIR)

# Run tests with coverage
test-coverage:
	@echo "==> Running Tests with Coverage"
	@mkdir -p coverage_report
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
	@echo "==> Running All Tests"
	@$(GOTEST) -v $(TEST_DIR)
	@mkdir -p coverage_report
	@$(GOTEST) -coverprofile=$(COVERAGE_OUT) $(PACKAGES)
	@$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@$(GOCMD) tool cover -func=$(COVERAGE_OUT) | tee $(COVERAGE_SUMMARY)
	@$(GOTEST) -run=^$ -bench=. -benchmem $(TEST_DIR) | tee $(BENCHMARK_RESULTS)
	@CGO_ENABLED=1 $(GOTEST) -race -short $(TEST_DIR)
	@echo "All tests completed!"

# View coverage report in the default browser (works on Linux, macOS, and Windows)
view-coverage:
	@echo "Opening coverage report..."
	@if [ -f "$(COVERAGE_HTML)" ]; then \
		case "$$(uname -s)" in \
			Linux*) xdg-open $(COVERAGE_HTML) ;; \
			Darwin*) open $(COVERAGE_HTML) ;; \
			CYGWIN*|MINGW*|MSYS*) start $(COVERAGE_HTML) ;; \
			*) echo "Unsupported OS. Please open $(COVERAGE_HTML) manually." ;; \
		esac; \
	else \
		echo "Coverage report not found. Run 'make test-coverage' first."; \
	fi

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -f $(COVERAGE_OUT)
	@rm -rf coverage_report
	@rm -f $(BENCHMARK_RESULTS)
	@echo "Clean complete!"

# Help information
help:
	@echo "Backend Go Tests Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make test          - Run all tests (same as test-all)"
	@echo "  make test-unit     - Run only unit tests"
	@echo "  make test-coverage - Run tests with coverage reporting"
	@echo "  make test-benchmark - Run benchmark tests"
	@echo "  make test-race     - Run race condition tests"
	@echo "  make test-all      - Run all types of tests"
	@echo "  make view-coverage - Open coverage report in browser"
	@echo "  make clean         - Clean test artifacts"
	@echo "  make help          - Show this help message"
	@echo ""
