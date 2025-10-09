# API Contracts

This directory contains API contract specifications for VolunteerSync.

## Files

- **`openapi.yaml`**: OpenAPI 3.0 specification for REST API endpoints

## OpenAPI Specification

The OpenAPI spec defines:

- **Authentication**: JWT-based authentication with refresh tokens
- **Rate Limiting**: 100 req/min per user, 5 login attempts per 15 min per IP
- **Error Handling**: Standardized error responses
- **Request/Response Schemas**: TypeScript-compatible JSON schemas

### API Modules

1. **Authentication** (`/auth/*`): Registration, login, password reset
2. **Users** (`/users/*`): User profile management
3. **Organizations** (`/organizations/*`): Organization CRUD
4. **Volunteers** (`/volunteers/*`): Volunteer profiles
5. **Opportunities** (`/opportunities/*`): Event management and search
6. **Registrations** (`/registrations/*`): Volunteer sign-ups
7. **Hours** (`/hours/*`): Hours tracking and verification
8. **Communications** (`/messages/*`, `/notifications/*`): Messaging system
9. **Analytics** (`/analytics/*`): Reporting dashboards
10. **Documents** (`/documents/*`): Document management
11. **Teams** (`/teams/*`): Team/group features
12. **Admin** (`/admin/*`): Platform administration

## Contract Testing

Contract tests validate that API implementation matches the OpenAPI specification.

**Test Strategy**:

1. Generate contract tests from OpenAPI spec
2. Tests must fail initially (TDD approach)
3. Implementation makes tests pass
4. Tests validate:
   - Request/response schemas
   - Status codes
   - Authentication requirements
   - Error responses

**Tools**:

- Backend: Go `httptest` + manual schema validation or `kin-openapi`
- Frontend: Mock Service Worker (MSW) for mocking based on OpenAPI

## Usage

### View Documentation

Use Swagger UI to view interactive API docs:

```bash
# Install swagger-ui
npm install -g swagger-ui-watcher

# Serve OpenAPI spec
swagger-ui-watcher openapi.yaml
```

### Generate Types

**TypeScript Types** (for frontend):

```bash
npx openapi-typescript openapi.yaml --output ../../../frontend/src/types/api.ts
```

**Go Types** (for backend):

```bash
# Use oapi-codegen or manual implementation
go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
oapi-codegen -package api openapi.yaml > ../../../backend/internal/api/generated.go
```

## Validation

Validate OpenAPI spec:

```bash
npx @apidevtools/swagger-cli validate openapi.yaml
```

## Status

✅ **Phase 1 Complete**: OpenAPI spec structure created with core endpoints (Authentication, Users, Organizations)

🔄 **Next Steps**:

- Expand OpenAPI spec to include all 111 functional requirements
- Generate contract tests for each endpoint
- Implement backend handlers with contract test coverage

---

**Note**: The OpenAPI spec in this directory represents the **contract** for the API. Implementation code should be driven by these contracts following Test-Driven Development (TDD) principles.
