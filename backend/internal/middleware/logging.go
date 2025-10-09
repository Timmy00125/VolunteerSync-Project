package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// RequestIDKey is the context key for request IDs
const RequestIDKey ContextKey = "request_id"

// LoggingMiddleware creates a middleware that logs all HTTP requests
// It adds a unique request ID to the context and logs request details including:
// - Method, path, status code, duration
// - User ID (if authenticated)
// - Request ID
// Note: This middleware ensures no PII (Personally Identifiable Information) is logged
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add request ID to Gin context
		c.Set("request_id", requestID)

		// Add request ID to request context for use in other middleware/handlers
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add request ID to response header for tracing
		c.Header("X-Request-ID", requestID)

		// Start timer
		startTime := time.Now()

		// Get logger with context
		log := logger.Get().WithContext(ctx)

		// Log incoming request
		log.WithField("method", c.Request.Method).
			WithField("path", sanitizePath(c.Request.URL.Path)).
			WithField("ip", c.ClientIP()).
			WithField("user_agent", sanitizeUserAgent(c.Request.UserAgent())).
			Info("Incoming request")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Get status code
		statusCode := c.Writer.Status()

		// Build log with request details
		logEntry := log.WithField("method", c.Request.Method).
			WithField("path", sanitizePath(c.Request.URL.Path)).
			WithField("status", statusCode).
			WithField("duration_ms", duration.Milliseconds()).
			WithField("ip", c.ClientIP())

		// Add user ID if authenticated (already sanitized by logger package)
		if userID := GetUserID(c); userID != "" {
			logEntry = logEntry.WithField("user_id", userID)
		}

		// Add error if present
		if len(c.Errors) > 0 {
			logEntry = logEntry.WithField("error", c.Errors.String())
		}

		// Log response based on status code
		switch {
		case statusCode >= 500:
			logEntry.Error("Request completed with server error")
		case statusCode >= 400:
			logEntry.Warn("Request completed with client error")
		case statusCode >= 300:
			logEntry.Info("Request completed with redirect")
		default:
			logEntry.Info("Request completed successfully")
		}
	}
}

// RequestIDMiddleware is a simpler middleware that only adds request ID to context
// Use this if you want request IDs without full request logging
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add request ID to Gin context
		c.Set("request_id", requestID)

		// Add request ID to request context
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add request ID to response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// GetRequestID extracts the request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	requestID, exists := c.Get("request_id")
	if !exists {
		return ""
	}
	return requestID.(string)
}

// sanitizePath removes any potential PII from URL paths
// It removes query parameters which might contain sensitive data
func sanitizePath(path string) string {
	// Remove query parameters to avoid logging sensitive data
	for i, char := range path {
		if char == '?' {
			return path[:i]
		}
	}
	return path
}

// sanitizeUserAgent truncates user agent strings to reasonable length
// to avoid logging excessively long or potentially problematic user agents
func sanitizeUserAgent(userAgent string) string {
	maxLength := 200
	if len(userAgent) > maxLength {
		return userAgent[:maxLength] + "..."
	}
	return userAgent
}

// SkipLogging is a middleware option to skip logging for specific paths
// Useful for health check endpoints that would clutter logs
func SkipLoggingForPaths(paths ...string) gin.HandlerFunc {
	skipMap := make(map[string]bool)
	for _, path := range paths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		if skipMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Call the regular logging middleware
		LoggingMiddleware()(c)
	}
}

// StructuredLoggingMiddleware provides more detailed structured logging
// with additional fields for advanced monitoring and debugging
func StructuredLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add request ID to contexts
		c.Set("request_id", requestID)
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add request ID to response header
		c.Header("X-Request-ID", requestID)

		// Start timer
		startTime := time.Now()

		// Get logger with context
		log := logger.Get().WithContext(ctx)

		// Log incoming request with more details
		log.WithField("method", c.Request.Method).
			WithField("path", sanitizePath(c.Request.URL.Path)).
			WithField("ip", c.ClientIP()).
			WithField("protocol", c.Request.Proto).
			WithField("content_length", c.Request.ContentLength).
			Debug("Request received")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Get status code and response size
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		// Build structured log entry
		logEntry := log.WithFields(map[string]interface{}{
			"method":         c.Request.Method,
			"path":           sanitizePath(c.Request.URL.Path),
			"status":         statusCode,
			"duration_ms":    duration.Milliseconds(),
			"duration_us":    duration.Microseconds(),
			"ip":             c.ClientIP(),
			"protocol":       c.Request.Proto,
			"response_size":  responseSize,
			"content_length": c.Request.ContentLength,
		})

		// Add user ID if authenticated
		if userID := GetUserID(c); userID != "" {
			logEntry = logEntry.WithField("authenticated", true)
			logEntry = logEntry.WithField("user_id", userID)
		} else {
			logEntry = logEntry.WithField("authenticated", false)
		}

		// Add error details if present
		if len(c.Errors) > 0 {
			logEntry = logEntry.WithField("errors_count", len(c.Errors))
			logEntry = logEntry.WithField("error", c.Errors.String())
		}

		// Log response based on status code
		switch {
		case statusCode >= 500:
			logEntry.Error("Server error")
		case statusCode >= 400:
			logEntry.Warn("Client error")
		case statusCode >= 300:
			logEntry.Info("Redirect")
		case statusCode >= 200:
			logEntry.Info("Success")
		default:
			logEntry.Info("Request completed")
		}
	}
}
