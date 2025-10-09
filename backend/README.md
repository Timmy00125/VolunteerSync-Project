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
│   ├── middleware/             # HTTP middleware (auth, CORS, rate limiting, context enrichment)
│   └── pkg/                    # Shared packages (database, cache, jwt, logging, etc.)
├── migrations/                 # Database migrations
├── tests/                      # Integration & unit tests
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── WIRING_IMPLEMENTATION.md    # Latest implementation notes (context enrichment, routing)
└── README.md                   # This file
```

## Recent Changes

**2025-10-09**: Completed module wiring and context enrichment middleware

- ✅ All module handlers registered in the router
- ✅ Context enrichment middleware converts JWT `user_id` string to `uuid.UUID`
- ✅ Middleware chain properly ordered (logging → recovery → CORS → rate limit → auth → enrichment → RBAC)
- ⚠️ Module services need implementation (currently placeholder `nil` services)
- See `WIRING_IMPLEMENTATION.md` for details and migration guide

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose (for local development)

### Local Development (Recommended)

The easiest way to develop is to run PostgreSQL and Redis in Docker, and the backend on your host machine:

```bash
# 1. Start database services (PostgreSQL on port 5433, Redis on port 6380)
make dev-up

# 2. Run the backend (automatically connects to Docker services via .env.local)
make run-local
```

The backend will be available at `http://localhost:8080`.

**Why this approach?**

- Fast hot-reload with Go's native compilation
- Direct access to logs and debugger
- No container rebuilds needed
- Database persists across restarts

### Available Make Commands

```bash
make help              # Show all available commands
make run-local         # Run backend locally (connects to Docker services)
make dev-up            # Start Docker services (postgres, redis)
make dev-down          # Stop Docker services
make dev-logs          # Show Docker service logs
make build             # Build the backend binary
make test              # Run all tests
make test-integration  # Run integration tests only
make test-contract     # Run contract tests only
make lint              # Run linter
make fmt               # Format code
make clean             # Clean build artifacts
```

### Alternative: Full Docker Stack

To run everything in Docker (backend, frontend, databases):

```bash
cd ../docker
docker compose up
```

See `../docker/LOCAL_DEVELOPMENT.md` for detailed port mappings and connection info.

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
