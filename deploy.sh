#!/bin/bash

# Exit on error
set -e

# Print commands
set -x

# Pull the latest code
git pull

# Update Go dependencies and generate code
cd backend
go mod tidy
go generate ./di
cd ..

# Build and deploy the services
COMPOSE_BAKE=true docker compose build
COMPOSE_BAKE=true docker compose up -d

# Check logs for errors (non-blocking)
docker compose logs --tail=100

# Print deployment completion message
echo "Deployment completed successfully!"
echo "The API is now accessible at https://valomap.com"
echo "The Traefik dashboard is available at https://traefik.valomap.com"

# Show container status
docker compose ps
