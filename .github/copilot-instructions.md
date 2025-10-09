# VolunteerSync-Project Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-01

## Active Technologies

- **Frontend**: Next.js 15 (App Router), TypeScript, React 19, Tailwind CSS, shadcn/ui
- **Backend**: Go 1.25, Gin framework, modular monolith with Domain-Driven Design
- **Database**: PostgreSQL 16 with GORM ORM, Redis for caching/sessions
- **Testing**: Jest, React Testing Library, Playwright (E2E), Go testing, Testify
- **Infrastructure**: Docker, Docker Compose, GitHub Actions CI/CD
- **Authentication**: JWT with refresh token rotation, Argon2 password hashing, RBAC

## Project Structure

```
backend/
├── cmd/api/                    # Application entry point
├── internal/modules/           # Domain modules (auth, users, orgs, volunteers, etc.)
├── internal/middleware/        # HTTP middleware
├── migrations/                 # Database migrations
└── tests/                      # Integration & unit tests

frontend/
├── src/app/                    # Next.js App Router pages
├── src/components/             # React components (ui, features, shared)
├── src/lib/                    # Utilities, API clients, hooks
└── tests/                      # Jest, React Testing Library, Playwright tests

specs/001-build-volunteersync-an/
├── spec.md                     # Feature specification
├── plan.md                     # Implementation plan
├── research.md                 # Technology decisions
├── data-model.md               # Database schema
├── contracts/                  # OpenAPI specifications
└── quickstart.md               # Integration test scenarios
```

## Commands

```bash
# Start local development environment
docker compose up

# Backend (Go)
cd backend
go run cmd/api/main.go          # Run backend API
go test ./...                   # Run all tests
go test ./tests/integration/... # Run integration tests

# Frontend (Next.js)
cd frontend
npm run dev                     # Run dev server (hot reload)
npm test                        # Run Jest unit tests
npm run test:e2e               # Run Playwright E2E tests
npm run build                  # Build for production

# Database migrations
cd backend
migrate -path migrations -database "postgresql://..." up
```

## Code Style

- **Go**: Follow standard Go conventions, use gofmt, golangci-lint
- **TypeScript**: ESLint + Prettier, strict TypeScript mode
- **Testing**: TDD approach, minimum 80% code coverage
- **Architecture**: Modular monolith with clean architecture (handlers → services → repositories)
- **Security**: Never log PII, use parameterized queries, validate all inputs

## Recent Changes

- 001-build-volunteersync-an: Initial platform implementation (volunteer management system)
