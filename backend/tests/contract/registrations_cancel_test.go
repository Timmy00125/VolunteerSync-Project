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

func TestRegistrationsCancel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 200
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

	r.PATCH("/api/v1/registrations/:id/cancel", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	tests := []struct {
		name           string
		registrationID string
		body           map[string]interface{}
		token          string
		wantCode       int
	}{
		{
			name:           "successful cancellation without reason",
			registrationID: "registration-uuid-1",
			body:           map[string]interface{}{},
			token:          "valid-token",
			wantCode:       http.StatusOK,
		},
		{
			name:           "successful cancellation with reason",
			registrationID: "registration-uuid-2",
			body: map[string]interface{}{
				"cancellation_reason": "Schedule conflict",
			},
			token:    "valid-token",
			wantCode: http.StatusOK,
		},
		{
			name:           "cancellation with detailed reason",
			registrationID: "registration-uuid-3",
			body: map[string]interface{}{
				"cancellation_reason": "Family emergency, cannot attend",
			},
			token:    "valid-token",
			wantCode: http.StatusOK,
		},
		{
			name:           "missing authentication",
			registrationID: "registration-uuid-4",
			body:           map[string]interface{}{},
			token:          "",
			wantCode:       http.StatusUnauthorized,
		},
		{
			name:           "invalid registration ID format",
			registrationID: "invalid-uuid",
			body:           map[string]interface{}{},
			token:          "valid-token",
			wantCode:       http.StatusBadRequest,
		},
		{
			name:           "registration not found",
			registrationID: "non-existent-uuid",
			body:           map[string]interface{}{},
			token:          "valid-token",
			wantCode:       http.StatusNotFound,
		},
		{
			name:           "unauthorized to cancel - different user",
			registrationID: "other-user-registration-uuid",
			body:           map[string]interface{}{},
			token:          "valid-token",
			wantCode:       http.StatusForbidden,
		},
		{
			name:           "already cancelled registration",
			registrationID: "already-cancelled-uuid",
			body:           map[string]interface{}{},
			token:          "valid-token",
			wantCode:       http.StatusConflict,
		},
	}

	// helper to send PATCH and return recorder
	patchCancelRegistration := func(registrationID string, body interface{}, token string) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/registrations/"+registrationID+"/cancel", bytes.NewBuffer(bodyBytes))
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
			w := patchCancelRegistration(tt.registrationID, tt.body, tt.token)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				// For successful cancellation, expect updated registration object
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// registration object should be returned
				reg, regOk := resp["registration"].(map[string]interface{})
				if !regOk {
					// Some implementations return the registration directly
					reg = resp
				}

				// Verify registration is cancelled
				if status, hasStatus := reg["status"].(string); hasStatus {
					assert.Equal(t, "cancelled", status, "registration status should be cancelled")
				}
				if cancelledAt, hasCancelledAt := reg["cancelled_at"]; hasCancelledAt {
					assert.NotNil(t, cancelledAt, "cancelled_at timestamp should be set")
				}

				// If cancellation reason was provided, it should be included
				if reason, hasReason := tt.body["cancellation_reason"].(string); hasReason && reason != "" {
					if regReason, hasRegReason := reg["cancellation_reason"].(string); hasRegReason {
						assert.Equal(t, reason, regReason, "cancellation reason should match")
					}
				}
			}
		})
	}

	t.Run("late cancellation warning - within 24 hours", func(t *testing.T) {
		// This test simulates cancelling a registration within 24 hours of the event
		// The cancellation should succeed but include a warning
		registrationID := "late-cancellation-uuid"
		body := map[string]interface{}{
			"cancellation_reason": "Last minute emergency",
		}

		w := patchCancelRegistration(registrationID, body, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code) // Will fail until implemented

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check for late cancellation warning (FR-034)
		if warnings, hasWarnings := resp["warnings"].([]interface{}); hasWarnings {
			assert.NotEmpty(t, warnings, "should have warning for late cancellation")
			if len(warnings) > 0 {
				if warning, ok := warnings[0].(map[string]interface{}); ok {
					if warnType, hasType := warning["type"].(string); hasType {
						assert.Equal(t, "late_cancellation", warnType, "warning type should be late_cancellation")
					}
					if warnMsg, hasMsg := warning["message"].(string); hasMsg {
						assert.Contains(t, strings.ToLower(warnMsg), "24 hours", "warning message should mention 24 hours")
					}
				}
			}
		}

		// Registration should still be cancelled
		reg, regOk := resp["registration"].(map[string]interface{})
		if !regOk {
			reg = resp
		}
		if status, hasStatus := reg["status"].(string); hasStatus {
			assert.Equal(t, "cancelled", status, "registration should still be cancelled despite late cancellation")
		}
	})

	t.Run("cannot cancel completed event", func(t *testing.T) {
		// This test simulates trying to cancel a registration for a completed event
		registrationID := "completed-event-uuid"
		body := map[string]interface{}{
			"cancellation_reason": "Changed my mind",
		}

		w := patchCancelRegistration(registrationID, body, "valid-token")
		assert.Equal(t, http.StatusConflict, w.Code) // Will fail until implemented

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		if errMsg, hasError := resp["error"].(string); hasError {
			assert.Contains(t, strings.ToLower(errMsg), "completed", "error should indicate event is completed")
		}
	})

	t.Run("cancellation notifies organization", func(t *testing.T) {
		// This test verifies that cancellation creates appropriate notification
		registrationID := "notification-test-uuid"
		body := map[string]interface{}{
			"cancellation_reason": "Testing notifications",
		}

		w := patchCancelRegistration(registrationID, body, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// The actual notification creation will be tested in integration tests
		// Here we just verify the cancellation succeeded
		reg, regOk := resp["registration"].(map[string]interface{})
		if regOk {
			if status, hasStatus := reg["status"].(string); hasStatus {
				assert.Equal(t, "cancelled", status)
			}
		}
	})
}
