#!/bin/bash

# Exit on error
set -e

# Print commands
set -x

# Create necessary directories
mkdir -p traefik

# Create acme.json file for Let's Encrypt certificates with correct permissions
touch traefik/acme.json
chmod 600 traefik/acme.json

# Create Docker network if it doesn't exist
if ! docker network inspect web >/dev/null 2>&1; then
  docker network create web
fi

# Generate a basic auth password for Traefik dashboard
echo "Creating admin user for Traefik dashboard"
echo "Please enter a password for the Traefik dashboard:"
read -s PASSWORD

# Generate hashed password using htpasswd (from apache2-utils)
HASHED_PASSWORD=$(docker run --rm httpd:alpine htpasswd -nbB admin "$PASSWORD" | cut -d ":" -f 2)

# Update the docker-compose.yml file with the hashed password
sed -i "s|\$apr1\$70hN10X7\$3QbzMaVnA3pagO1OJl1o90|$HASHED_PASSWORD|g" docker-compose.yml

# Ask for email for Let's Encrypt
echo "Please enter your email address for Let's Encrypt notifications:"
read EMAIL

# Update the traefik.yml file with the email
sed -i "s|your-email@example.com|$EMAIL|g" traefik/traefik.yml

echo "Setup completed successfully!"
echo "You can now run 'docker-compose up -d' to start the services."
