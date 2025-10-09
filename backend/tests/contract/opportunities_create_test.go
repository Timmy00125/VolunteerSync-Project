package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOpportunitiesCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 201
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
		// Mock user context - will be replaced with actual JWT validation
		c.Set("user_id", "mock-user-id")
		c.Set("user_type", "organization_admin")
		c.Set("organization_id", "mock-org-id")
		c.Next()
	})

	r.POST("/api/v1/opportunities", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{})
	})

	tests := []struct {
		name     string
		body     map[string]interface{}
		token    string
		wantCode int
	}{
		{
			name: "valid opportunity creation with required fields only",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"description":    "Help clean up the local beach and protect marine life",
				"start_date":     "2025-12-01",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       20,
			},
			token:    "valid-token",
			wantCode: http.StatusCreated,
		},
		{
			name: "valid opportunity creation with all fields",
			body: map[string]interface{}{
				"title":          "Community Garden Planting Day",
				"description":    "Join us for a day of planting vegetables and flowers in our community garden. All skill levels welcome!",
				"start_date":     "2025-11-15",
				"start_time":     "10:00:00",
				"end_date":       "2025-11-15",
				"end_time":       "14:00:00",
				"timezone":       "America/Los_Angeles",
				"address_line_1": "456 Park Avenue",
				"address_line_2": "Garden Section B",
				"city":           "Portland",
				"state":          "Oregon",
				"postal_code":    "97201",
				"country":        "United States",
				"capacity":       15,
				"min_age":        16,
				"cause_ids":      []string{"cause-uuid-1", "cause-uuid-2"},
				"skill_ids":      []string{"skill-uuid-1"},
				"document_ids":   []string{"doc-uuid-1"},
			},
			token:    "valid-token",
			wantCode: http.StatusCreated,
		},
		{
			name: "recurring opportunity creation",
			body: map[string]interface{}{
				"title":               "Weekly Food Bank Sorting",
				"description":         "Sort and organize donations at the local food bank",
				"start_date":          "2025-10-15",
				"start_time":          "13:00:00",
				"end_date":            "2025-10-15",
				"end_time":            "16:00:00",
				"timezone":            "America/Chicago",
				"address_line_1":      "789 Service Street",
				"city":                "Austin",
				"state":               "Texas",
				"postal_code":         "78701",
				"capacity":            10,
				"is_recurring":        true,
				"recurrence_pattern":  "weekly",
				"recurrence_end_date": "2025-12-31",
			},
			token:    "valid-token",
			wantCode: http.StatusCreated,
		},
		{
			name: "missing required field - title",
			body: map[string]interface{}{
				"description":    "Help clean up the local beach",
				"start_date":     "2025-12-01",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       20,
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing required field - description",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"start_date":     "2025-12-01",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       20,
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing required field - start_date",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"description":    "Help clean up the local beach",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       20,
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing required field - capacity",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"description":    "Help clean up the local beach",
				"start_date":     "2025-12-01",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid date format",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"description":    "Help clean up the local beach",
				"start_date":     "12-01-2025", // Wrong format
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       20,
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid capacity - negative",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"description":    "Help clean up the local beach",
				"start_date":     "2025-12-01",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       -5,
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "end date before start date",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"description":    "Help clean up the local beach",
				"start_date":     "2025-12-02",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01", // Before start
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       20,
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing authentication",
			body: map[string]interface{}{
				"title":          "Beach Cleanup Event",
				"description":    "Help clean up the local beach",
				"start_date":     "2025-12-01",
				"start_time":     "09:00:00",
				"end_date":       "2025-12-01",
				"end_time":       "12:00:00",
				"timezone":       "America/New_York",
				"address_line_1": "123 Ocean Drive",
				"city":           "Miami",
				"state":          "Florida",
				"postal_code":    "33101",
				"capacity":       20,
			},
			token:    "",
			wantCode: http.StatusUnauthorized,
		},
	}

	// helper to send POST and return recorder
	postOpportunity := func(body interface{}, token string) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/opportunities", bytes.NewBuffer(bodyBytes))
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
			// Ensure unique titles for tests that include a title to avoid cross-test pollution
			if title, ok := tt.body["title"].(string); ok && title != "" {
				tt.body["title"] = fmt.Sprintf("%s %d", title, time.Now().UnixNano())
			}

			w := postOpportunity(tt.body, tt.token)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusCreated {
				// For successful creation, expect an opportunity object
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// opportunity object should be returned
				opp, oppOk := resp["opportunity"].(map[string]interface{})
				if !oppOk {
					// Some implementations return the opportunity directly
					opp = resp
				}

				// Basic validation of returned opportunity
				if id, hasID := opp["id"].(string); hasID {
					assert.NotEmpty(t, id, "opportunity id should not be empty")
				}
				if status, hasStatus := opp["status"].(string); hasStatus {
					assert.Equal(t, "published", status, "new opportunity should be published by default")
				}
				if publishedAt, hasPublishedAt := opp["published_at"]; hasPublishedAt {
					assert.NotNil(t, publishedAt, "published_at should be set")
				}
			}
		})
	}

	t.Run("rate limiting", func(t *testing.T) {
		body := map[string]interface{}{
			"title":          "Rate Limited Event",
			"description":    "Testing rate limiting",
			"start_date":     "2025-12-01",
			"start_time":     "09:00:00",
			"end_date":       "2025-12-01",
			"end_time":       "12:00:00",
			"timezone":       "America/New_York",
			"address_line_1": "123 Test Street",
			"city":           "Test City",
			"state":          "Test State",
			"postal_code":    "12345",
			"capacity":       10,
		}

		bodyBytes, _ := json.Marshal(body)

		// Send 100 requests - should succeed (rate limit is 100 per minute)
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/opportunities", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code, "Request %d should succeed", i+1)
		}

		// 101st request - should be rate limited
		req := httptest.NewRequest(http.MethodPost, "/api/v1/opportunities", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code) // Will fail until implemented
	})
}
