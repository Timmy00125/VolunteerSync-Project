package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// RecoveryMiddleware creates a middleware that recovers from panics
// It catches any panics, logs the error with stack trace, and returns a 500 error
// This prevents the entire application from crashing due to a panic in a handler
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get logger with context
				log := logger.Get().WithContext(c.Request.Context())

				// Get stack trace
				stackTrace := string(debug.Stack())

				// Log the panic with stack trace
				log.WithField("panic", fmt.Sprintf("%v", err)).
					WithField("stack_trace", stackTrace).
					WithField("method", c.Request.Method).
					WithField("path", c.Request.URL.Path).
					WithField("ip", c.ClientIP()).
					Error("Panic recovered")

				// Check if the connection is still open
				// If the client has already disconnected, don't try to send a response
				if !c.Writer.Written() {
					// Return 500 Internal Server Error
					errors.AbortWithError(c,
						errors.NewInternalServerError("An unexpected error occurred"),
					)
				}

				// Abort the request chain
				c.Abort()
			}
		}()

		// Process request
		c.Next()
	}
}

// RecoveryWithCustomHandler creates a recovery middleware with a custom handler
// This allows you to customize what happens when a panic is recovered
func RecoveryWithCustomHandler(handler func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get logger with context
				log := logger.Get().WithContext(c.Request.Context())

				// Get stack trace
				stackTrace := string(debug.Stack())

				// Log the panic with stack trace
				log.WithField("panic", fmt.Sprintf("%v", err)).
					WithField("stack_trace", stackTrace).
					WithField("method", c.Request.Method).
					WithField("path", c.Request.URL.Path).
					WithField("ip", c.ClientIP()).
					Error("Panic recovered with custom handler")

				// Call custom handler
				handler(c, err)

				// Abort the request chain
				c.Abort()
			}
		}()

		// Process request
		c.Next()
	}
}

// DetailedRecoveryMiddleware is similar to RecoveryMiddleware but provides more detailed error information
// WARNING: This should only be used in development as it may expose sensitive information
func DetailedRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get logger with context
				log := logger.Get().WithContext(c.Request.Context())

				// Get stack trace
				stackTrace := string(debug.Stack())

				// Log the panic with full details
				log.WithField("panic", fmt.Sprintf("%v", err)).
					WithField("stack_trace", stackTrace).
					WithField("method", c.Request.Method).
					WithField("path", c.Request.URL.Path).
					WithField("query", c.Request.URL.RawQuery).
					WithField("ip", c.ClientIP()).
					WithField("user_agent", c.Request.UserAgent()).
					Error("Panic recovered (detailed)")

				// Check if the connection is still open
				if !c.Writer.Written() {
					// Return detailed error response (for development only)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "internal_server_error",
						"message": "An unexpected error occurred",
						"details": gin.H{
							"panic":       fmt.Sprintf("%v", err),
							"stack_trace": stackTrace,
							"request_id":  GetRequestID(c),
						},
					})
				}

				// Abort the request chain
				c.Abort()
			}
		}()

		// Process request
		c.Next()
	}
}

// ProductionRecoveryMiddleware is a recovery middleware optimized for production
// It logs minimal details and always returns a generic error message
func ProductionRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get logger with context
				log := logger.Get().WithContext(c.Request.Context())

				// Get stack trace
				stackTrace := string(debug.Stack())

				// Log the panic with stack trace (internal use only)
				log.WithField("panic", fmt.Sprintf("%v", err)).
					WithField("stack_trace", stackTrace).
					WithField("method", c.Request.Method).
					WithField("path", c.Request.URL.Path).
					Error("Panic recovered in production")

				// Check if the connection is still open
				if !c.Writer.Written() {
					// Return generic error response (no details exposed)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":      "internal_server_error",
						"message":    "An unexpected error occurred. Please try again later.",
						"request_id": GetRequestID(c),
					})
				}

				// Abort the request chain
				c.Abort()
			}
		}()

		// Process request
		c.Next()
	}
}

// RecoveryWithCallback creates a recovery middleware that also calls a callback function
// This is useful for sending panic information to error tracking services (e.g., Sentry, Bugsnag)
func RecoveryWithCallback(callback func(c *gin.Context, err interface{}, stackTrace string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get logger with context
				log := logger.Get().WithContext(c.Request.Context())

				// Get stack trace
				stackTrace := string(debug.Stack())

				// Log the panic
				log.WithField("panic", fmt.Sprintf("%v", err)).
					WithField("stack_trace", stackTrace).
					WithField("method", c.Request.Method).
					WithField("path", c.Request.URL.Path).
					Error("Panic recovered with callback")

				// Call the callback for external error tracking
				if callback != nil {
					callback(c, err, stackTrace)
				}

				// Check if the connection is still open
				if !c.Writer.Written() {
					// Return 500 error
					errors.AbortWithError(c,
						errors.NewInternalServerError("An unexpected error occurred"),
					)
				}

				// Abort the request chain
				c.Abort()
			}
		}()

		// Process request
		c.Next()
	}
}
