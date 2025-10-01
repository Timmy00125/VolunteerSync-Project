# VolunteerSync Backend

Backend API for the VolunteerSync volunteer management platform.

## Tech Stack

- **Language**: Go 1.25
- **Framework**: Gin
- **Database**: PostgreSQL 16 with GORM ORM
- **Cache**: Redis
- **Authentication**: JWT with Argon2 password hashing
- **Architecture**: Modular monolith with Domain-Driven Design (DDD)

## Project Structure

```
backend/
├── cmd/
│   └── api/                    # Application entry point
├── internal/
│   ├── modules/                # Domain modules (auth, users, orgs, volunteers, etc.)
│   ├── middleware/             # HTTP middleware
│   └── pkg/                    # Shared packages (database, cache, jwt, logging, etc.)
├── migrations/                 # Database migrations
├── tests/                      # Integration & unit tests
├── go.mod                      # Go module definition
└── go.sum                      # Dependency checksums
```

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Redis
- Docker & Docker Compose (for local development)

### Installation

```bash
# Install dependencies
go mod download

# Run database migrations
./scripts/migrate.sh up

# Run the application
go run cmd/api/main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run integration tests
go test ./tests/integration/...

# Run contract tests
go test ./tests/contract/...

# Run with coverage
go test ./... -cover
```

## Development

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Minimum 80% test coverage

### Module Structure

Each domain module follows clean architecture:

```
internal/modules/{module}/
├── models/         # Domain entities
├── repositories/   # Data access layer
├── services/       # Business logic
└── handlers/       # HTTP handlers
```

## Environment Variables

Create a `.env` file in the backend directory:

```
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=volunteersync
DB_PASSWORD=your_password
DB_NAME=volunteersync
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your_jwt_secret
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Server
PORT=8080
GIN_MODE=debug

# External Services
GEOCODING_API_URL=https://nominatim.openstreetmap.org
```

## API Documentation

See `/docs/api.md` for detailed API documentation and OpenAPI specification at `/specs/001-build-volunteersync-an/contracts/openapi.yaml`.
