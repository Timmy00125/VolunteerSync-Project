package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationsUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200 for admin, 403 for others
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
		case "admin-token":
			c.Set("user_id", "admin-user-id")
			c.Set("user_type", "admin")
		case "coordinator-token":
			c.Set("user_id", "coordinator-user-id")
			c.Set("user_type", "coordinator")
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

	r.PATCH("/api/v1/organizations/:id", func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists || userType != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin access required",
			})
			return
		}

		orgID := c.Param("id")
		if orgID == "nonexistent-org" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "Organization not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      orgID,
			"updated": true,
		})
	})

	tests := []struct {
		name     string
		orgID    string
		body     map[string]interface{}
		token    string
		wantCode int
	}{
		{
			name:  "successful organization update by admin",
			orgID: "org-uuid-123",
			body: map[string]interface{}{
				"name":              "Updated Organization Name",
				"mission_statement": "Updated mission to serve better",
				"description":       "Updated comprehensive description",
				"website":           "https://updated.org",
				"email":             "updated@org.com",
				"phone":             "+1-555-9999",
			},
			token:    "admin-token",
			wantCode: http.StatusOK,
		},
		{
			name:  "partial organization update by admin",
			orgID: "org-uuid-456",
			body: map[string]interface{}{
				"name": "Partially Updated Org",
			},
			token:    "admin-token",
			wantCode: http.StatusOK,
		},
		{
			name:  "organization update by coordinator (forbidden)",
			orgID: "org-uuid-123",
			body: map[string]interface{}{
				"name": "Should Not Update",
			},
			token:    "coordinator-token",
			wantCode: http.StatusForbidden,
		},
		{
			name:  "organization update by volunteer (forbidden)",
			orgID: "org-uuid-123",
			body: map[string]interface{}{
				"name": "Should Not Update",
			},
			token:    "volunteer-token",
			wantCode: http.StatusForbidden,
		},
		{
			name:  "organization update without authentication",
			orgID: "org-uuid-123",
			body: map[string]interface{}{
				"name": "Should Not Update",
			},
			token:    "",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:  "organization update for nonexistent organization",
			orgID: "nonexistent-org",
			body: map[string]interface{}{
				"name": "Updated Name",
			},
			token:    "admin-token",
			wantCode: http.StatusNotFound,
		},
	}

	// Helper function to send PATCH request and return recorder
	patchOrganization := func(orgID string, body interface{}, token string) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		url := fmt.Sprintf("/api/v1/organizations/%s", orgID)
		req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := patchOrganization(tt.orgID, tt.body, tt.token)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				// For successful updates, validate response structure
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.orgID, resp["id"])
				assert.True(t, resp["updated"].(bool))
			}
		})
	}
}
