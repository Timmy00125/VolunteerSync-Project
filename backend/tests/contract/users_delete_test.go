package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUsersDeleteMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200 with dummy response
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.DELETE("/api/v1/users/me/delete", func(c *gin.Context) {
		// Check for Authorization header (dummy check)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		// Return dummy success response
		c.JSON(http.StatusOK, gin.H{
			"message": "Account deletion requested",
		})
	})

	t.Run("Authenticated user deletion request", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/me/delete", nil)
		req.Header.Set("Authorization", "Bearer dummy.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check success message
		message, ok := resp["message"].(string)
		assert.True(t, ok)
		assert.Equal(t, "Account deletion requested", message)
	})

	t.Run("Without authentication", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/me/delete", nil)
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
