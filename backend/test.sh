#!/bin/bash
#
# Comprehensive test script for the valomap backend
# This script runs unit tests, coverage tests, benchmark tests, and race condition tests
#

set -e

# Print section header
print_header() {
    echo
    echo "======================================================================"
    echo "  $1"
    echo "======================================================================"
    echo
}

# Clean up previous test results
cleanup() {
    print_header "Cleaning up previous test results"
    rm -f coverage.out
    rm -f benchmark_results.txt
    rm -rf coverage_report
    mkdir -p coverage_report
    echo "Done!"
}

# Run unit tests
run_unit_tests() {
    print_header "Running Unit Tests"
    go test -v ./...
}

# Run tests with coverage
run_coverage_tests() {
    print_header "Running Tests with Coverage"
    # Run tests excluding the problematic packages
    PACKAGES=$(go list ./... | grep -v "github.com/jungtechou/valomap/cmd" | grep -v "github.com/jungtechou/valomap/di")
    go test -coverprofile=coverage.out $PACKAGES
    go tool cover -html=coverage.out -o coverage_report/coverage.html
    go tool cover -func=coverage.out | tee coverage_report/coverage_summary.txt

    # Print coverage summary
    echo
    echo "Coverage Summary:"
    tail -n 1 coverage_report/coverage_summary.txt
    echo
    echo "Detailed coverage report saved to coverage_report/coverage.html"
}

# Run benchmark tests
run_benchmark_tests() {
    print_header "Running Benchmark Tests"
    go test -run=^$ -bench=. -benchmem ./... | tee benchmark_results.txt
    echo
    echo "Benchmark results saved to benchmark_results.txt"
}

# Run race condition tests
run_race_tests() {
    print_header "Running Race Condition Tests"
    CGO_ENABLED=1 go test -race -short ./...
}

# Run all tests with a nice summary
run_all_tests() {
    # Start timer
    start_time=$(date +%s)

    cleanup
    run_unit_tests
    run_coverage_tests
    run_benchmark_tests
    run_race_tests

    # End timer and calculate duration
    end_time=$(date +%s)
    duration=$((end_time - start_time))

    print_header "Test Summary"
    echo "All tests completed successfully in $duration seconds"
    echo
    echo "Coverage Summary:"
    tail -n 1 coverage_report/coverage_summary.txt
    echo
    echo "Test artifacts saved to:"
    echo "- coverage_report/coverage.html (HTML coverage report)"
    echo "- coverage_report/coverage_summary.txt (Coverage summary)"
    echo "- benchmark_results.txt (Benchmark results)"
}

# Process command line arguments
case "$1" in
    clean)
        cleanup
        ;;
    unit)
        run_unit_tests
        ;;
    coverage)
        cleanup
        run_coverage_tests
        ;;
    benchmark)
        cleanup
        run_benchmark_tests
        ;;
    race)
        run_race_tests
        ;;
    *)
        run_all_tests
        ;;
esac
