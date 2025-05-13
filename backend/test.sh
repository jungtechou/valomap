#!/bin/bash
set -e

# Default test type
TEST_TYPE="all"

# Default mode (local or docker)
MODE="local"

# Help function
function show_help {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  -t, --test-type TYPE  Test type: unit, coverage, benchmark, race, or all (default)"
  echo "  -d, --docker          Run tests in Docker container"
  echo "  -h, --help            Show this help message"
  echo ""
  echo "Examples:"
  echo "  $0                    Run all tests locally"
  echo "  $0 -t unit            Run unit tests locally"
  echo "  $0 -t coverage -d     Run coverage tests in Docker"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -t|--test-type)
      TEST_TYPE="$2"
      shift 2
      ;;
    -d|--docker)
      MODE="docker"
      shift
      ;;
    -h|--help)
      show_help
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      show_help
      exit 1
      ;;
  esac
done

# Validate test type
if [[ ! "$TEST_TYPE" =~ ^(unit|coverage|benchmark|race|all)$ ]]; then
  echo "Error: Invalid test type '$TEST_TYPE'"
  show_help
  exit 1
fi

# Function to run tests locally
function run_local_tests {
  echo "Running $TEST_TYPE tests locally..."

  case "$TEST_TYPE" in
    unit)
      go test -v ./...
      ;;
    coverage)
      PACKAGES=$(go list ./... | grep -v "github.com/jungtechou/valomap/cmd" | grep -v "github.com/jungtechou/valomap/di")
      go test -coverprofile=coverage.out $PACKAGES
      mkdir -p coverage_report
      go tool cover -html=coverage.out -o coverage_report/coverage.html
      go tool cover -func=coverage.out | tee coverage_report/coverage_summary.txt
      ;;
    benchmark)
      go test -run=^$ -bench=. -benchmem ./... | tee benchmark_results.txt
      ;;
    race)
      CGO_ENABLED=1 go test -race -short ./...
      ;;
    all)
      go test -v ./...
      PACKAGES=$(go list ./... | grep -v "github.com/jungtechou/valomap/cmd" | grep -v "github.com/jungtechou/valomap/di")
      go test -coverprofile=coverage.out $PACKAGES
      mkdir -p coverage_report
      go tool cover -html=coverage.out -o coverage_report/coverage.html
      go tool cover -func=coverage.out | tee coverage_report/coverage_summary.txt
      go test -run=^$ -bench=. -benchmem ./... | tee benchmark_results.txt
      CGO_ENABLED=1 go test -race -short ./...
      ;;
  esac
}

# Function to run tests in Docker
function run_docker_tests {
  echo "Running $TEST_TYPE tests in Docker container..."

  # Build the test Docker image
  docker build -t valomap-backend-test -f Dockerfile.test .

  # Create a volume for coverage reports
  if [[ "$TEST_TYPE" == "coverage" || "$TEST_TYPE" == "all" ]]; then
    docker volume create valomap-coverage-volume
  fi

  # Run the tests in Docker
  if [[ "$TEST_TYPE" == "coverage" || "$TEST_TYPE" == "all" ]]; then
    docker run --rm -v valomap-coverage-volume:/app/coverage_report valomap-backend-test "$TEST_TYPE"

    # Copy coverage reports from volume to host
    echo "Copying coverage reports from Docker volume..."
    TEMP_CONTAINER=$(docker create -v valomap-coverage-volume:/data busybox)
    mkdir -p ./coverage_report
    docker cp $TEMP_CONTAINER:/data/. ./coverage_report/
    docker rm $TEMP_CONTAINER
  else
    docker run --rm valomap-backend-test "$TEST_TYPE"
  fi
}

# Execute tests based on mode
if [[ "$MODE" == "docker" ]]; then
  run_docker_tests
else
  run_local_tests
fi

echo "Done!"
