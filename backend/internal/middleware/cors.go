package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// CORSConfig holds the configuration for CORS middleware
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins (e.g., ["http://localhost:3000", "https://app.volunteersync.com"])
	// Use ["*"] to allow all origins (not recommended for production)
	AllowedOrigins []string

	// AllowedMethods is a list of allowed HTTP methods
	AllowedMethods []string

	// AllowedHeaders is a list of allowed HTTP headers
	AllowedHeaders []string

	// ExposedHeaders is a list of headers that are safe to expose to the API client
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include credentials (cookies, HTTP authentication)
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int
}

// DefaultCORSConfig returns the default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Retry-After",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig returns CORS configuration for production
// This should be updated with actual production domain
func ProductionCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{
			"https://volunteersync.com",
			"https://app.volunteersync.com",
			"https://www.volunteersync.com",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Retry-After",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] == "*" {
			// Allow all origins
			allowed = true
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			// Check if origin is in allowed list
			for _, allowedOrigin := range config.AllowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// If origin is not allowed and it's not a preflight request, continue without CORS headers
		if !allowed && c.Request.Method != http.MethodOptions {
			log.WithField("origin", origin).Debug("Origin not in allowed list")
			c.Next()
			return
		}

		// Set CORS headers
		if allowed {
			// Allow credentials if configured
			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			// Set allowed methods
			if len(config.AllowedMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			}

			// Set allowed headers
			if len(config.AllowedHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			}

			// Set exposed headers
			if len(config.ExposedHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			// Set max age for preflight requests
			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == http.MethodOptions {
			log.WithField("origin", origin).Debug("Handling preflight request")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		log.WithField("origin", origin).Debug("CORS headers set")
		c.Next()
	}
}

// AllowAllOriginsMiddleware is a convenience function that allows all origins
// WARNING: This should only be used for development/testing
func AllowAllOriginsMiddleware() gin.HandlerFunc {
	config := &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Retry-After",
		},
		AllowCredentials: false, // Cannot be true when AllowedOrigins is "*"
		MaxAge:           86400,
	}

	return CORSMiddleware(config)
}
