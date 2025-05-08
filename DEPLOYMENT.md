# Deployment Guide for Valorant Map Picker

This guide will help you deploy the Valorant Map Picker API with HTTPS support to your server using Docker and Traefik.

## Prerequisites

- A server with Docker and Docker Compose installed
- A domain name pointed to your server (in this case, workspace.chat-mail.uno)
- Basic familiarity with the terminal and Docker

## Deployment Steps

### 1. Clone the Repository

```bash
git clone https://github.com/jungtechou/valomap.git
cd valomap
```

### 2. Prepare the Environment

Make the setup script executable and run it:

```bash
chmod +x setup.sh
./setup.sh
```

This script will:

- Create necessary directories
- Set up acme.json for Let's Encrypt certificates
- Create a Docker network called "web"
- Prompt for a password for the Traefik dashboard
- Ask for your email for Let's Encrypt notifications

### 3. Review the Configuration

Verify the configuration in the following files:

- `docker-compose.yml`: Check service configuration and labels
- `traefik/traefik.yml`: Review Traefik configuration
- `backend/config/config.yaml`: Check API configuration

You may want to adjust the domain names in the `docker-compose.yml` file:

```yaml
- "traefik.http.routers.valomap.rule=Host(`workspace.chat-mail.uno`)"
- "traefik.http.routers.traefik.rule=Host(`traefik.workspace.chat-mail.uno`)"
```

### 4. Deploy the Application

Make the deployment script executable and run it:

```bash
chmod +x deploy.sh
./deploy.sh
```

This will:

- Pull the latest code (if it's a Git repository)
- Build the Docker containers
- Start the services
- Display the logs

### 5. Verify the Deployment

Once deployment is complete, you should be able to access:

- The API at: https://workspace.chat-mail.uno
- The Traefik dashboard at: https://traefik.workspace.chat-mail.uno (password protected)

### 6. Troubleshooting

If you encounter any issues, check the logs:

```bash
docker-compose logs -f api
docker-compose logs -f traefik
```

Common issues and solutions:

1. **Certificate errors**: Ensure your domain is correctly pointed to your server. Let's Encrypt needs to validate your domain.

2. **Port conflicts**: Make sure ports 80 and 443 are available on your server and not used by other services.

3. **Network issues**: Verify that the "web" network is created and that all services are using it.

### 7. Updating the Application

To update the application:

```bash
./deploy.sh
```

This will pull the latest code, rebuild the containers, and restart the services.

## Additional Configuration

### Custom SSL Certificates

If you prefer to use your own SSL certificates instead of Let's Encrypt:

1. Place your certificates in the `traefik/certs` directory:

   - `cert.pem`: Your certificate
   - `key.pem`: Your private key

2. Modify `traefik/traefik.yml` to use these certificates instead of Let's Encrypt.

### Scaling the API

If you need to run multiple instances of the API:

```bash
docker-compose up -d --scale api=3
```

Traefik will automatically load balance between the instances.

## Security Considerations

- The Traefik dashboard is protected with basic authentication, but consider restricting it further.
- Review the Docker container security settings.
- Consider implementing rate limiting for the API endpoints.
- Regularly update the Docker images and dependencies.

## Maintenance

- Regularly backup the `acme.json` file which contains your Let's Encrypt certificates.
- Monitor disk space usage for logs and containers.
- Set up automated health checks to ensure the API remains available.
