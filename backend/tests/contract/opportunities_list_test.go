package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOpportunitiesList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with placeholder handlers that return expected status codes
	// These will be replaced with actual handlers during implementation
	r := gin.New()

	r.GET("/api/v1/opportunities", func(c *gin.Context) {
		// Extract query parameters
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")
		search := c.Query("search")
		status := c.Query("status")
		city := c.Query("city")
		state := c.Query("state")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		organizationId := c.Query("organization_id")

		// Simulate response based on filters
		// This will be replaced with actual database query logic

		// Mock data - example opportunities
		opportunities := []gin.H{
			{
				"id":                    "770e8400-e29b-41d4-a716-446655440000",
				"organization_id":       "550e8400-e29b-41d4-a716-446655440000",
				"created_by_user_id":    "330e8400-e29b-41d4-a716-446655440000",
				"title":                 "Beach Cleanup - Ocean Park",
				"description":           "Help clean up the beautiful Ocean Park beach and protect marine life",
				"status":                "published",
				"start_date":            "2024-10-15",
				"start_time":            "09:00:00",
				"end_date":              "2024-10-15",
				"end_time":              "12:00:00",
				"timezone":              "America/Los_Angeles",
				"address_line_1":        "Ocean Park Beach",
				"address_line_2":        nil,
				"city":                  "Santa Monica",
				"state":                 "CA",
				"postal_code":           "90405",
				"country":               "United States",
				"latitude":              34.0194,
				"longitude":             -118.4912,
				"capacity":              20,
				"current_registrations": 5,
				"min_age":               16,
				"is_recurring":          false,
				"published_at":          "2024-10-01T10:00:00Z",
				"created_at":            "2024-10-01T09:00:00Z",
			},
			{
				"id":                    "880e8400-e29b-41d4-a716-446655440001",
				"organization_id":       "660e8400-e29b-41d4-a716-446655440001",
				"created_by_user_id":    "440e8400-e29b-41d4-a716-446655440001",
				"title":                 "Animal Shelter Volunteer Day",
				"description":           "Spend the day helping care for animals at our shelter",
				"status":                "published",
				"start_date":            "2024-10-20",
				"start_time":            "10:00:00",
				"end_date":              "2024-10-20",
				"end_time":              "16:00:00",
				"timezone":              "America/Los_Angeles",
				"address_line_1":        "123 Shelter Lane",
				"address_line_2":        nil,
				"city":                  "Los Angeles",
				"state":                 "CA",
				"postal_code":           "90001",
				"country":               "United States",
				"latitude":              34.0522,
				"longitude":             -118.2437,
				"capacity":              10,
				"current_registrations": 3,
				"min_age":               18,
				"is_recurring":          false,
				"published_at":          "2024-10-05T14:00:00Z",
				"created_at":            "2024-10-05T13:00:00Z",
			},
			{
				"id":                    "990e8400-e29b-41d4-a716-446655440002",
				"organization_id":       "550e8400-e29b-41d4-a716-446655440000",
				"created_by_user_id":    "330e8400-e29b-41d4-a716-446655440000",
				"title":                 "Food Bank Sorting Event",
				"description":           "Help sort and package food donations for distribution",
				"status":                "draft",
				"start_date":            "2024-10-25",
				"start_time":            "13:00:00",
				"end_date":              "2024-10-25",
				"end_time":              "17:00:00",
				"timezone":              "America/Los_Angeles",
				"address_line_1":        "456 Distribution Way",
				"address_line_2":        "Suite B",
				"city":                  "San Francisco",
				"state":                 "CA",
				"postal_code":           "94102",
				"country":               "United States",
				"latitude":              37.7749,
				"longitude":             -122.4194,
				"capacity":              15,
				"current_registrations": 0,
				"min_age":               nil,
				"is_recurring":          false,
				"published_at":          nil,
				"created_at":            "2024-10-10T11:00:00Z",
			},
		}

		// Apply filters (placeholder logic)
		filteredOpps := []gin.H{}
		for _, opp := range opportunities {
			// Default filter: only return published opportunities unless status filter is explicitly set
			if status == "" {
				oppStatus := opp["status"].(string)
				if oppStatus != "published" {
					continue
				}
			}

			// Filter by search term (title)
			if search != "" {
				title := opp["title"].(string)
				if !containsCaseInsensitive(title, search) {
					continue
				}
			}

			// Filter by status (when explicitly provided)
			if status != "" {
				oppStatus := opp["status"].(string)
				if oppStatus != status {
					continue
				}
			}

			// Filter by city
			if city != "" {
				oppCity := opp["city"].(string)
				if oppCity != city {
					continue
				}
			}

			// Filter by state
			if state != "" {
				oppState := opp["state"].(string)
				if oppState != state {
					continue
				}
			}

			// Filter by organization_id
			if organizationId != "" {
				oppOrgId := opp["organization_id"].(string)
				if oppOrgId != organizationId {
					continue
				}
			}

			// Filter by start_date (basic date filtering)
			if startDate != "" {
				oppStartDate := opp["start_date"].(string)
				if oppStartDate < startDate {
					continue
				}
			}

			// Filter by end_date (basic date filtering)
			if endDate != "" {
				oppEndDate := opp["end_date"].(string)
				if oppEndDate > endDate {
					continue
				}
			}

			filteredOpps = append(filteredOpps, opp)
		}

		// Calculate pagination (placeholder logic)
		totalItems := len(filteredOpps)
		limitInt := 10 // Default
		pageInt := 1   // Default

		// In real implementation, parse page and limit from query params
		_ = page
		_ = limit

		totalPages := (totalItems + limitInt - 1) / limitInt
		hasNext := pageInt < totalPages
		hasPrev := pageInt > 1

		c.JSON(http.StatusOK, gin.H{
			"data": filteredOpps,
			"pagination": gin.H{
				"page":        pageInt,
				"limit":       limitInt,
				"total_pages": totalPages,
				"total_items": totalItems,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		})
	})

	tests := []struct {
		name           string
		queryParams    map[string]string
		wantCode       int
		wantMinResults int
		wantMaxResults int
		validateFunc   func(t *testing.T, resp map[string]interface{})
	}{
		{
			name:           "successful list without filters",
			queryParams:    map[string]string{},
			wantCode:       http.StatusOK,
			wantMinResults: 2, // Should return published opportunities only
			wantMaxResults: 2,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				// Validate pagination exists
				assert.Contains(t, resp, "pagination")
				pagination := resp["pagination"].(map[string]interface{})
				assert.Contains(t, pagination, "page")
				assert.Contains(t, pagination, "limit")
				assert.Contains(t, pagination, "total_pages")
				assert.Contains(t, pagination, "total_items")
				assert.Contains(t, pagination, "has_next")
				assert.Contains(t, pagination, "has_prev")

				// Validate data array exists
				assert.Contains(t, resp, "data")
				data := resp["data"].([]interface{})
				assert.GreaterOrEqual(t, len(data), 2)

				// Validate first opportunity structure
				if len(data) > 0 {
					opp := data[0].(map[string]interface{})
					assert.Contains(t, opp, "id")
					assert.Contains(t, opp, "organization_id")
					assert.Contains(t, opp, "title")
					assert.Contains(t, opp, "description")
					assert.Contains(t, opp, "status")
					assert.Contains(t, opp, "start_date")
					assert.Contains(t, opp, "start_time")
					assert.Contains(t, opp, "end_date")
					assert.Contains(t, opp, "end_time")
					assert.Contains(t, opp, "timezone")
					assert.Contains(t, opp, "city")
					assert.Contains(t, opp, "state")
					assert.Contains(t, opp, "capacity")
					assert.Contains(t, opp, "current_registrations")
					assert.Contains(t, opp, "created_at")

					// Validate status is published (only published should be returned by default)
					status := opp["status"].(string)
					assert.Equal(t, "published", status)
				}
			},
		},
		{
			name: "list with pagination",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "10",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 2,
			wantMaxResults: 10,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				pagination := resp["pagination"].(map[string]interface{})
				assert.Equal(t, float64(1), pagination["page"])
				assert.Equal(t, float64(10), pagination["limit"])
			},
		},
		{
			name: "filter by search title",
			queryParams: map[string]string{
				"search": "Beach Cleanup",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 1,
			wantMaxResults: 1,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				assert.Equal(t, 1, len(data))

				opp := data[0].(map[string]interface{})
				title := opp["title"].(string)
				assert.Contains(t, title, "Beach Cleanup")
			},
		},
		{
			name: "filter by status published",
			queryParams: map[string]string{
				"status": "published",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 2,
			wantMaxResults: 2,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				assert.Equal(t, 2, len(data))

				// All opportunities should be published
				for _, item := range data {
					opp := item.(map[string]interface{})
					status := opp["status"].(string)
					assert.Equal(t, "published", status)
				}
			},
		},
		{
			name: "filter by status draft",
			queryParams: map[string]string{
				"status": "draft",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 1,
			wantMaxResults: 1,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				assert.Equal(t, 1, len(data))

				opp := data[0].(map[string]interface{})
				status := opp["status"].(string)
				assert.Equal(t, "draft", status)
			},
		},
		{
			name: "filter by city",
			queryParams: map[string]string{
				"city": "Los Angeles",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 1,
			wantMaxResults: 1,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				assert.Equal(t, 1, len(data))

				opp := data[0].(map[string]interface{})
				city := opp["city"].(string)
				assert.Equal(t, "Los Angeles", city)
			},
		},
		{
			name: "filter by state",
			queryParams: map[string]string{
				"state": "CA",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 2,
			wantMaxResults: 3,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})

				// All opportunities should be in CA
				for _, item := range data {
					opp := item.(map[string]interface{})
					state := opp["state"].(string)
					assert.Equal(t, "CA", state)
				}
			},
		},
		{
			name: "filter by organization_id",
			queryParams: map[string]string{
				"organization_id": "550e8400-e29b-41d4-a716-446655440000",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 1,
			wantMaxResults: 2,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})

				// All opportunities should belong to the specified organization
				for _, item := range data {
					opp := item.(map[string]interface{})
					orgId := opp["organization_id"].(string)
					assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", orgId)
				}
			},
		},
		{
			name: "filter by date range",
			queryParams: map[string]string{
				"start_date": "2024-10-20",
				"end_date":   "2024-10-25",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 1,
			wantMaxResults: 2,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})

				// All opportunities should be within the date range
				for _, item := range data {
					opp := item.(map[string]interface{})
					startDate := opp["start_date"].(string)
					assert.GreaterOrEqual(t, startDate, "2024-10-20")
					assert.LessOrEqual(t, startDate, "2024-10-25")
				}
			},
		},
		{
			name: "search with no results",
			queryParams: map[string]string{
				"search": "NonexistentEvent",
			},
			wantCode:       http.StatusOK,
			wantMinResults: 0,
			wantMaxResults: 0,
			validateFunc: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				assert.Equal(t, 0, len(data))

				pagination := resp["pagination"].(map[string]interface{})
				assert.Equal(t, float64(0), pagination["total_items"])
			},
		},
	}

	// Helper function to send GET request with query params
	listOpportunities := func(queryParams map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/opportunities", nil)

		// Add query parameters
		q := req.URL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := listOpportunities(tt.queryParams)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// Validate data array length
				data := resp["data"].([]interface{})
				assert.GreaterOrEqual(t, len(data), tt.wantMinResults)
				assert.LessOrEqual(t, len(data), tt.wantMaxResults)

				// Run custom validation if provided
				if tt.validateFunc != nil {
					tt.validateFunc(t, resp)
				}
			}
		})
	}
}
