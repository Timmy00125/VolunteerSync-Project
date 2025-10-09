package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextEnrichmentMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successfully enriches context with valid UUID", func(t *testing.T) {
		// Setup
		router := gin.New()
		testUserID := uuid.New()

		// Middleware chain: set user_id -> enrich context
		router.Use(func(c *gin.Context) {
			c.Set("user_id", testUserID.String())
			c.Next()
		})
		router.Use(ContextEnrichmentMiddleware())

		// Test handler that checks the enriched context
		router.GET("/test", func(c *gin.Context) {
			userUUID := GetUserUUID(c)
			assert.Equal(t, testUserID, userUUID)
			c.JSON(http.StatusOK, gin.H{"user_uuid": userUUID.String()})
		})

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("handles missing user_id gracefully", func(t *testing.T) {
		// Setup
		router := gin.New()
		router.Use(ContextEnrichmentMiddleware())

		// Test handler
		router.GET("/test", func(c *gin.Context) {
			userUUID := GetUserUUID(c)
			assert.Equal(t, uuid.Nil, userUUID)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert - should not fail, just skip enrichment
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns error for invalid UUID format", func(t *testing.T) {
		// Setup
		router := gin.New()

		// Middleware chain: set invalid user_id -> enrich context
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "not-a-valid-uuid")
			c.Next()
		})
		router.Use(ContextEnrichmentMiddleware())

		// Test handler (should not be reached)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert - should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GetUserUUID returns Nil for missing context value", func(t *testing.T) {
		// Setup
		router := gin.New()

		// Test handler without enrichment
		router.GET("/test", func(c *gin.Context) {
			userUUID := GetUserUUID(c)
			assert.Equal(t, uuid.Nil, userUUID)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MustGetUserUUID panics when user_uuid not in context", func(t *testing.T) {
		// Setup
		router := gin.New()

		// Test handler
		router.GET("/test", func(c *gin.Context) {
			assert.Panics(t, func() {
				MustGetUserUUID(c)
			})
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MustGetUserUUID returns UUID when present", func(t *testing.T) {
		// Setup
		router := gin.New()
		testUserID := uuid.New()

		// Middleware chain: set user_id -> enrich context
		router.Use(func(c *gin.Context) {
			c.Set("user_id", testUserID.String())
			c.Next()
		})
		router.Use(ContextEnrichmentMiddleware())

		// Test handler
		router.GET("/test", func(c *gin.Context) {
			userUUID := MustGetUserUUID(c)
			require.NotEqual(t, uuid.Nil, userUUID)
			assert.Equal(t, testUserID, userUUID)
			c.JSON(http.StatusOK, gin.H{"user_uuid": userUUID.String()})
		})

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
