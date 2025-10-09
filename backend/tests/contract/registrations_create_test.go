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

func TestRegistrationsCreate(t *testing.T) {
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
		c.Set("user_type", "volunteer")
		c.Set("volunteer_profile_id", "mock-volunteer-profile-id")
		c.Next()
	})

	r.POST("/api/v1/registrations", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{})
	})

	tests := []struct {
		name     string
		body     map[string]interface{}
		token    string
		wantCode int
	}{
		{
			name: "immediate registration - successful",
			body: map[string]interface{}{
				"opportunity_id": "opportunity-uuid-1",
			},
			token:    "valid-token",
			wantCode: http.StatusCreated,
		},
		{
			name: "registration with optional note",
			body: map[string]interface{}{
				"opportunity_id": "opportunity-uuid-2",
				"note":           "Looking forward to helping out!",
			},
			token:    "valid-token",
			wantCode: http.StatusCreated,
		},
		{
			name: "missing required field - opportunity_id",
			body: map[string]interface{}{
				"note": "No opportunity specified",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing authentication",
			body: map[string]interface{}{
				"opportunity_id": "opportunity-uuid-1",
			},
			token:    "",
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "invalid opportunity_id format",
			body: map[string]interface{}{
				"opportunity_id": "invalid-uuid",
			},
			token:    "valid-token",
			wantCode: http.StatusBadRequest,
		},
	}

	// helper to send POST and return recorder
	postRegistration := func(body interface{}, token string) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/registrations", bytes.NewBuffer(bodyBytes))
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
			w := postRegistration(tt.body, tt.token)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusCreated {
				// For successful registration, expect a registration object
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// registration object should be returned
				reg, regOk := resp["registration"].(map[string]interface{})
				if !regOk {
					// Some implementations return the registration directly
					reg = resp
				}

				// Basic validation of returned registration
				if id, hasID := reg["id"].(string); hasID {
					assert.NotEmpty(t, id, "registration id should not be empty")
				}
				if status, hasStatus := reg["status"].(string); hasStatus {
					assert.Contains(t, []string{"confirmed", "waitlisted"}, status, "registration status should be confirmed or waitlisted")
				}
				if registeredAt, hasRegisteredAt := reg["registered_at"]; hasRegisteredAt {
					assert.NotNil(t, registeredAt, "registered_at should be set")
				}
			}
		})
	}

	t.Run("waitlist when at capacity", func(t *testing.T) {
		// This test simulates an opportunity that is at full capacity
		// The registration should be placed on the waitlist
		body := map[string]interface{}{
			"opportunity_id": "full-capacity-opportunity-uuid",
		}

		w := postRegistration(body, "valid-token")
		// Should still succeed but with waitlist status
		assert.Equal(t, http.StatusCreated, w.Code) // Will fail until implemented

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		reg, regOk := resp["registration"].(map[string]interface{})
		if !regOk {
			reg = resp
		}

		if status, hasStatus := reg["status"].(string); hasStatus {
			assert.Equal(t, "waitlisted", status, "registration should be waitlisted when at capacity")
		}
	})

	t.Run("duplicate registration prevention", func(t *testing.T) {
		body := map[string]interface{}{
			"opportunity_id": "duplicate-test-opportunity-uuid",
		}

		// First registration - should succeed
		w1 := postRegistration(body, "valid-token")
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Second registration for same opportunity - should fail
		w2 := postRegistration(body, "valid-token")
		assert.Equal(t, http.StatusConflict, w2.Code) // Will fail until implemented

		var resp map[string]interface{}
		err := json.Unmarshal(w2.Body.Bytes(), &resp)
		assert.NoError(t, err)

		if errMsg, hasError := resp["error"].(string); hasError {
			assert.Contains(t, strings.ToLower(errMsg), "already registered", "error should indicate duplicate registration")
		}
	})

	t.Run("overlapping event warning", func(t *testing.T) {
		// This test simulates registering for an opportunity that overlaps with an existing registration
		// The registration should succeed but include a warning
		body := map[string]interface{}{
			"opportunity_id": "overlapping-opportunity-uuid",
		}

		w := postRegistration(body, "valid-token")
		assert.Equal(t, http.StatusCreated, w.Code) // Will fail until implemented

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check for warning about overlapping events
		if warnings, hasWarnings := resp["warnings"].([]interface{}); hasWarnings {
			assert.NotEmpty(t, warnings, "should have warnings for overlapping events")
			if len(warnings) > 0 {
				if warning, ok := warnings[0].(map[string]interface{}); ok {
					if warnType, hasType := warning["type"].(string); hasType {
						assert.Equal(t, "overlapping_event", warnType, "warning type should be overlapping_event")
					}
				}
			}
		}
	})

	t.Run("performance test - registration completes within 1 second", func(t *testing.T) {
		body := map[string]interface{}{
			"opportunity_id": fmt.Sprintf("perf-test-opportunity-%d", time.Now().UnixNano()),
		}

		start := time.Now()
		w := postRegistration(body, "valid-token")
		duration := time.Since(start)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Less(t, duration, time.Second, "registration should complete within 1 second (NFR-006)")
	})
}
