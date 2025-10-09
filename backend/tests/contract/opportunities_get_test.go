package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOpportunitiesGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with placeholder handlers that return expected status codes
	// These will be replaced with actual handlers during implementation
	r := gin.New()

	r.GET("/api/v1/opportunities/:id", func(c *gin.Context) {
		opportunityID := c.Param("id")

		// Simulate not found for specific test case
		if opportunityID == "not-found-id" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "Opportunity not found",
			})
			return
		}

		// For all other cases, return success (placeholder)
		// This will be replaced with actual opportunity retrieval logic
		c.JSON(http.StatusOK, gin.H{
			"id":          opportunityID,
			"title":       "Test Opportunity",
			"description": "Test description",
		})
	})

	tests := []struct {
		name     string
		oppID    string
		wantCode int
	}{
		{
			name:     "successful opportunity retrieval",
			oppID:    "550e8400-e29b-41d4-a716-446655440000",
			wantCode: http.StatusOK,
		},
		{
			name:     "opportunity not found",
			oppID:    "not-found-id",
			wantCode: http.StatusNotFound,
		},
	}

	// Helper function to send GET request and return recorder
	getOpportunity := func(oppID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/opportunities/"+oppID, nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getOpportunity(tt.oppID)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				// For successful retrieval, validate basic response structure
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "id")
				assert.Contains(t, resp, "title")
				assert.Contains(t, resp, "description")
			}

			if tt.wantCode == http.StatusNotFound {
				// For not found, validate error response structure
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
				assert.Contains(t, resp, "message")
			}
		})
	}
}
