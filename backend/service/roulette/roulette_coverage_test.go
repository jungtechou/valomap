package roulette

/*
# Test Coverage Guide

This file contains instructions for running the tests with coverage reporting.

## Running Tests with Coverage

To run tests with coverage reporting in the terminal:

```bash
go test -cover ./...
```

For more detailed coverage information, you can generate an HTML report:

```bash
# Run the tests and generate a coverage profile
go test -coverprofile=coverage.out ./...

# Generate an HTML report
go tool cover -html=coverage.out
```

## Coverage Goals

Aim for at least 80% code coverage for all service implementations. For critical
components like data processing and error handling, aim for 90% or higher coverage.

## Understanding Coverage Reports

The coverage report shows which lines of code are executed during tests:
- Green: Code executed during tests
- Red: Code not executed during tests
- Gray: Non-executable code (e.g., comments, blank lines)

## Important Areas to Test

For the RouletteService, ensure good coverage for:
1. Error handling in the API client
2. Map filtering logic including edge cases (no maps, all banned, etc.)
3. Random map selection
4. Concurrent access safeguards

## Running Race Detection Tests

To check for race conditions, run tests with the race detector:

```bash
go test -race ./...
```

This is especially important for this service due to the shared random number generator
and potential concurrent access to external resources.
*/

// The tests for this service are split into different files for better organization:
// - roulette_test.go: Unit tests for main functionality
// - roulette_benchmark_test.go: Performance benchmarks for key operations
// - roulette_conflict_test.go: Tests for concurrent access and race conditions
