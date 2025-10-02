package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that always returns 201
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{})
	})

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{
			name: "valid registration",
			body: map[string]interface{}{
				"email":      "test@example.com",
				"password":   "SecurePass123",
				"first_name": "John",
				"last_name":  "Doe",
				"user_type":  "volunteer",
				"security_questions": []map[string]string{
					{"question": "What is your mother's maiden name?", "answer": "Smith"},
					{"question": "What is the name of your first pet?", "answer": "Buddy"},
					{"question": "What city were you born in?", "answer": "New York"},
				},
				"phone": "+1234567890",
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "missing email",
			body: map[string]interface{}{
				"password":   "SecurePass123",
				"first_name": "John",
				"last_name":  "Doe",
				"user_type":  "volunteer",
				"phone":      "+1234567890",
				"security_questions": []map[string]string{
					{"question": "What is your mother's maiden name?", "answer": "Smith"},
					{"question": "What is the name of your first pet?", "answer": "Buddy"},
					{"question": "What city were you born in?", "answer": "New York"},
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "weak password",
			body: map[string]interface{}{
				"email":      "weak@example.com",
				"password":   "weak", // less than 8 chars
				"first_name": "John",
				"last_name":  "Doe",
				"user_type":  "volunteer",
				"phone":      "+1234567890",
				"security_questions": []map[string]string{
					{"question": "What is your mother's maiden name?", "answer": "Smith"},
					{"question": "What is the name of your first pet?", "answer": "Buddy"},
					{"question": "What city were you born in?", "answer": "New York"},
				},
			},
			wantCode: http.StatusBadRequest,
		},
	}

	// helper to send register POST and return recorder
	postRegister := func(body interface{}) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure unique emails for tests that include an email to avoid cross-test pollution
			if email, ok := tt.body["email"].(string); ok && email != "" {
				tt.body["email"] = fmt.Sprintf("%s+%d@example.com", email, time.Now().UnixNano())
			}

			w := postRegister(tt.body)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusCreated {
				// For successful registration, expect tokens and a user object. Assert types to avoid panics.
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// access_token
				at, atOk := resp["access_token"].(string)
				assert.True(t, atOk, "access_token should be a string")
				assert.NotEmpty(t, at, "access_token should not be empty")

				// refresh_token
				rt, rtOk := resp["refresh_token"].(string)
				assert.True(t, rtOk, "refresh_token should be a string")
				assert.NotEmpty(t, rt, "refresh_token should not be empty")

				// user object
				user, userOk := resp["user"].(map[string]interface{})
				assert.True(t, userOk, "user should be an object")
				if userOk {
					// email should match what we sent (if present)
					if sentEmail, ok := tt.body["email"].(string); ok {
						// Some implementations may lowercase or normalize email — check contains
						if ue, have := user["email"].(string); have {
							assert.Contains(t, ue, sentEmail[:len(sentEmail)-len(fmt.Sprintf("+%d@example.com", 0))], "user.email should contain base email")
						}
					}
				}
			}
		})
	}

	t.Run("duplicate email", func(t *testing.T) {
		body := map[string]interface{}{
			"email":      "duplicate@example.com",
			"password":   "SecurePass123",
			"first_name": "John",
			"last_name":  "Doe",
			"user_type":  "volunteer",
			"security_questions": []map[string]string{
				{"question": "What is your mother's maiden name?", "answer": "Smith"},
				{"question": "What is the name of your first pet?", "answer": "Buddy"},
				{"question": "What city were you born in?", "answer": "New York"},
			},
			"phone": "+1234567890",
		}

		// First registration
		bodyBytes, _ := json.Marshal(body)
		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// Second registration with same email
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusConflict, w2.Code) // Will fail until implemented
	})

	t.Run("rate limiting", func(t *testing.T) {
		body := map[string]interface{}{
			"email":      "rate@example.com",
			"password":   "SecurePass123",
			"first_name": "Rate",
			"last_name":  "Limit",
			"user_type":  "volunteer",
			"security_questions": []map[string]string{
				{"question": "What is your mother's maiden name?", "answer": "Smith"},
				{"question": "What is the name of your first pet?", "answer": "Buddy"},
				{"question": "What city were you born in?", "answer": "New York"},
			},
		}

		bodyBytes, _ := json.Marshal(body)

		// Send 5 requests - should succeed
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code, "Request %d should succeed", i+1)
		}

		// 6th request - should be rate limited
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code) // Will fail until implemented
	})
}
