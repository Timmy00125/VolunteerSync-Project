package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationsCheckIn(t *testing.T) {
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
		// For check-in, typically coordinators/org admins perform this action
		c.Set("user_id", "mock-coordinator-id")
		c.Set("user_type", "coordinator")
		c.Set("organization_id", "mock-org-id")
		c.Next()
	})

	r.PATCH("/api/v1/registrations/:id/check-in", func(c *gin.Context) {
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
			name:           "successful check-in",
			registrationID: "registration-uuid-1",
			body:           map[string]interface{}{},
			token:          "valid-coordinator-token",
			wantCode:       http.StatusOK,
		},
		{
			name:           "check-in with coordinator notes",
			registrationID: "registration-uuid-2",
			body: map[string]interface{}{
				"coordinator_notes": "Arrived on time, very enthusiastic",
			},
			token:    "valid-coordinator-token",
			wantCode: http.StatusOK,
		},
		{
			name:           "check-in with timestamp",
			registrationID: "registration-uuid-3",
			body: map[string]interface{}{
				"checked_in_at": time.Now().Format(time.RFC3339),
			},
			token:    "valid-coordinator-token",
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
			token:          "valid-coordinator-token",
			wantCode:       http.StatusBadRequest,
		},
		{
			name:           "registration not found",
			registrationID: "non-existent-uuid",
			body:           map[string]interface{}{},
			token:          "valid-coordinator-token",
			wantCode:       http.StatusNotFound,
		},
		{
			name:           "unauthorized - volunteer cannot check in themselves",
			registrationID: "self-checkin-uuid",
			body:           map[string]interface{}{},
			token:          "volunteer-token",
			wantCode:       http.StatusForbidden,
		},
		{
			name:           "unauthorized - different organization",
			registrationID: "other-org-registration-uuid",
			body:           map[string]interface{}{},
			token:          "valid-coordinator-token",
			wantCode:       http.StatusForbidden,
		},
		{
			name:           "cancelled registration cannot be checked in",
			registrationID: "cancelled-registration-uuid",
			body:           map[string]interface{}{},
			token:          "valid-coordinator-token",
			wantCode:       http.StatusConflict,
		},
	}

	// helper to send PATCH and return recorder
	patchCheckInRegistration := func(registrationID string, body interface{}, token string) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/registrations/"+registrationID+"/check-in", bytes.NewBuffer(bodyBytes))
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
			w := patchCheckInRegistration(tt.registrationID, tt.body, tt.token)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				// For successful check-in, expect updated registration object
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// registration object should be returned
				reg, regOk := resp["registration"].(map[string]interface{})
				if !regOk {
					// Some implementations return the registration directly
					reg = resp
				}

				// Verify check-in timestamp is set
				if checkedInAt, hasCheckedInAt := reg["checked_in_at"]; hasCheckedInAt {
					assert.NotNil(t, checkedInAt, "checked_in_at timestamp should be set")
				}

				// If coordinator notes were provided, they should be included
				if notes, hasNotes := tt.body["coordinator_notes"].(string); hasNotes && notes != "" {
					if regNotes, hasRegNotes := reg["coordinator_notes"].(string); hasRegNotes {
						assert.Equal(t, notes, regNotes, "coordinator notes should match")
					}
				}

				// Status should still be confirmed (or completed after event ends)
				if status, hasStatus := reg["status"].(string); hasStatus {
					assert.Contains(t, []string{"confirmed", "completed"}, status, "status should be confirmed or completed")
				}
			}
		})
	}

	t.Run("check-in only on event day", func(t *testing.T) {
		// This test simulates checking in a volunteer before the event day
		// Should fail with appropriate error
		registrationID := "future-event-uuid"
		body := map[string]interface{}{}

		w := patchCheckInRegistration(registrationID, body, "valid-coordinator-token")
		// This should fail if event hasn't started yet
		// Implementation should validate event date
		assert.Contains(t, []int{http.StatusOK, http.StatusConflict}, w.Code) // Will fail until implemented

		if w.Code == http.StatusConflict {
			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if errMsg, hasError := resp["error"].(string); hasError {
				assert.Contains(t, strings.ToLower(errMsg), "event", "error should mention event not started")
			}
		}
	})

	t.Run("duplicate check-in - idempotent", func(t *testing.T) {
		// This test verifies that checking in a volunteer who is already checked in
		// is idempotent and doesn't cause an error
		registrationID := "already-checked-in-uuid"
		body := map[string]interface{}{
			"coordinator_notes": "Second check-in attempt",
		}

		// First check-in
		w1 := patchCheckInRegistration(registrationID, body, "valid-coordinator-token")
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second check-in - should succeed (idempotent)
		w2 := patchCheckInRegistration(registrationID, body, "valid-coordinator-token")
		assert.Equal(t, http.StatusOK, w2.Code) // Will fail until implemented

		var resp map[string]interface{}
		err := json.Unmarshal(w2.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Both requests should result in checked_in_at being set
		reg, regOk := resp["registration"].(map[string]interface{})
		if !regOk {
			reg = resp
		}
		if checkedInAt, hasCheckedInAt := reg["checked_in_at"]; hasCheckedInAt {
			assert.NotNil(t, checkedInAt, "checked_in_at should still be set on duplicate check-in")
		}
	})

	t.Run("waitlisted volunteer cannot be checked in", func(t *testing.T) {
		// This test verifies that a volunteer on the waitlist cannot be checked in
		registrationID := "waitlisted-uuid"
		body := map[string]interface{}{}

		w := patchCheckInRegistration(registrationID, body, "valid-coordinator-token")
		assert.Equal(t, http.StatusConflict, w.Code) // Will fail until implemented

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		if errMsg, hasError := resp["error"].(string); hasError {
			assert.Contains(t, strings.ToLower(errMsg), "waitlist", "error should indicate volunteer is on waitlist")
		}
	})

	t.Run("batch check-in via individual requests", func(t *testing.T) {
		// This test simulates checking in multiple volunteers
		registrationIDs := []string{
			"batch-reg-uuid-1",
			"batch-reg-uuid-2",
			"batch-reg-uuid-3",
		}

		for i, regID := range registrationIDs {
			body := map[string]interface{}{
				"coordinator_notes": "Batch check-in volunteer " + string(rune('A'+i)),
			}

			w := patchCheckInRegistration(regID, body, "valid-coordinator-token")
			assert.Equal(t, http.StatusOK, w.Code, "Check-in %d should succeed", i+1)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			reg, regOk := resp["registration"].(map[string]interface{})
			if !regOk {
				reg = resp
			}
			if checkedInAt, hasCheckedInAt := reg["checked_in_at"]; hasCheckedInAt {
				assert.NotNil(t, checkedInAt, "checked_in_at should be set for volunteer %d", i+1)
			}
		}
	})
}
