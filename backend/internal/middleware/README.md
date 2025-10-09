# Middleware Package

This package contains HTTP middleware for the VolunteerSync backend API. These middleware components provide essential functionality for authentication, authorization, security, logging, and error handling.

## Middleware Components

### 1. Authentication Middleware (`auth.go`)

**Purpose**: Validates JWT access tokens and adds user information to the request context.

**Features**:

- Validates Bearer tokens from Authorization header
- Extracts user ID and role from JWT claims
- Adds user information to both Gin and request contexts
- Supports optional authentication for public endpoints

**Usage**:

```go
// Required authentication
router.Use(middleware.AuthMiddleware(jwtManager))

// Optional authentication
router.Use(middleware.OptionalAuthMiddleware(jwtManager))
```

**Helper Functions**:

- `GetUserID(c *gin.Context)` - Extract user ID from context
- `GetUserRole(c *gin.Context)` - Extract user role from context
- `GetUserClaims(c *gin.Context)` - Extract full JWT claims
- `RequireAuth(c *gin.Context)` - Verify user is authenticated

---

### 2. RBAC Middleware (`rbac.go`)

**Purpose**: Implements role-based access control and resource ownership verification.

**Roles**:

- `super_admin` - Full system access
- `org_admin` - Organization administration
- `coordinator` - Event coordination
- `volunteer` - Basic volunteer access

**Features**:

- Role-based access control
- Organization membership verification
- Resource ownership checks
- Hierarchical permissions (super admin > org admin > coordinator > volunteer)

**Usage**:

```go
// Require specific role
router.Use(middleware.RequireRole(middleware.RoleSuperAdmin))

// Convenience functions
router.Use(middleware.RequireSuperAdmin())
router.Use(middleware.RequireOrgAdmin())
router.Use(middleware.RequireCoordinator())
router.Use(middleware.RequireStaff())

// Verify organization membership
router.Use(middleware.RequireOrgMembership("org_id"))

// Verify resource ownership
router.Use(middleware.RequireResourceOwnership("user_id"))
```

**Helper Functions**:

- `IsSuperAdmin(c)`, `IsOrgAdmin(c)`, `IsCoordinator(c)`, `IsVolunteer(c)`, `IsStaff(c)`
- `GetOrgID(c *gin.Context)` - Extract organization ID from context

---

### 3. Rate Limiting Middleware (`rate_limit.go`)

**Purpose**: Protects API from abuse by limiting request rates using Redis.

**Configurations**:

- **General**: 100 requests per minute per user
- **Login**: 5 attempts per 15 minutes per IP address

**Features**:

- Per-user rate limiting (when authenticated)
- Per-IP rate limiting (when not authenticated)
- Configurable limits and time windows
- Rate limit headers in responses
- Graceful degradation if Redis is unavailable

**Usage**:

```go
// General rate limiting
router.Use(middleware.RateLimitMiddleware(redisClient, nil))

// Custom configuration
config := &middleware.RateLimitConfig{
    MaxRequests: 50,
    Window:      time.Minute,
    KeyPrefix:   "rate_limit:custom:",
}
router.Use(middleware.RateLimitMiddleware(redisClient, config))

// Login rate limiting
loginRouter.Use(middleware.IPRateLimitMiddleware(redisClient, middleware.LoginRateLimitConfig()))
```

**Response Headers**:

- `X-RateLimit-Limit` - Maximum requests allowed
- `X-RateLimit-Remaining` - Requests remaining in window
- `X-RateLimit-Reset` - Unix timestamp when limit resets
- `Retry-After` - Seconds to wait (when limit exceeded)

---

### 4. CORS Middleware (`cors.go`)

**Purpose**: Handles Cross-Origin Resource Sharing (CORS) for frontend-backend communication.

**Features**:

- Configurable allowed origins, methods, and headers
- Preflight request handling
- Credentials support
- Exposed response headers

**Usage**:

```go
// Development (default config)
router.Use(middleware.CORSMiddleware(nil))

// Production
router.Use(middleware.CORSMiddleware(middleware.ProductionCORSConfig()))

// Allow all origins (development only)
router.Use(middleware.AllowAllOriginsMiddleware())

// Custom configuration
config := &middleware.CORSConfig{
    AllowedOrigins:   []string{"https://app.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge:          86400,
}
router.Use(middleware.CORSMiddleware(config))
```

