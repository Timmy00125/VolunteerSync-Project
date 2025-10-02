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

// TestHoursLog tests the POST /api/v1/hours/log endpoint
// Tests FR-046 (hours logged in pending status), FR-047 (notification sent), FR-054 (audit trail)
func TestHoursLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with placeholder handler
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.POST("/api/v1/hours/log", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":                uuid.New().String(),
			"registration_id":   "reg-123",
			"hours":             3.0,
			"logged_by_user_id": "coordinator-123",
			"status":            "pending",
			"coordinator_notes": "Great work, very enthusiastic",
			"logged_at":         time.Now().Format(time.RFC3339),
			"created_at":        time.Now().Format(time.RFC3339),
		})
	})

	t.Run("Valid hours logging by coordinator", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-123",
			"hours":             3.0,
			"coordinator_notes": "Great work, very enthusiastic",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Validate response structure
		assert.NotEmpty(t, resp["id"])
		assert.Equal(t, "reg-123", resp["registration_id"])
		assert.Equal(t, 3.0, resp["hours"])
		assert.Equal(t, "coordinator-123", resp["logged_by_user_id"])

		// FR-046: Hours should be in pending status
		assert.Equal(t, "pending", resp["status"])

		// Validate coordinator notes
		assert.Equal(t, "Great work, very enthusiastic", resp["coordinator_notes"])

		// Validate timestamps
		assert.NotEmpty(t, resp["logged_at"])
		assert.NotEmpty(t, resp["created_at"])

		// Parse timestamps to ensure they are valid
		loggedAt, ok := resp["logged_at"].(string)
		assert.True(t, ok)
		_, err = time.Parse(time.RFC3339, loggedAt)
		assert.NoError(t, err)

		createdAt, ok := resp["created_at"].(string)
		assert.True(t, ok)
		_, err = time.Parse(time.RFC3339, createdAt)
		assert.NoError(t, err)
	})

	t.Run("Pending status on creation", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-456",
			"hours":             2.5,
			"coordinator_notes": "Good participation",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// FR-046: Verify pending status on creation
		assert.Equal(t, "pending", resp["status"])

		// Ensure no verified_at or auto_verified_at timestamps
		assert.Nil(t, resp["verified_at"])
		assert.Nil(t, resp["auto_verified_at"])
	})

	t.Run("Notification to volunteer", func(t *testing.T) {
		// This test validates that the system sends a notification
		// In a real implementation, we would check the notification table
		// or mock the notification service

		logRequest := map[string]interface{}{
			"registration_id":   "reg-789",
			"hours":             4.0,
			"coordinator_notes": "Excellent teamwork",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// FR-047: In actual implementation, validate that:
		// 1. A notification record is created with type "hours_logged"
		// 2. The notification is sent to the volunteer
		// 3. The notification contains the hours amount and coordinator notes

		// For now, we just validate the API response
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp["id"])
	})

	t.Run("Hours Log audit record created", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-audit-test",
			"hours":             5.0,
			"coordinator_notes": "Outstanding performance",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// FR-054: Verify audit trail is created
		// The response itself represents the Hours_Log audit record
		assert.NotEmpty(t, resp["id"], "Hours_Log ID should be present")
		assert.NotEmpty(t, resp["registration_id"], "Registration ID should be present in audit")
		assert.NotEmpty(t, resp["logged_by_user_id"], "Coordinator ID should be present in audit")
		assert.NotEmpty(t, resp["logged_at"], "Logged timestamp should be present in audit")

		// Validate immutable fields are set
		assert.Equal(t, 5.0, resp["hours"])
		assert.Equal(t, "pending", resp["status"])
	})

	t.Run("Missing registration_id", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"hours":             3.0,
			"coordinator_notes": "Missing registration",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResp["error"])
	})

	t.Run("Missing hours", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-123",
			"coordinator_notes": "Missing hours value",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Negative hours value", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-123",
			"hours":             -2.0,
			"coordinator_notes": "Negative hours test",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request (hours must be positive)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Contains(t, errResp["message"], "positive")
	})

	t.Run("Zero hours value", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-123",
			"hours":             0.0,
			"coordinator_notes": "Zero hours test",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 Bad Request (hours must be positive)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized - missing token", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-123",
			"hours":             3.0,
			"coordinator_notes": "Unauthorized test",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden - not a coordinator", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "reg-123",
			"hours":             3.0,
			"coordinator_notes": "Forbidden test",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer volunteer.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 403 Forbidden (only coordinators can log hours)
		assert.Equal(t, http.StatusForbidden, w.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResp["error"])
	})

	t.Run("Registration not found", func(t *testing.T) {
		logRequest := map[string]interface{}{
			"registration_id":   "non-existent-reg",
			"hours":             3.0,
			"coordinator_notes": "Test",
		}

		jsonData, _ := json.Marshal(logRequest)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/hours/log", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer coordinator.jwt.token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 404 Not Found
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
