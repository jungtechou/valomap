#!/bin/bash

# Set some color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}===== VALORANT MAP PICKER TESTS & PERFORMANCE METRICS =====${NC}"
echo

# Verify working directory
if [ ! -d "frontend" ] || [ ! -d "backend" ]; then
  echo -e "${RED}Error: Please run this script from the project root directory.${NC}"
  exit 1
fi

# Create results directory
RESULTS_DIR="test_results"
mkdir -p "$RESULTS_DIR"

# ======= FRONTEND TESTS =======
echo -e "${GREEN}Running Frontend Tests...${NC}"
cd frontend

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
  echo "Installing frontend dependencies..."
  npm install
fi

# Run tests with coverage
echo "Running Jest tests with coverage..."
JEST_JUNIT_OUTPUT_DIR="$RESULTS_DIR" npm test -- --coverage --ci --reporters=default --reporters=jest-junit

# Save coverage report
cp -r coverage "../$RESULTS_DIR/frontend-coverage"

# ======= BACKEND TESTS =======
echo -e "${GREEN}Running Backend Tests...${NC}"
cd ../backend

# Run Go tests with coverage
echo "Running Go tests with coverage..."
go test -cover -coverprofile=../test_results/backend-coverage.out ./...

# Generate coverage HTML report
go tool cover -html=../test_results/backend-coverage.out -o ../test_results/backend-coverage.html

cd ..

# ======= PERFORMANCE BENCHMARKS =======
echo -e "${GREEN}Running Performance Benchmarks...${NC}"

# Ensure server is running (you might need to adjust this part)
SERVER_RUNNING=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/health || echo "000")

if [ "$SERVER_RUNNING" != "200" ]; then
  echo -e "${YELLOW}Warning: Server doesn't seem to be running. Skipping performance tests.${NC}"
  echo -e "${YELLOW}Start the server with 'docker-compose up' to run performance tests.${NC}"
else
  # Run benchmark with ApacheBench (ab)
  echo "Benchmarking API Endpoint..."
  ab -n 1000 -c 10 -H "Accept-Encoding: gzip, deflate" http://localhost:3000/api/v1/map/all > "$RESULTS_DIR/api-benchmark.txt"

  # Run benchmark for image loading
  echo "Benchmarking Image Loading..."
  # Get a cached image URL first
  IMAGE_URL=$(curl -s http://localhost:3000/api/v1/map/all | grep -o '/api/cache/[^"]*' | head -1)
  if [ -n "$IMAGE_URL" ]; then
    ab -n 200 -c 5 http://localhost:3000$IMAGE_URL > "$RESULTS_DIR/image-benchmark.txt"
  else
    echo -e "${YELLOW}Warning: No cached image found. Skipping image benchmark.${NC}"
  fi
fi

# ======= LIGHTHOUSE PERFORMANCE =======
echo -e "${GREEN}Running Lighthouse Performance Audit...${NC}"

if command -v lighthouse &> /dev/null; then
  lighthouse http://localhost --output-path="$RESULTS_DIR/lighthouse-report.html" --quiet --chrome-flags="--headless"
else
  echo -e "${YELLOW}Lighthouse not found. Install with 'npm install -g lighthouse'${NC}"
fi

# ======= SUMMARY =======
echo
echo -e "${BLUE}===== TEST AND PERFORMANCE SUMMARY =====${NC}"
echo
echo -e "${GREEN}Frontend Test Coverage:${NC}"
cat "$RESULTS_DIR/frontend-coverage/lcov-report/index.html" | grep -o "Total.*" | head -1

echo -e "${GREEN}Backend Test Coverage:${NC}"
go tool cover -func=test_results/backend-coverage.out | grep "total:"

if [ -f "$RESULTS_DIR/api-benchmark.txt" ]; then
  echo -e "${GREEN}API Performance:${NC}"
  grep "Requests per second" "$RESULTS_DIR/api-benchmark.txt"
  grep "Time per request" "$RESULTS_DIR/api-benchmark.txt" | head -1
fi

if [ -f "$RESULTS_DIR/image-benchmark.txt" ]; then
  echo -e "${GREEN}Image Loading Performance:${NC}"
  grep "Requests per second" "$RESULTS_DIR/image-benchmark.txt"
  grep "Time per request" "$RESULTS_DIR/image-benchmark.txt" | head -1
fi

echo
echo -e "${BLUE}All results saved to ${RESULTS_DIR}/${NC}"
echo
