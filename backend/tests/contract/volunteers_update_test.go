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

func TestVolunteersUpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200 with updated volunteer profile response
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.PATCH("/api/v1/volunteers/me", func(c *gin.Context) {
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

		// Return updated volunteer profile data (dummy response)
		c.JSON(http.StatusOK, gin.H{
			"id":                         "550e8400-e29b-41d4-a716-446655440001",
			"user_id":                    "550e8400-e29b-41d4-a716-446655440000",
			"profile_photo_url":          "https://example.com/photo.jpg",
			"biography":                  "Updated volunteer biography",
			"location":                   "San Francisco, CA",
			"latitude":                   37.7749,
			"longitude":                  -122.4194,
			"availability_monday":        true,
			"availability_tuesday":       false,
			"availability_wednesday":     true,
			"availability_thursday":      false,
			"availability_friday":        true,
			"availability_saturday":      true,
			"availability_sunday":        false,
			"preferred_time":             "morning",
			"total_hours":                25.5,
			"total_events":               5,
			"emergency_contact_name":     "Jane Smith",
			"emergency_contact_phone":    "+1987654321",
			"privacy_show_hours":         true,
			"privacy_show_events":        false,
			"privacy_show_organizations": true,
			"notification_in_app":        true,
			"notification_browser_push":  false,
			"created_at":                 "2024-01-01T00:00:00Z",
			"updated_at":                 "2024-01-15T11:00:00Z",
		})
	})

	t.Run("Authenticated volunteer profile update", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"biography":                 "Updated volunteer biography",
			"location":                  "San Francisco, CA",
			"availability_monday":       true,
			"availability_wednesday":    true,
			"availability_friday":       true,
			"availability_saturday":     true,
			"preferred_time":            "morning",
			"emergency_contact_name":    "Jane Smith",
			"emergency_contact_phone":   "+1987654321",
			"privacy_show_events":       false,
			"notification_browser_push": false,
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/volunteers/me", bytes.NewBuffer(payloadBytes))
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
		assert.NotEmpty(t, resp["user_id"])
		assert.Equal(t, "Updated volunteer biography", resp["biography"])
		assert.Equal(t, "San Francisco, CA", resp["location"])
		assert.Equal(t, true, resp["availability_monday"])
		assert.Equal(t, false, resp["availability_tuesday"])
		assert.Equal(t, true, resp["availability_wednesday"])
		assert.Equal(t, false, resp["availability_thursday"])
		assert.Equal(t, true, resp["availability_friday"])
		assert.Equal(t, true, resp["availability_saturday"])
		assert.Equal(t, false, resp["availability_sunday"])
		assert.Equal(t, "morning", resp["preferred_time"])
		assert.Equal(t, "Jane Smith", resp["emergency_contact_name"])
		assert.Equal(t, "+1987654321", resp["emergency_contact_phone"])
		assert.Equal(t, true, resp["privacy_show_hours"])
		assert.Equal(t, false, resp["privacy_show_events"])
		assert.Equal(t, true, resp["privacy_show_organizations"])
		assert.Equal(t, true, resp["notification_in_app"])
		assert.Equal(t, false, resp["notification_browser_push"])
		assert.NotEmpty(t, resp["created_at"])
		assert.NotEmpty(t, resp["updated_at"])
	})

	t.Run("Without authentication", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"biography": "Updated bio",
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/volunteers/me", bytes.NewBuffer(payloadBytes))
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
		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/volunteers/me", bytes.NewBufferString("invalid json"))
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
