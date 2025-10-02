package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUsersGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200 with dummy user response
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.GET("/api/v1/users/me", func(c *gin.Context) {
		// Check for Authorization header (dummy check)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		// Return dummy user data
		c.JSON(http.StatusOK, gin.H{
			"id":             "550e8400-e29b-41d4-a716-446655440000",
			"email":          "test@example.com",
			"first_name":     "John",
			"last_name":      "Doe",
			"phone":          "+1234567890",
			"account_status": "active",
			"last_login_at":  "2024-01-15T10:30:00Z",
			"created_at":     "2024-01-01T00:00:00Z",
			"updated_at":     "2024-01-15T10:30:00Z",
		})
	})

	t.Run("Authenticated user retrieval", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer dummy.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check required fields
		assert.NotEmpty(t, resp["id"])
		assert.Equal(t, "test@example.com", resp["email"])
		assert.Equal(t, "John", resp["first_name"])
		assert.Equal(t, "Doe", resp["last_name"])
		assert.Equal(t, "+1234567890", resp["phone"])
		assert.Equal(t, "active", resp["account_status"])
		assert.NotEmpty(t, resp["created_at"])
		assert.NotEmpty(t, resp["updated_at"])
	})

	t.Run("Without authentication", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		// No Authorization header

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check error response
		errorMsg, ok := resp["error"].(string)
		assert.True(t, ok)
		assert.Equal(t, "Unauthorized", errorMsg)
	})
}
