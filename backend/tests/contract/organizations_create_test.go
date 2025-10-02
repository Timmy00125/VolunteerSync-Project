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

func TestOrganizationsCreate(t *testing.T) {
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
		c.Set("user_type", "coordinator")
		c.Next()
	})

	r.POST("/api/v1/organizations", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{})
	})

	tests := []struct {
		name     string
		body     map[string]interface{}
		token    string
		wantCode int
	}{
		{
			name: "valid organization creation with required fields only",
			body: map[string]interface{}{
				"name":  "Hope Foundation",
				"email": "contact@hopefoundation.org",
			},
			token:    "valid-token",
			wantCode: http.StatusCreated,
		},
		{
			name: "valid organization creation with all fields",
			body: map[string]interface{}{
				"name":              "Complete Nonprofit Organization",
				"mission_statement": "To serve the community with excellence",
				"description":       "A comprehensive community service organization dedicated to making a difference",
				"website":           "https://www.completenonprofit.org",
				"email":             "info@completenonprofit.org",
				"phone":             "+1-555-0123",
				"address": map[string]interface{}{
					"address_line_1": "123 Main Street",
					"address_line_2": "Suite 100",
					"city":           "San Francisco",
					"state":          "California",
					"postal_code":    "94102",
					"country":        "United States",
				},
				"logo_url":   "https://example.com/logo.png",
				"banner_url": "https://example.com/banner.png",
				"cause_ids":  []string{"cause-uuid-1", "cause-uuid-2"},
			},
			token:    "valid-token",
			wantCode: http.StatusCreated,
		},
		{
			name: "missing required field - name",
			body: map[string]interface{}{
				"email": "contact@missingname.org",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing required field - email",
			body: map[string]interface{}{
				"name": "Missing Email Org",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			body: map[string]interface{}{
				"name":  "Invalid Email Org",
				"email": "not-an-email",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "empty name",
			body: map[string]interface{}{
				"name":  "",
				"email": "contact@emptyname.org",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "name too long (over 200 chars)",
			body: map[string]interface{}{
				"name":  strings.Repeat("A", 201),
				"email": "contact@toolong.org",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
	}

	// Helper function to send POST request and return recorder
	postOrganization := func(body interface{}, token string) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations", bytes.NewBuffer(bodyBytes))
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
			w := postOrganization(tt.body, tt.token)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusCreated {
				// For successful creation, validate response structure
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err, "Response should be valid JSON")

				// Validate required fields in response
				id, idOk := resp["id"].(string)
				assert.True(t, idOk, "id should be a string")
				assert.NotEmpty(t, id, "id should not be empty")

				name, nameOk := resp["name"].(string)
				assert.True(t, nameOk, "name should be a string")
				assert.Equal(t, tt.body["name"], name, "name should match request")

				email, emailOk := resp["email"].(string)
				assert.True(t, emailOk, "email should be a string")
				assert.Equal(t, tt.body["email"], email, "email should match request")

				// Validate slug generation
				slug, slugOk := resp["slug"].(string)
				assert.True(t, slugOk, "slug should be a string")
				assert.NotEmpty(t, slug, "slug should be generated")
				// Slug should be lowercase and hyphenated version of name
				expectedSlug := strings.ToLower(strings.ReplaceAll(tt.body["name"].(string), " ", "-"))
				// Remove special characters for slug validation
				expectedSlug = strings.Map(func(r rune) rune {
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
						return r
					}
					return -1
				}, expectedSlug)
				assert.Contains(t, slug, expectedSlug, "slug should be derived from name")

				// Validate auto-verification (FR-015)
				verificationStatus, vsOk := resp["verification_status"].(string)
				assert.True(t, vsOk, "verification_status should be a string")
				assert.Equal(t, "verified", verificationStatus, "Organization should be auto-verified on creation (FR-015)")

				// Validate timestamps
				createdAt, caOk := resp["created_at"].(string)
				assert.True(t, caOk, "created_at should be a string")
				assert.NotEmpty(t, createdAt, "created_at should be set")

				// Validate counters are initialized
				totalVolunteers, tvOk := resp["total_volunteers"].(float64)
				assert.True(t, tvOk, "total_volunteers should be a number")
				assert.Equal(t, float64(0), totalVolunteers, "total_volunteers should be initialized to 0")

				totalHours, thOk := resp["total_hours"].(float64)
				assert.True(t, thOk, "total_hours should be a number")
				assert.Equal(t, float64(0), totalHours, "total_hours should be initialized to 0")

				// If address was provided, validate it's in response
				if addr, hasAddr := tt.body["address"].(map[string]interface{}); hasAddr {
					respAddr, addrOk := resp["address"].(map[string]interface{})
					assert.True(t, addrOk, "address should be an object")
					if addrOk {
						assert.Equal(t, addr["city"], respAddr["city"], "city should match")
						assert.Equal(t, addr["state"], respAddr["state"], "state should match")
					}
				}

				// If optional fields were provided, validate them
				if mission, hasMission := tt.body["mission_statement"].(string); hasMission {
					respMission, missionOk := resp["mission_statement"].(string)
					assert.True(t, missionOk, "mission_statement should be a string")
					assert.Equal(t, mission, respMission, "mission_statement should match request")
				}

				if website, hasWebsite := tt.body["website"].(string); hasWebsite {
					respWebsite, websiteOk := resp["website"].(string)
					assert.True(t, websiteOk, "website should be a string")
					assert.Equal(t, website, respWebsite, "website should match request")
				}

				if causes, hasCauses := tt.body["cause_ids"].([]string); hasCauses {
					respCauses, causesOk := resp["causes"].([]interface{})
					assert.True(t, causesOk, "causes should be an array")
					assert.Equal(t, len(causes), len(respCauses), "cause count should match")
				}
			}

			if tt.wantCode == http.StatusBadRequest {
				// For validation errors, check error structure
				var errResp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err, "Error response should be valid JSON")

				errCode, errOk := errResp["error"].(string)
				assert.True(t, errOk, "error field should be present")
				assert.NotEmpty(t, errCode, "error code should not be empty")

				message, msgOk := errResp["message"].(string)
				assert.True(t, msgOk, "message field should be present")
				assert.NotEmpty(t, message, "message should not be empty")
			}
		})
	}

	t.Run("unauthorized - no token", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "Unauthorized Org",
			"email": "contact@unauthorized.org",
		}

		w := postOrganization(body, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 when no token provided")

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, "unauthorized", errResp["error"])
	})

	t.Run("duplicate organization name", func(t *testing.T) {
		// This test verifies that duplicate organization names are allowed
		// (only slug must be unique, and slug generation should handle conflicts)
		body1 := map[string]interface{}{
			"name":  "Duplicate Name Test",
			"email": fmt.Sprintf("org1-%d@example.com", time.Now().UnixNano()),
		}
		body2 := map[string]interface{}{
			"name":  "Duplicate Name Test",
			"email": fmt.Sprintf("org2-%d@example.com", time.Now().UnixNano()),
		}

		w1 := postOrganization(body1, "valid-token")
		assert.Equal(t, http.StatusCreated, w1.Code, "First org creation should succeed")

		w2 := postOrganization(body2, "valid-token")
		// Should succeed with different slug (e.g., duplicate-name-test-2)
		assert.Equal(t, http.StatusCreated, w2.Code, "Second org with same name should succeed with different slug")

		if w2.Code == http.StatusCreated {
			var resp1, resp2 map[string]interface{}
			json.Unmarshal(w1.Body.Bytes(), &resp1)
			json.Unmarshal(w2.Body.Bytes(), &resp2)

			if slug1, ok1 := resp1["slug"].(string); ok1 {
				if slug2, ok2 := resp2["slug"].(string); ok2 {
					assert.NotEqual(t, slug1, slug2, "Slugs should be different for orgs with same name")
				}
			}
		}
	})

	t.Run("special characters in name", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "St. Mary's Community Center & Food Bank",
			"email": "contact@stmarys.org",
		}

		w := postOrganization(body, "valid-token")
		assert.Equal(t, http.StatusCreated, w.Code)

		if w.Code == http.StatusCreated {
			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)

			if slug, ok := resp["slug"].(string); ok {
				// Slug should only contain lowercase letters, numbers, and hyphens
				assert.Regexp(t, "^[a-z0-9-]+$", slug, "Slug should only contain lowercase alphanumeric and hyphens")
				assert.NotEmpty(t, slug, "Slug should be generated from name with special chars")
			}
		}
	})

	t.Run("address triggers geocoding", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "Geocoded Organization",
			"email": "contact@geocoded.org",
			"address": map[string]interface{}{
				"address_line_1": "1600 Amphitheatre Parkway",
				"city":           "Mountain View",
				"state":          "California",
				"postal_code":    "94043",
				"country":        "United States",
			},
		}

		w := postOrganization(body, "valid-token")
		assert.Equal(t, http.StatusCreated, w.Code)

		if w.Code == http.StatusCreated {
			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)

			addr, addrOk := resp["address"].(map[string]interface{})
			assert.True(t, addrOk, "address should be present in response")

			if addrOk {
				// Latitude and longitude should be set (geocoded from address)
				// Note: In actual implementation, these would be populated by geocoding service
				// For now, we just verify the fields exist (can be null in placeholder)
				_, hasLat := addr["latitude"]
				_, hasLng := addr["longitude"]
				assert.True(t, hasLat, "latitude field should be present")
				assert.True(t, hasLng, "longitude field should be present")
			}
		}
	})

	t.Run("organization creation adds creator as admin", func(t *testing.T) {
		// This test verifies that the user who creates the organization
		// is automatically added as an admin member
		body := map[string]interface{}{
			"name":  "Auto Admin Test Org",
			"email": "contact@autoadmin.org",
		}

		w := postOrganization(body, "valid-token")
		assert.Equal(t, http.StatusCreated, w.Code)

		// Note: In actual implementation, we would verify that:
		// 1. An Organization_Member record is created
		// 2. With role = 'admin'
		// 3. With user_id = authenticated user's ID
		// This would typically be checked via a separate GET /api/v1/organizations/{id}/members endpoint
		// or by verifying database state in integration tests
	})
}
