#!/bin/bash

set -e

echo "Generating Swagger documentation..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Run swag init with minimal flags
echo "Running swag init from $(pwd)"
swag init -g cmd/main.go -o docs

echo "Swagger documentation generated successfully!"

# How to use:
# 1. Make sure you have the Go compiler installed
# 2. Run this script from the backend directory
# 3. The documentation will be generated in the docs directory
#
# Note: If you encounter CGO errors, you may need to install gcc:
#   - On Ubuntu/Debian: apt-get install gcc
#   - On Alpine: apk add build-base
#   - On CentOS/RHEL: yum install gcc
