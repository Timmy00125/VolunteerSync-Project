package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestVolunteersDashboard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that returns volunteer dashboard data
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.GET("/api/v1/volunteers/me/dashboard", func(c *gin.Context) {
		// Check for Authorization header (dummy check)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		// Return volunteer dashboard data (dummy response based on Scenario 4.1)
		c.JSON(http.StatusOK, gin.H{
			"total_hours":             3.0,
			"total_events":            1,
			"organizations_supported": 1,
			"organizations": []gin.H{
				{
					"id":          "550e8400-e29b-41d4-a716-446655440002",
					"name":        "Green Earth Initiative",
					"logo_url":    "https://example.com/logo.jpg",
					"total_hours": 3.0,
				},
			},
			"recent_events": []gin.H{
				{
					"id":         "550e8400-e29b-41d4-a716-446655440003",
					"title":      "Community Garden Cleanup",
					"start_date": "2025-10-15",
					"start_time": "09:00:00",
					"end_time":   "12:00:00",
					"status":     "completed",
					"hours":      3.0,
					"organization": gin.H{
						"id":   "550e8400-e29b-41d4-a716-446655440002",
						"name": "Green Earth Initiative",
					},
				},
			},
			"hours_chart": []gin.H{
				{
					"month": "2025-10",
					"hours": 3.0,
				},
			},
			"upcoming_events": []gin.H{}, // Empty for this test scenario
			"achievements":    []gin.H{}, // Empty for this test scenario
		})
	})

	t.Run("Get dashboard with valid auth", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/volunteers/me/dashboard", nil)
		req.Header.Set("Authorization", "Bearer dummy.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check impact metrics
		assert.Equal(t, 3.0, resp["total_hours"])
		assert.Equal(t, float64(1), resp["total_events"])
		assert.Equal(t, float64(1), resp["organizations_supported"])

		// Check organizations array
		organizations, ok := resp["organizations"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, organizations, 1)

		org := organizations[0].(map[string]interface{})
		assert.Equal(t, "Green Earth Initiative", org["name"])
		assert.Equal(t, 3.0, org["total_hours"])

		// Check recent events
		recentEvents, ok := resp["recent_events"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, recentEvents, 1)

		event := recentEvents[0].(map[string]interface{})
		assert.Equal(t, "Community Garden Cleanup", event["title"])
		assert.Equal(t, "completed", event["status"])
		assert.Equal(t, 3.0, event["hours"])

		// Check hours chart
		hoursChart, ok := resp["hours_chart"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, hoursChart, 1)

		chart := hoursChart[0].(map[string]interface{})
		assert.Equal(t, "2025-10", chart["month"])
		assert.Equal(t, 3.0, chart["hours"])
	})

	t.Run("Get dashboard without auth", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/volunteers/me/dashboard", nil)
		// No Authorization header

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
