# Logger Package

A structured logging package for the VolunteerSync backend built on top of [zerolog](https://github.com/rs/zerolog).

## Features

- **Structured Logging**: JSON or console format for easy parsing and human readability
- **Contextual Logging**: Automatically extract request IDs, user IDs, and organization IDs from context
- **PII Protection**: Built-in filtering to ensure no Personally Identifiable Information is logged
- **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal
- **Helper Functions**: Specialized logging for HTTP requests, database queries, cache operations, authentication, and authorization events
- **Thread-Safe**: Safe for concurrent use across goroutines

## Installation

The logger package is already included in the internal packages. Import it:

```go
import "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
```

## Quick Start

### Initialize the Logger

Initialize the logger at application startup (typically in `main.go`):

```go
import "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"

func main() {
    // Initialize with JSON format for production
    logger.Init(logger.Config{
        Level:      "info",
        Format:     "json",
        WithCaller: false,
    })

    // Or console format for development
    logger.Init(logger.Config{
        Level:      "debug",
        Format:     "console",
        WithCaller: true,
    })
}
```

### Basic Logging

```go
log := logger.Get()

log.Debug("This is a debug message")
log.Info("This is an info message")
log.Warn("This is a warning message")
log.Error("This is an error message")

// Formatted messages
log.Infof("User %s logged in", userID)
log.Errorf("Failed to connect to database: %v", err)

// Error with error object
log.ErrorWithErr("Database operation failed", err)
```

### Contextual Logging

Add context values to automatically include them in logs:

```go
import "context"

// In middleware or handler
ctx := context.Background()
ctx = context.WithValue(ctx, logger.RequestIDKey, "req-12345")
ctx = context.WithValue(ctx, logger.UserIDKey, "user-67890")
ctx = context.WithValue(ctx, logger.OrgIDKey, "org-abcde")

// Create logger with context
log := logger.Get().WithContext(ctx)
log.Info("Processing request")

// Output: {"level":"info","request_id":"req-12345","user_id":"user-67890","org_id":"org-abcde","time":"...","message":"Processing request"}
```

### Adding Custom Fields

```go
log := logger.Get()

// Single field
log.WithField("operation", "user_registration").Info("User registered successfully")

// Multiple fields
fields := map[string]interface{}{
    "operation": "database_query",
    "table":     "users",
    "rows":      42,
}
log.WithFields(fields).Info("Query executed")
```

### Specialized Logging Functions

#### HTTP Requests

```go
log := logger.Get()
log.LogRequest("GET", "/api/v1/users", 200, 150*time.Millisecond)

// Output: {"level":"info","method":"GET","path":"/api/v1/users","status":200,"duration_ms":150,"time":"...","message":"HTTP request"}
```

#### Database Queries

```go
log := logger.Get()
log.LogDatabaseQuery("SELECT", "users", 50*time.Millisecond)

// Output: {"level":"debug","operation":"SELECT","table":"users","duration_ms":50,"time":"...","message":"Database query"}
```

#### Cache Operations

```go
log := logger.Get()
log.LogCacheOperation("GET", "user:123", true) // true = cache hit

// Output: {"level":"debug","operation":"GET","key":"user:123","hit":true,"time":"...","message":"Cache operation"}
```

#### Authentication Events

```go
log := logger.Get()
log.LogAuthentication("user-123", "login", true) // true = success

// Output: {"level":"info","user_id":"user-123","action":"login","success":true,"time":"...","message":"Authentication event"}
```

#### Authorization Events

```go
log := logger.Get()
log.LogAuthorization("user-123", "organizations", "create", true) // true = allowed

// Output: {"level":"info","user_id":"user-123","resource":"organizations","action":"create","allowed":true,"time":"...","message":"Authorization event"}
```

## Configuration

### Config Options

```go
type Config struct {
    Level      string    // "debug", "info", "warn", "error", "fatal"
    Format     string    // "json" or "console"
    Output     io.Writer // Custom output writer (defaults to os.Stdout)
    WithCaller bool      // Include file and line number in logs
}
```

### Log Levels

- **debug**: Detailed information for debugging (e.g., database queries, cache hits)
- **info**: General informational messages (e.g., HTTP requests, user actions)
- **warn**: Warning messages (e.g., failed authentication, denied authorization)
- **error**: Error messages (e.g., database errors, external service failures)
- **fatal**: Fatal errors that cause the application to exit

### Format Options

#### JSON Format (Production)

```json
{
  "level": "info",
  "request_id": "req-123",
  "time": "2025-10-03T12:00:00Z",
  "message": "User logged in"
}
```

#### Console Format (Development)

```
12:00:00 INF User logged in request_id=req-123
```

## PII Protection

The logger **automatically filters out** fields that may contain Personally Identifiable Information (PII):

### Blocked Fields

The following field names are automatically filtered and will **never** appear in logs:

- `password`
- `email`
- `phone`
- `ssn`
- `credit_card`
- `address`
- `name`
- `first_name`
- `last_name`
- `ip_address`
- `date_of_birth`
- `dob`
- `emergency_contact`
- `security_question`
- `security_answer`

### Example

```go
fields := map[string]interface{}{
    "operation": "user_login",
    "password":  "secret123",        // Will NOT be logged
    "email":     "user@example.com", // Will NOT be logged
    "user_id":   "user-123",         // Will be logged
}

log := logger.Get().WithFields(fields)
log.Info("User login attempt")

// Output: {"level":"info","operation":"user_login","user_id":"user-123","time":"...","message":"User login attempt"}
// Note: password and email are NOT in the output
```

## Context Keys

The logger provides predefined context keys for common values:

```go
const (
    RequestIDKey ContextKey = "request_id" // HTTP request ID
    UserIDKey    ContextKey = "user_id"    // Authenticated user ID
    OrgIDKey     ContextKey = "org_id"     // Organization ID
)
```

Use these in middleware to propagate values through the request lifecycle:

```go
// In middleware
requestID := uuid.New().String()
ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)

// In handlers
log := logger.Get().WithContext(ctx)
log.Info("Processing user request") // Automatically includes request_id
```

## Best Practices

### 1. Initialize Once at Startup

```go
func main() {
    logger.Init(logger.Config{
        Level:  getEnv("LOG_LEVEL", "info"),
        Format: getEnv("LOG_FORMAT", "json"),
    })
    // ... rest of application
}
```

### 2. Use Contextual Logging in HTTP Handlers

```go
func MyHandler(c *gin.Context) {
    log := logger.Get().WithContext(c.Request.Context())
    log.Info("Handling request")
    // ... handler logic
}
```

### 3. Never Log PII Directly

```go
// BAD - logs email address
log.Infof("User %s logged in", user.Email)

// GOOD - logs user ID
log.Infof("User %s logged in", user.ID)
```

### 4. Use Specialized Functions for Common Patterns

```go
// Instead of this:
log.WithField("method", "GET").
    WithField("path", "/api/users").
    WithField("status", 200).
    Info("Request completed")

// Use this:
log.LogRequest("GET", "/api/users", 200, duration)
```

### 5. Log Errors with Context

```go
// BAD
log.Error("Failed to save user")

// GOOD
log.WithField("user_id", userID).
    ErrorWithErr("Failed to save user", err)
```

### 6. Use Appropriate Log Levels

```go
log.Debug("Cache hit for key: %s", key)           // Debug info
log.Info("User registered successfully")           // Normal operations
log.Warn("Rate limit exceeded")                    // Warning conditions
log.Error("Failed to connect to database")         // Error conditions
log.Fatal("Cannot start server")                   // Fatal errors (exits app)
```

## Testing

The logger can be configured with a custom output writer for testing:

```go
func TestMyFunction(t *testing.T) {
    buf := &bytes.Buffer{}
    logger.Init(logger.Config{
        Level:  "debug",
        Format: "json",
        Output: buf,
    })

    // Run your code that logs
    MyFunction()

    // Assert on log output
    output := buf.String()
    assert.Contains(t, output, "expected message")
}
```

## Performance Considerations

- **Disabled Log Levels**: Logs below the configured level have minimal overhead
- **Structured Fields**: Use structured fields instead of formatted strings for better performance
- **JSON Format**: Faster parsing and indexing in log aggregation systems
- **No PII Filtering Overhead**: Field filtering only applies when using `WithFields()`

## Integration Examples

### With Gin Framework

```go
func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // Add request ID to context
        requestID := uuid.New().String()
        ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
        c.Request = c.Request.WithContext(ctx)

        // Process request
        c.Next()

        // Log request
        log := logger.Get().WithContext(ctx)
        log.LogRequest(
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            time.Since(start),
        )
    }
}
```

### With GORM

```go
type GormLogger struct {
    logger *logger.Logger
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
    return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    l.logger.WithContext(ctx).Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    l.logger.WithContext(ctx).Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    l.logger.WithContext(ctx).Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()

    log := l.logger.WithContext(ctx).
        WithField("duration_ms", elapsed.Milliseconds()).
        WithField("rows", rows)

    if err != nil {
        log.ErrorWithErr("Database query failed", err)
    } else {
        log.Debug("Database query executed")
    }
}
```

## Troubleshooting

### Logs Not Appearing

Check that the log level is set correctly:

```go
// If you're using Debug() but Level is "info", debug logs won't appear
logger.Init(logger.Config{
    Level: "debug", // Set to debug to see all logs
})
```

### PII Appearing in Logs

Ensure you're using `WithFields()` for dynamic fields, which automatically filters PII:

```go
// This bypasses PII filtering:
log.Infof("User %s with email %s", userID, email) // BAD

// This filters PII:
log.WithFields(map[string]interface{}{
    "user_id": userID,
    "email":   email, // Will be filtered out
}).Info("User action") // GOOD
```

### Context Values Not Appearing

Ensure you're using the correct context keys and `WithContext()`:

```go
ctx = context.WithValue(ctx, logger.RequestIDKey, requestID) // Correct key
log := logger.Get().WithContext(ctx) // Must call WithContext()
```

## License

Internal package for VolunteerSync Project.
