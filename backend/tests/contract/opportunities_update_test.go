package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOpportunitiesUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200 for coordinator, 403 for others
	// This will be replaced with the actual handler during implementation
	r := gin.New()

	// Mock authentication middleware (will be replaced with actual middleware)
	r.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		switch token {
		case "coordinator-token":
			c.Set("user_id", "coordinator-user-id")
			c.Set("user_type", "organization_admin")
			c.Set("organization_id", "org-uuid-123")
		case "volunteer-token":
			c.Set("user_id", "volunteer-user-id")
			c.Set("user_type", "volunteer")
		default:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	r.PATCH("/api/v1/opportunities/:id", func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "organization_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Coordinator access required",
			})
			return
		}

		oppID := c.Param("id")
		if oppID == "nonexistent-opp" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "Opportunity not found",
			})
			return
		}

		// Mock check for past events - if opportunity ID contains "past", it's a past event
		if strings.Contains(oppID, "past") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "cannot_edit_past_event",
				"message": "Cannot edit past events",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      oppID,
			"updated": true,
		})
	})

	tests := []struct {
		name     string
		oppID    string
		body     map[string]interface{}
		token    string
		wantCode int
	}{
		{
			name:  "successful opportunity update by coordinator",
			oppID: "opp-uuid-123",
			body: map[string]interface{}{
				"title":       "Updated Beach Cleanup Event",
				"description": "Updated description for beach cleanup",
				"capacity":    25,
			},
			token:    "coordinator-token",
			wantCode: http.StatusOK,
		},
		{
			name:  "cannot edit past event",
			oppID: "past-event-uuid",
			body: map[string]interface{}{
				"title": "Updated Past Event",
			},
			token:    "coordinator-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name:  "volunteer cannot update opportunity",
			oppID: "opp-uuid-123",
			body: map[string]interface{}{
				"title": "Volunteer Update Attempt",
			},
			token:    "volunteer-token",
			wantCode: http.StatusForbidden,
		},
		{
			name:     "opportunity not found",
			oppID:    "nonexistent-opp",
			body:     map[string]interface{}{},
			token:    "coordinator-token",
			wantCode: http.StatusNotFound,
		},
		{
			name:  "unauthorized access",
			oppID: "opp-uuid-123",
			body: map[string]interface{}{
				"title": "Unauthorized Update",
			},
			token:    "invalid-token",
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPatch, "/api/v1/opportunities/"+tt.oppID, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.oppID, response["id"])
				assert.True(t, response["updated"].(bool))
			}
		})
	}
}
