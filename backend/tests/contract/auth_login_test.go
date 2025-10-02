package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200 with dummy response
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"access_token":  "dummy.access.token",
			"refresh_token": "dummy.refresh.token",
			"user":          gin.H{"email": "test@example.com"},
		})
	})

	t.Run("Valid credentials", func(t *testing.T) {
		loginReq := map[string]interface{}{
			"email":    "test@example.com",
			"password": "SecurePass123",
		}

		jsonData, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check for tokens
		accessToken, ok := resp["access_token"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, accessToken)

		refreshToken, ok := resp["refresh_token"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, refreshToken)

		// Check user object
		user, ok := resp["user"].(map[string]interface{})
		assert.True(t, ok)
		if ok {
			userEmail, ok := user["email"].(string)
			assert.True(t, ok)
			assert.Equal(t, "test@example.com", userEmail)
		}

		// Test JWT token format and expiry for access token
		token, err := jwt.Parse(accessToken, nil) // No validation key for format check
		assert.NoError(t, err)
		assert.True(t, token.Valid) // Basic validity without key

		claims, ok := token.Claims.(jwt.MapClaims)
		assert.True(t, ok)

		// Check email in claims
		emailClaim, ok := claims["email"].(string)
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", emailClaim)

		// Check expiry (assuming 15 minutes)
		expI, ok := claims["exp"].(float64)
		assert.True(t, ok)
		expTime := time.Unix(int64(expI), 0)
		assert.WithinDuration(t, time.Now().Add(15*time.Minute), expTime, 30*time.Second)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		loginReq := map[string]interface{}{
			"email":    "invalid@example.com",
			"password": "wrongpassword",
		}

		jsonData, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code) // Will fail until implemented
	})

	t.Run("Rate limiting", func(t *testing.T) {
		loginReq := map[string]interface{}{
			"email":    "rate@limit.com",
			"password": "wrongpassword", // Invalid to trigger rate limiting on failures
		}

		jsonData, _ := json.Marshal(loginReq)

		// First 5 attempts should return 401
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, fmt.Sprintf("Attempt %d should return 401", i+1))
		}

		// 6th attempt should return 429
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code) // Will fail until implemented
	})
}
