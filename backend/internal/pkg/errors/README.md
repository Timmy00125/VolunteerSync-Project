# Error Handling Package

This package provides standardized error handling utilities for the VolunteerSync backend API. It ensures consistent error responses across all endpoints, following the OpenAPI specification.

## Features

- **Custom Error Types**: Predefined error constructors for common HTTP status codes
- **Error Wrapping**: Support for wrapping underlying errors with context
- **Consistent API Responses**: All errors follow the OpenAPI Error schema
- **Error Chaining**: Compatible with Go 1.13+ error wrapping (`errors.Is`, `errors.As`)
- **Gin Integration**: Helper functions for Gin framework
- **Type Checking**: Helper functions to check error types

## Error Schema

All errors are returned in the following JSON format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {
    "field": "additional context (optional)"
  }
}
```

## Usage

### Basic Error Creation

```go
import "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"

// Create a 400 Bad Request error
err := errors.NewBadRequestError("Invalid email format")

// Create a 401 Unauthorized error
err := errors.NewUnauthorizedError("Token expired")

// Create a 403 Forbidden error
err := errors.NewForbiddenError("You don't have permission to access this resource")

// Create a 404 Not Found error
err := errors.NewNotFoundError("User")  // Message: "User not found"

// Create a 409 Conflict error
err := errors.NewConflictError("Email already exists")

// Create a 429 Rate Limit error
err := errors.NewRateLimitError("Too many login attempts")

// Create a 500 Internal Server Error
err := errors.NewInternalServerError("Database connection failed")
```

### Validation Errors

```go
// Create a validation error with field-specific details
validationDetails := map[string]interface{}{
    "email": "must be a valid email address",
    "password": "must be at least 8 characters",
}
err := errors.NewValidationError("Validation failed", validationDetails)
```

### Error with Additional Details

```go
err := errors.NewBadRequestError("Invalid request").
    WithDetails("field", "email").
    WithDetails("reason", "domain not allowed")
```

### Wrapping Errors

```go
// Wrap an underlying error with context
dbErr := db.Query(...)
if dbErr != nil {
    err := errors.NewInternalServerError("Failed to fetch user").
        WithError(dbErr)
    return err
}

// Or use WrapError for more control
err := errors.WrapError(
    dbErr,
    "Failed to fetch user",
    http.StatusInternalServerError,
    "database_error",
)
```

### Gin Handler Usage

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
)

func GetUser(c *gin.Context) {
    userID := c.Param("id")

    user, err := userService.GetByID(userID)
    if err != nil {
        // Option 1: Use HandleError (writes response but doesn't abort)
        errors.HandleError(c, err)
        return
    }

    // Option 2: Use AbortWithError (writes response and aborts)
    if !user.IsActive {
        errors.AbortWithError(c, errors.NewForbiddenError("User account is inactive"))
        return
    }

    c.JSON(http.StatusOK, user)
}
```

### Error Type Checking

```go
// Check error type
err := someOperation()
if errors.IsNotFound(err) {
    // Handle not found case
}

if errors.IsUnauthorized(err) {
    // Handle unauthorized case
}

if errors.IsForbidden(err) {
    // Handle forbidden case
}

if errors.IsBadRequest(err) {
    // Handle bad request case
}
```

### Using ErrorResponse for Custom Responses

```go
func CustomHandler(c *gin.Context) {
    err := errors.NewBadRequestError("Invalid input")

    // Get error response as gin.H for custom composition
    response := errors.ErrorResponse(err)
    response["timestamp"] = time.Now()

    c.JSON(err.HTTPStatus, response)
}
```

## Available Error Constructors

| Function                               | HTTP Status | Error Code              |
| -------------------------------------- | ----------- | ----------------------- |
| `NewBadRequestError(message)`          | 400         | `invalid_request`       |
| `NewValidationError(message, details)` | 400         | `validation_error`      |
| `NewUnauthorizedError(message)`        | 401         | `unauthorized`          |
| `NewForbiddenError(message)`           | 403         | `forbidden`             |
| `NewNotFoundError(resource)`           | 404         | `not_found`             |
| `NewConflictError(message)`            | 409         | `conflict`              |
| `NewRateLimitError(message)`           | 429         | `rate_limit_exceeded`   |
| `NewInternalServerError(message)`      | 500         | `internal_server_error` |
| `NewServiceUnavailableError(message)`  | 503         | `service_unavailable`   |

## Best Practices

1. **Use Specific Error Types**: Choose the most appropriate error constructor for your use case
2. **Add Context**: Use `WithDetails()` or `WithError()` to add relevant context
3. **Don't Expose Internal Details**: Never include sensitive information (stack traces, connection strings) in error messages
4. **Log Underlying Errors**: When wrapping errors, log the original error for debugging
5. **Consistent Error Codes**: Stick to the predefined error codes for API consistency
6. **Validation Errors**: Always use `NewValidationError()` with field-specific details for validation failures

## Example: Complete Handler

```go
func CreateUser(c *gin.Context) {
    var req CreateUserRequest

    // Bind request
    if err := c.ShouldBindJSON(&req); err != nil {
        errors.AbortWithError(c, errors.NewBadRequestError("Invalid request body"))
        return
    }

    // Validate request
    if validationErrs := validate(req); len(validationErrs) > 0 {
        errors.AbortWithError(c, errors.NewValidationError("Validation failed", validationErrs))
        return
    }

    // Check authorization
    if !hasPermission(c) {
        errors.AbortWithError(c, errors.NewForbiddenError("Insufficient permissions"))
        return
    }

    // Create user
    user, err := userService.Create(req)
    if err != nil {
        // Check if it's a conflict (duplicate email)
        if isDuplicateError(err) {
            errors.AbortWithError(c, errors.NewConflictError("Email already exists"))
            return
        }

        // Generic error - log details but return generic message
        log.Error().Err(err).Msg("Failed to create user")
        errors.AbortWithError(c, errors.NewInternalServerError("Failed to create user"))
        return
    }

    c.JSON(http.StatusCreated, user)
}
```

## Testing

The package includes comprehensive tests covering:

- Error creation and formatting
- Error wrapping and unwrapping
- Gin handler integration
- Error type checking
- Concurrent operations

Run tests with:

```bash
go test ./internal/pkg/errors/...
```