---

### 5. Logging Middleware (`logging.go`)

**Purpose**: Logs all HTTP requests with detailed information while ensuring no PII is logged.

**Features**:

- Request ID generation and tracking
- Request/response logging
- Duration tracking
- PII sanitization
- Structured logging with context
- Configurable log levels based on status codes

**Usage**:

```go
// Standard logging
router.Use(middleware.LoggingMiddleware())

// Request ID only (no logging)
router.Use(middleware.RequestIDMiddleware())

// Skip logging for specific paths (e.g., health checks)
router.Use(middleware.SkipLoggingForPaths("/health", "/metrics"))

// Structured logging with more details
router.Use(middleware.StructuredLoggingMiddleware())
```

**Helper Functions**:

- `GetRequestID(c *gin.Context)` - Extract request ID from context

**Logged Information**:

- Method, path, status code, duration
- User ID (if authenticated)
- Client IP, request ID
- Error messages (if any)
- NO PII (emails, passwords, tokens, etc.)

---

### 6. Recovery Middleware (`recovery.go`)

**Purpose**: Catches panics to prevent application crashes and provides graceful error handling.

**Features**:

- Panic recovery
- Stack trace logging
- 500 error responses
- Multiple recovery strategies
- Custom error handling support

**Usage**:

```go
// Standard recovery (production)
router.Use(middleware.RecoveryMiddleware())

// Production-optimized (minimal error exposure)
router.Use(middleware.ProductionRecoveryMiddleware())

// Development (detailed error information)
router.Use(middleware.DetailedRecoveryMiddleware())

// Custom handler
router.Use(middleware.RecoveryWithCustomHandler(func(c *gin.Context, err interface{}) {
    // Custom error handling logic
}))

// With callback for error tracking services
router.Use(middleware.RecoveryWithCallback(func(c *gin.Context, err interface{}, stackTrace string) {
    // Send to Sentry, Bugsnag, etc.
}))
```

---

## Middleware Chain Order

The recommended middleware chain order (as specified in tasks.md):

```go
router := gin.New()

// 1. Recovery - Must be first to catch panics from other middleware
router.Use(middleware.RecoveryMiddleware())

// 2. Logging - Should be early to log all requests
router.Use(middleware.LoggingMiddleware())

// 3. CORS - Handle CORS before other processing
router.Use(middleware.CORSMiddleware(corsConfig))

// 4. Rate Limiting - Protect before authentication
router.Use(middleware.RateLimitMiddleware(redisClient, nil))

// 5. Authentication - Parse and validate tokens
router.Use(middleware.AuthMiddleware(jwtManager))

// 6. RBAC - Check permissions after authentication
router.Use(middleware.RequireRole("coordinator", "org_admin", "super_admin"))
```

## Security Considerations

1. **No PII in Logs**: All middleware ensures no personally identifiable information is logged
2. **Token Security**: JWT tokens are validated and expired tokens are rejected
3. **Rate Limiting**: Protects against brute force and DDoS attacks
4. **CORS**: Restricts cross-origin access to trusted domains
5. **Panic Recovery**: Prevents information leakage through stack traces in production
6. **Error Handling**: Returns generic error messages to clients, detailed logs internally

## Testing

Each middleware component should be tested with:

- Unit tests for individual functions
- Integration tests for middleware chains
- Security tests for bypass attempts
- Performance tests for overhead measurement

## Dependencies

- **JWT Manager**: `internal/pkg/jwt` - Token validation
- **Logger**: `internal/pkg/logger` - Structured logging
- **Error Handler**: `internal/pkg/errors` - Standardized error responses
- **Redis Client**: `internal/pkg/cache` - Rate limiting and session storage
- **Gin Framework**: HTTP routing and middleware support

## Environment Variables

Middleware can be configured via environment variables:

- `CORS_ALLOWED_ORIGINS` - Comma-separated list of allowed origins
- `RATE_LIMIT_MAX_REQUESTS` - Maximum requests per window
- `RATE_LIMIT_WINDOW` - Time window for rate limiting
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

## Notes

- All middleware is designed to be composable and reusable
- Middleware can be applied globally or to specific route groups
- Performance overhead is minimized through efficient implementations
- Redis fallback ensures availability even when cache is down
