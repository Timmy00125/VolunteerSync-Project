package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a custom application error with additional context.
// It follows the Error schema from OpenAPI specification.
type AppError struct {
	// Code is the error code/type (e.g., "invalid_request", "unauthorized")
	Code string `json:"error"`

	// Message is the human-readable error message
	Message string `json:"message"`

	// Details contains additional context about the error (optional)
	Details map[string]interface{} `json:"details,omitempty"`

	// HTTPStatus is the HTTP status code to return
	HTTPStatus int `json:"-"`

	// Err is the underlying error (for wrapping, not serialized)
	Err error `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (wrapped: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error chain support.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds additional context to the error.
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithError wraps an underlying error.
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// Common error constructors following HTTP status codes and OpenAPI spec.

// NewBadRequestError creates a 400 Bad Request error.
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       "invalid_request",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewValidationError creates a 400 Bad Request error for validation failures.
func NewValidationError(message string, validationDetails map[string]interface{}) *AppError {
	return &AppError{
		Code:       "validation_error",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Details:    validationDetails,
	}
}

// NewUnauthorizedError creates a 401 Unauthorized error.
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       "unauthorized",
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a 403 Forbidden error.
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       "forbidden",
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewNotFoundError creates a 404 Not Found error.
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       "not_found",
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// NewConflictError creates a 409 Conflict error.
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       "conflict",
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewRateLimitError creates a 429 Too Many Requests error.
func NewRateLimitError(message string) *AppError {
	return &AppError{
		Code:       "rate_limit_exceeded",
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// NewInternalServerError creates a 500 Internal Server Error.
func NewInternalServerError(message string) *AppError {
	return &AppError{
		Code:       "internal_server_error",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewServiceUnavailableError creates a 503 Service Unavailable error.
func NewServiceUnavailableError(message string) *AppError {
	return &AppError{
		Code:       "service_unavailable",
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// HandleError is a Gin middleware helper that sends a standardized error response.
// It checks if the error is an AppError, otherwise creates a generic internal server error.
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		// It's our custom error type
		c.JSON(appErr.HTTPStatus, gin.H{
			"error":   appErr.Code,
			"message": appErr.Message,
			"details": appErr.Details,
		})
		return
	}

	// Generic error - don't expose internal details
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal_server_error",
		"message": "An unexpected error occurred",
	})
}

// AbortWithError is a convenience function that calls HandleError and aborts the request.
func AbortWithError(c *gin.Context, err error) {
	HandleError(c, err)
	c.Abort()
}

// ErrorResponse creates a gin.H map from an AppError for manual JSON responses.
// Useful when you need to combine error response with other data.
func ErrorResponse(err *AppError) gin.H {
	response := gin.H{
		"error":   err.Code,
		"message": err.Message,
	}
	if len(err.Details) > 0 {
		response["details"] = err.Details
	}
	return response
}

// WrapError wraps a generic error into an AppError with context.
// This is useful for wrapping database errors, external API errors, etc.
func WrapError(err error, message string, httpStatus int, errorCode string) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		// Already an AppError, return as-is or add context
		return appErr
	}

	return &AppError{
		Code:       errorCode,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusNotFound
	}
	return false
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusUnauthorized
	}
	return false
}

// IsForbidden checks if an error is a forbidden error.
func IsForbidden(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusForbidden
	}
	return false
}

// IsBadRequest checks if an error is a bad request error.
func IsBadRequest(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusBadRequest
	}
	return false
}
