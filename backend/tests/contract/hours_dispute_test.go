package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestHoursDispute tests the POST /api/v1/hours/{id}/dispute endpoint
// Tests FR-050 (hours dispute by volunteer, coordinator notification)
func TestHoursDispute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with placeholder handler
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.POST("/api/v1/hours/:id/dispute", func(c *gin.Context) {
		hoursLogID := c.Param("id")

		c.JSON(http.StatusOK, gin.H{
			"id":                hoursLogID,
			"registration_id":   "reg-123",
			"hours":             3.0,
			"logged_by_user_id": "coordinator-123",
			"status":            "disputed",
			"coordinator_notes": "Great work, very enthusiastic",
			"dispute_reason":    "I worked 4 hours, not 3",
			"disputed_at":       time.Now().Format(time.RFC3339),
			"logged_at":         time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"created_at":        time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"updated_at":        time.Now().Format(time.RFC3339),
		})
	})

	t.Run("Valid hours dispute by volunteer", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "I worked 4 hours, not 3",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// FR-050: Hours status should change to "disputed"
		assert.Equal(t, "disputed", resp["status"])

		// Validate dispute_reason is stored
		assert.Equal(t, "I worked 4 hours, not 3", resp["dispute_reason"])

		// Validate disputed_at timestamp is set
		assert.NotEmpty(t, resp["disputed_at"])
		disputedAt, ok := resp["disputed_at"].(string)
		assert.True(t, ok)
		_, err = time.Parse(time.RFC3339, disputedAt)
		assert.NoError(t, err)

		// Validate other fields
		assert.Equal(t, hoursLogID, resp["id"])
		assert.Equal(t, 3.0, resp["hours"])
		assert.NotEmpty(t, resp["logged_at"])
		assert.NotEmpty(t, resp["updated_at"])
	})

	t.Run("Dispute status change", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Hours don't match my records",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Verify status transitioned from "pending" to "disputed"
		assert.Equal(t, "disputed", resp["status"])

		// Ensure resolution fields are not yet set
		assert.Nil(t, resp["resolved_at"])
		assert.Nil(t, resp["resolution_notes"])
	})

	t.Run("Coordinator notification on dispute", func(t *testing.T) {
		// This test validates that the coordinator receives a notification
		// In actual implementation, we would check:
		// 1. A notification record is created with type "hours_disputed"
		// 2. The notification is sent to the coordinator who logged the hours
		// 3. The notification contains the dispute reason

		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Timeline doesn't match event duration",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// FR-050: Validate coordinator notification would be sent
		assert.Equal(t, "disputed", resp["status"])
		assert.NotEmpty(t, resp["logged_by_user_id"], "Coordinator ID should be present for notification")
		assert.NotEmpty(t, resp["dispute_reason"], "Dispute reason should be present in notification")
	})

	t.Run("Dispute resolution workflow initiation", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Event was longer than logged",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Validate dispute workflow is initiated
		assert.Equal(t, "disputed", resp["status"])
		assert.NotEmpty(t, resp["disputed_at"])

		// Audit trail remains immutable
		assert.Equal(t, 3.0, resp["hours"], "Original hours should not change")
		assert.NotEmpty(t, resp["coordinator_notes"], "Original coordinator notes preserved")
	})

	t.Run("Volunteer can provide detailed dispute reason", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		detailedReason := "I arrived at 9:00 AM and left at 1:30 PM, including a 30-minute break. This should be 4 hours, not 3 hours as logged."
		disputeRequest := map[string]interface{}{
			"dispute_reason": detailedReason,
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, "disputed", resp["status"])
		assert.Equal(t, detailedReason, resp["dispute_reason"])
	})

	t.Run("Missing dispute reason", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request (dispute reason is required)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResp["error"])
		assert.Contains(t, errResp["message"], "dispute_reason")
	})

	t.Run("Empty dispute reason", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized - missing token", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Test dispute",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden - different volunteer trying to dispute", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Test dispute",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer other.volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 403 Forbidden (only the volunteer who worked the hours can dispute)
		assert.Equal(t, http.StatusForbidden, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResp["error"])
	})

	t.Run("Hours log not found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Test dispute",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+nonExistentID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 404 Not Found
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Cannot dispute already verified hours", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Test dispute",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request or 409 Conflict
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusConflict}, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResp["error"])
	})

	t.Run("Cannot dispute already disputed hours", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Second dispute attempt",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request or 409 Conflict
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusConflict}, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResp["error"])
	})

	t.Run("Invalid hours log ID format", func(t *testing.T) {
		invalidID := "invalid-uuid"
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Test dispute",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+invalidID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Dispute preserves original hours value", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Hours should be 5, not 3",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Verify audit trail integrity - hours value doesn't change
		assert.Equal(t, 3.0, resp["hours"], "Original hours value must remain unchanged")
		assert.Equal(t, "disputed", resp["status"])

		// The dispute reason explains the discrepancy, but original data is preserved
		assert.NotEmpty(t, resp["dispute_reason"])
	})

	t.Run("Status transition validation - pending to disputed", func(t *testing.T) {
		// Valid status transitions:
		// pending → verified
		// pending → disputed
		// disputed → verified (after resolution)
		// Auto-verify: pending → verified (after 7 days)

		hoursLogID := uuid.New().String()
		disputeRequest := map[string]interface{}{
			"dispute_reason": "Validating status transition",
		}

		jsonData, _ := json.Marshal(disputeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/dispute", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Verify status transitioned to disputed
		assert.Equal(t, "disputed", resp["status"])

		// In actual implementation, validate Registration.hours_status also changed to "disputed"
	})
}
