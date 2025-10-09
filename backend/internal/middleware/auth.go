package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/jwt"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// ContextKey is the type for context keys used in middleware
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserRoleKey is the context key for user role
	UserRoleKey ContextKey = "user_role"
	// UserClaimsKey is the context key for full JWT claims
	UserClaimsKey ContextKey = "user_claims"
)

// AuthMiddleware creates a middleware that validates JWT access tokens
// It extracts the token from the Authorization header (Bearer token)
// and validates it using the JWT manager
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Missing Authorization header")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Missing authorization token"))
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn("Invalid Authorization header format")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Invalid authorization header format"))
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			log.Warn("Empty bearer token")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Empty authorization token"))
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			log.WithField("error", err.Error()).Warn("Token validation failed")

			// Map JWT errors to appropriate HTTP errors
			var appErr *errors.AppError
			switch err {
			case jwt.ErrExpiredToken:
				appErr = errors.NewUnauthorizedError("Token has expired")
			case jwt.ErrInvalidSignature:
				appErr = errors.NewUnauthorizedError("Invalid token signature")
			case jwt.ErrInvalidToken, jwt.ErrMalformedToken:
				appErr = errors.NewUnauthorizedError("Invalid token")
			case jwt.ErrInvalidClaims, jwt.ErrMissingUserID, jwt.ErrMissingRole:
				appErr = errors.NewUnauthorizedError("Invalid token claims")
			case jwt.ErrInvalidTokenType:
				appErr = errors.NewUnauthorizedError("Invalid token type")
			default:
				appErr = errors.NewUnauthorizedError("Token validation failed")
			}

			errors.AbortWithError(c, appErr)
			return
		}

		// Add user information to request context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, UserClaimsKey, claims)
		ctx = context.WithValue(ctx, logger.UserIDKey, claims.UserID)

		// Update request with new context
		c.Request = c.Request.WithContext(ctx)

		// Also set in Gin context for easy access
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		log.WithField("user_id", claims.UserID).
			WithField("role", claims.Role).
			Debug("User authenticated successfully")

		c.Next()
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if no token is provided
// This is useful for endpoints that work differently for authenticated vs unauthenticated users
func OptionalAuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided - this is OK for optional auth
			c.Next()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format but optional, so just skip
			log.Debug("Invalid Authorization header format in optional auth")
			c.Next()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			// Empty token but optional, so just skip
			c.Next()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			// Token validation failed but it's optional, so just log and continue
			log.WithField("error", err.Error()).Debug("Optional token validation failed")
			c.Next()
			return
		}

		// Add user information to request context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, UserClaimsKey, claims)
		ctx = context.WithValue(ctx, logger.UserIDKey, claims.UserID)

		// Update request with new context
		c.Request = c.Request.WithContext(ctx)

		// Also set in Gin context for easy access
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		log.WithField("user_id", claims.UserID).
			WithField("role", claims.Role).
			Debug("Optional authentication succeeded")

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context
// Returns empty string if not authenticated
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUserRole extracts the user role from the Gin context
// Returns empty string if not authenticated
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get("user_role")
	if !exists {
		return ""
	}
	return role.(string)
}

// GetUserClaims extracts the full JWT claims from the Gin context
// Returns nil if not authenticated
func GetUserClaims(c *gin.Context) *jwt.Claims {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil
	}
	return claims.(*jwt.Claims)
}

// RequireAuth returns an error if the user is not authenticated
// This is a helper for handlers that need authentication
func RequireAuth(c *gin.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return errors.NewUnauthorizedError("Authentication required")
	}
	return nil
}
