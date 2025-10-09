package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// ContextKey types for storing parsed user data
const (
	// UserUUIDKey is the context key for user ID as uuid.UUID
	UserUUIDKey ContextKey = "user_uuid"
)

// ContextEnrichmentMiddleware converts the string user_id from JWT claims
// into a uuid.UUID and stores it in both request context and Gin context.
// This prevents handlers from having to parse the UUID repeatedly and ensures
// type safety across the application.
//
// This middleware MUST be placed after AuthMiddleware in the chain.
func ContextEnrichmentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Get the user_id string from context (set by AuthMiddleware)
		userIDStr := GetUserID(c)
		if userIDStr == "" {
			// No user ID present - this shouldn't happen after AuthMiddleware,
			// but we'll handle it gracefully
			log.Debug("No user_id found in context, skipping enrichment")
			c.Next()
			return
		}

		// Parse the user ID string into a UUID
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.WithField("user_id", userIDStr).
				WithField("error", err.Error()).
				Error("Failed to parse user_id as UUID")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Invalid user ID format"))
			return
		}

		// Add the parsed UUID to both request context and Gin context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, UserUUIDKey, userUUID)
		c.Request = c.Request.WithContext(ctx)

		// Also set in Gin context for handlers that prefer c.Get()
		c.Set("user_uuid", userUUID)

		log.WithField("user_id", userIDStr).
			WithField("user_uuid", userUUID.String()).
			Debug("Context enriched with UUID")

		c.Next()
	}
}

// GetUserUUID extracts the user UUID from the Gin context
// Returns uuid.Nil if not authenticated or enrichment middleware wasn't run
func GetUserUUID(c *gin.Context) uuid.UUID {
	userUUID, exists := c.Get("user_uuid")
	if !exists {
		return uuid.Nil
	}

	// Type assert to uuid.UUID
	if uuidVal, ok := userUUID.(uuid.UUID); ok {
		return uuidVal
	}

	return uuid.Nil
}

// MustGetUserUUID extracts the user UUID from the Gin context
// Panics if not authenticated - use only in handlers after auth middleware
func MustGetUserUUID(c *gin.Context) uuid.UUID {
	userUUID := GetUserUUID(c)
	if userUUID == uuid.Nil {
		panic("user_uuid not found in context - ensure ContextEnrichmentMiddleware is configured")
	}
	return userUUID
}
