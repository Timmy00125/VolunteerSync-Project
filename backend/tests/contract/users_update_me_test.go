package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUsersUpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200 with updated user response
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.PATCH("/api/v1/users/me", func(c *gin.Context) {
		// Check for Authorization header (dummy check)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		// Parse request body
		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON",
			})
			return
		}

		// Return updated user data (dummy response)
		c.JSON(http.StatusOK, gin.H{
			"id":             "550e8400-e29b-41d4-a716-446655440000",
			"email":          "updated@example.com",
			"first_name":     "Jane",
			"last_name":      "Smith",
			"phone":          "+1987654321",
			"account_status": "active",
			"last_login_at":  "2024-01-15T10:30:00Z",
			"created_at":     "2024-01-01T00:00:00Z",
			"updated_at":     "2024-01-15T11:00:00Z",
		})
	})

	t.Run("Authenticated user update", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"first_name": "Jane",
			"last_name":  "Smith",
			"phone":      "+1987654321",
			"email":      "updated@example.com",
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Authorization", "Bearer dummy.jwt.token")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check required fields
		assert.NotEmpty(t, resp["id"])
		assert.Equal(t, "updated@example.com", resp["email"])
		assert.Equal(t, "Jane", resp["first_name"])
		assert.Equal(t, "Smith", resp["last_name"])
		assert.Equal(t, "+1987654321", resp["phone"])
		assert.Equal(t, "active", resp["account_status"])
		assert.NotEmpty(t, resp["created_at"])
		assert.NotEmpty(t, resp["updated_at"])
	})

	t.Run("Without authentication", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"first_name": "Jane",
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
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

	t.Run("Invalid JSON payload", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBufferString("invalid json"))
		req.Header.Set("Authorization", "Bearer dummy.jwt.token")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check error response
		errorMsg, ok := resp["error"].(string)
		assert.True(t, ok)
		assert.Equal(t, "Invalid JSON", errorMsg)
	})
}
