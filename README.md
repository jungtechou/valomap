# Valorant Map Picker

A Go-based API for randomly selecting Valorant maps with customizable bans and preferences.

## Features

- ✅ Random map selection from Valorant's official map pool
- ✅ Map banning/exclusion support
- ✅ RESTful API with Swagger documentation
- ✅ Structured logging and request tracing
- ✅ Configuration via environment variables or YAML file
- ✅ Graceful shutdown and error handling
- ✅ Health check and monitoring endpoints

## Getting Started

### Prerequisites

- Go 1.19 or higher
- Make (optional, for using the Makefile)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/jungtechou/valomap.git
cd valomap
```

2. Install dependencies:

```bash
go mod download
```

3. Run the application:

```bash
go run backend/cmd/main.go
```

### Configuration

Configuration can be provided in several ways:

1. Environment variables (prefixed with `VALOMAP_`)
2. Configuration file (`config.yaml`)
3. Command line flags (if implemented)

Example configuration:

```yaml
server:
  port: 3000
  read_timeout: 10s
  write_timeout: 10s
  shutdown_timeout: 5s

logging:
  level: info
  format: text
  report_caller: false
```

## API Endpoints

### Map Roulette

```
GET /api/v1/map/roulette
```

Returns a randomly selected map from the Valorant map pool.

### Health Check

```
GET /api/v1/health
```

Provides system health information.

### API Documentation

Swagger documentation is available at:

```
GET /swagger/index.html
```

## Development

### Project Structure

```
├── backend/
│   ├── api/            # API layer (handlers, router, engine)
│   ├── cmd/            # Application entrypoints
│   ├── config/         # Configuration management
│   ├── di/             # Dependency injection
│   ├── domain/         # Domain models
│   ├── docs/           # API documentation
│   ├── pkg/            # Reusable packages
│   └── service/        # Business logic
├── frontend/           # (To be implemented)
```

### Dependencies

- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [Wire](https://github.com/google/wire) - Dependency injection
- [Logrus](https://github.com/sirupsen/logrus) - Structured logging
- [Viper](https://github.com/spf13/viper) - Configuration management

### Generating Wire

If you modify the dependency injection setup, regenerate the wire_gen.go file:

```bash
go generate ./backend/di
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
