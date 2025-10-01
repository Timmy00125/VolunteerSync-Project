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

### Linting and Formatting

This project uses `golangci-lint` for code quality checks and `gofmt` for formatting.

```bash
# Format code
go fmt ./...
goimports -w .

# Run linter
./scripts/lint.sh
# or directly:
golangci-lint run --config .golangci.yml ./...

# Install golangci-lint
# macOS
brew install golangci-lint

# Linux/WSL
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or see: https://golangci-lint.run/usage/install/
```

**Pre-commit Hooks** (recommended):

```bash
# Install pre-commit
pip install pre-commit

# Setup hooks
cd backend
pre-commit install

# Run manually on all files
pre-commit run --all-files
```

The pre-commit hooks will automatically:

- Format code with `gofmt` and `goimports`
- Run `golangci-lint` to catch issues
- Check for common problems (trailing whitespace, merge conflicts, etc.)
- Detect secrets and sensitive data

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
