# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev && \
    update-ca-certificates

# Copy the entire source code
COPY . .

# Move to backend directory and build
WORKDIR /app/backend

# Install wire, swag and update dependencies
RUN go install github.com/google/wire/cmd/wire@latest && \
    go install github.com/swaggo/swag/cmd/swag@v1.8.12 && \
    go mod download && \
    wire ./di

# Generate Swagger documentation
RUN swag init -g cmd/main.go -o docs

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/valorant-map-picker ./cmd/main.go

# Final stage
FROM alpine:3.18

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/valorant-map-picker /usr/local/bin/valorant-map-picker

# Copy configuration
COPY backend/config/config.yaml /etc/valorant-map-picker/config.yaml

# Create a non-root user to run the application
RUN adduser -D -u 1000 appuser
USER appuser

# Expose the application port
EXPOSE 3000

# Set environment variables
ENV GIN_MODE=release
ENV VALOMAP_SERVER_PORT=3000
ENV VALOMAP_LOGGING_FORMAT=json

# Run the application
ENTRYPOINT ["/usr/local/bin/valorant-map-picker"]
