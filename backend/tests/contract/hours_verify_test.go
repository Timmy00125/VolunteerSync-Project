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

// TestHoursVerify tests the POST /api/v1/hours/{id}/verify endpoint
// Tests FR-048 (volunteer verification), FR-051 (verified status), total hours increment
func TestHoursVerify(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with placeholder handler
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.POST("/api/v1/hours/:id/verify", func(c *gin.Context) {
		hoursLogID := c.Param("id")

		c.JSON(http.StatusOK, gin.H{
			"id":                hoursLogID,
			"registration_id":   "reg-123",
			"hours":             3.0,
			"logged_by_user_id": "coordinator-123",
			"status":            "verified",
			"coordinator_notes": "Great work, very enthusiastic",
			"verified_at":       time.Now().Format(time.RFC3339),
			"logged_at":         time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"created_at":        time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"updated_at":        time.Now().Format(time.RFC3339),
		})
	})

	t.Run("Valid volunteer verification", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// FR-048: Hours should be verified
		assert.Equal(t, "verified", resp["status"])

		// Validate verified_at timestamp is set
		assert.NotEmpty(t, resp["verified_at"])
		verifiedAt, ok := resp["verified_at"].(string)
		assert.True(t, ok)
		_, err = time.Parse(time.RFC3339, verifiedAt)
		assert.NoError(t, err)

		// Validate other fields
		assert.Equal(t, hoursLogID, resp["id"])
		assert.Equal(t, 3.0, resp["hours"])
		assert.NotEmpty(t, resp["logged_at"])
		assert.NotEmpty(t, resp["updated_at"])
	})

	t.Run("Verified status after confirmation", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// FR-051: Registration should show "verified" status
		assert.Equal(t, "verified", resp["status"])

		// Ensure verified_at is set but auto_verified_at is not
		assert.NotNil(t, resp["verified_at"])
		assert.Nil(t, resp["auto_verified_at"])
	})

	t.Run("Total hours increment after verification", func(t *testing.T) {
		// This test validates that after verification, the volunteer's total hours increase
		// In actual implementation, this would:
		// 1. Update Hours_Log status to "verified"
		// 2. Update Registration status to "completed" and hours_status to "verified"
		// 3. Increment Volunteer_Profile total_hours by the hours amount
		// 4. Increment Volunteer_Profile total_events

		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Verify the hours log is verified
		assert.Equal(t, "verified", resp["status"])

		// In actual implementation, we would also query:
		// - Registration.hours_status == "verified"
		// - Registration.status == "completed"
		// - Volunteer_Profile.total_hours increased by resp["hours"]
		// - Volunteer_Profile.total_events incremented by 1

		assert.Equal(t, 3.0, resp["hours"])
	})

	t.Run("Volunteer can add notes during verification", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{
			"volunteer_notes": "Thanks, this is accurate!",
		}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, "verified", resp["status"])
		// Note: The dummy handler doesn't return volunteer_notes, but actual implementation should
	})

	t.Run("Unauthorized - missing token", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden - different volunteer trying to verify", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer other.volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 403 Forbidden (only the volunteer who worked the hours can verify)
		assert.Equal(t, http.StatusForbidden, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResp["error"])
	})

	t.Run("Hours log not found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+nonExistentID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 404 Not Found
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Cannot verify already verified hours", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
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

	t.Run("Cannot verify disputed hours without resolution", func(t *testing.T) {
		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
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
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+invalidID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Verification completes registration status", func(t *testing.T) {
		// This test validates the full workflow:
		// 1. Hours are logged (pending status)
		// 2. Volunteer verifies hours
		// 3. Registration status changes to "completed"
		// 4. Hours_status changes to "verified"

		hoursLogID := uuid.New().String()
		verifyRequest := map[string]interface{}{}

		jsonData, _ := json.Marshal(verifyRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/"+hoursLogID+"/verify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Verify status is "verified"
		assert.Equal(t, "verified", resp["status"])

		// In actual implementation, we would also validate:
		// - Registration.status == "completed" (was "attended")
		// - Registration.hours_status == "verified" (was "pending")
		// - Registration.hours_verified_at timestamp is set
	})
}
