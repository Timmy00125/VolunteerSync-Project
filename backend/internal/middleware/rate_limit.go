package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/cache"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	// MaxRequests is the maximum number of requests allowed in the time window
	MaxRequests int
	// Window is the time window for rate limiting
	Window time.Duration
	// KeyPrefix is the prefix for Redis keys (e.g., "rate_limit:general:", "rate_limit:login:")
	KeyPrefix string
}

// DefaultRateLimitConfig returns the default rate limit configuration
// 100 requests per minute per user
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		MaxRequests: 100,
		Window:      time.Minute,
		KeyPrefix:   "rate_limit:general:",
	}
}

// LoginRateLimitConfig returns rate limit configuration for login endpoints
// 5 login attempts per 15 minutes per IP
func LoginRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		MaxRequests: 5,
		Window:      15 * time.Minute,
		KeyPrefix:   "rate_limit:login:",
	}
}

// RateLimitMiddleware creates a rate limiting middleware using Redis
// It uses the user ID if authenticated, otherwise falls back to IP address
func RateLimitMiddleware(redisClient *cache.Client, config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Get identifier for rate limiting (user ID if authenticated, otherwise IP)
		identifier := getIdentifier(c)

		// Create Redis key
		key := fmt.Sprintf("%s%s", config.KeyPrefix, identifier)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Check current request count
		count, err := redisClient.Increment(ctx, key)
		if err != nil {
			// If Redis is unavailable, log error but don't block the request
			log.ErrorWithErr("Failed to increment rate limit counter", err)
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			if err := redisClient.Expire(ctx, key, config.Window); err != nil {
				log.ErrorWithErr("Failed to set rate limit expiration", err)
			}
		}

		// Get remaining TTL for rate limit reset
		ttl, err := redisClient.TTL(ctx, key)
		if err != nil {
			log.ErrorWithErr("Failed to get rate limit TTL", err)
			ttl = config.Window // Fallback to window duration
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(config.MaxRequests)-count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		// Check if rate limit exceeded
		if count > int64(config.MaxRequests) {
			log.WithField("identifier", identifier).
				WithField("count", count).
				WithField("max_requests", config.MaxRequests).
				Warn("Rate limit exceeded")

			retryAfter := int(ttl.Seconds())
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			errors.AbortWithError(c,
				errors.NewRateLimitError(
					fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", retryAfter),
				),
			)
			return
		}

		log.WithField("identifier", identifier).
			WithField("count", count).
			WithField("max_requests", config.MaxRequests).
			Debug("Rate limit check passed")

		c.Next()
	}
}

// IPRateLimitMiddleware creates a rate limiting middleware that always uses IP address
// This is useful for public endpoints like registration where there's no user ID
func IPRateLimitMiddleware(redisClient *cache.Client, config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Get IP address
		ip := c.ClientIP()

		// Create Redis key
		key := fmt.Sprintf("%s%s", config.KeyPrefix, ip)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Check current request count
		count, err := redisClient.Increment(ctx, key)
		if err != nil {
			// If Redis is unavailable, log error but don't block the request
			log.ErrorWithErr("Failed to increment IP rate limit counter", err)
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			if err := redisClient.Expire(ctx, key, config.Window); err != nil {
				log.ErrorWithErr("Failed to set IP rate limit expiration", err)
			}
		}

		// Get remaining TTL for rate limit reset
		ttl, err := redisClient.TTL(ctx, key)
		if err != nil {
			log.ErrorWithErr("Failed to get IP rate limit TTL", err)
			ttl = config.Window // Fallback to window duration
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(config.MaxRequests)-count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		// Check if rate limit exceeded
		if count > int64(config.MaxRequests) {
			log.WithField("ip", ip).
				WithField("count", count).
				WithField("max_requests", config.MaxRequests).
				Warn("IP rate limit exceeded")

			retryAfter := int(ttl.Seconds())
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			errors.AbortWithError(c,
				errors.NewRateLimitError(
					fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", retryAfter),
				),
			)
			return
		}

		log.WithField("ip", ip).
			WithField("count", count).
			WithField("max_requests", config.MaxRequests).
			Debug("IP rate limit check passed")

		c.Next()
	}
}

// getIdentifier returns an identifier for rate limiting
// It uses user ID if authenticated, otherwise falls back to IP address
func getIdentifier(c *gin.Context) string {
	// Try to get user ID from context (if authenticated)
	userID := GetUserID(c)
	if userID != "" {
		return fmt.Sprintf("user:%s", userID)
	}

	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// max returns the maximum of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
