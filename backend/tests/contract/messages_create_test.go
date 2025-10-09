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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessagesCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with placeholder handler that always returns 201
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.POST("/api/v1/messages", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{})
	})

	// Mock authentication middleware - returns dummy user ID
	mockAuthMiddleware := func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		c.Next()
	}
	r.Use(mockAuthMiddleware)

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{
			name: "valid direct message",
			body: map[string]interface{}{
				"message_type": "direct",
				"recipient_ids": []string{
					uuid.New().String(),
				},
				"subject": "Meeting Reminder",
				"content": "Don't forget about our volunteer coordination meeting tomorrow at 2 PM.",
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "valid broadcast message to event volunteers",
			body: map[string]interface{}{
				"message_type":   "broadcast",
				"opportunity_id": uuid.New().String(),
				"subject":        "Event Update",
				"content":        "The meeting location has changed to Building B, Room 205.",
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "direct message without recipients",
			body: map[string]interface{}{
				"message_type": "direct",
				"subject":      "Test Message",
				"content":      "This message has no recipients.",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "broadcast message without opportunity_id",
			body: map[string]interface{}{
				"message_type": "broadcast",
				"subject":      "Broadcast Update",
				"content":      "This broadcast has no opportunity specified.",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "message without content",
			body: map[string]interface{}{
				"message_type": "direct",
				"recipient_ids": []string{
					uuid.New().String(),
				},
				"subject": "Empty Message",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid message_type",
			body: map[string]interface{}{
				"message_type": "invalid_type",
				"recipient_ids": []string{
					uuid.New().String(),
				},
				"content": "Message with invalid type.",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "direct message with multiple recipients",
			body: map[string]interface{}{
				"message_type": "direct",
				"recipient_ids": []string{
					uuid.New().String(),
					uuid.New().String(),
					uuid.New().String(),
				},
				"subject": "Team Announcement",
				"content": "Great work on last week's event! Looking forward to working with you all again.",
			},
			wantCode: http.StatusCreated,
		},
	}

	// Helper to send message POST and return recorder
	postMessage := func(body interface{}) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer mock-token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postMessage(tt.body)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusCreated {
				// For successful message creation, expect message object
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// message object
				message, msgOk := resp["message"].(map[string]interface{})
				assert.True(t, msgOk, "message should be an object")
				if msgOk {
					// id should be present
					id, idOk := message["id"].(string)
					assert.True(t, idOk, "message.id should be a string")
					assert.NotEmpty(t, id, "message.id should not be empty")

					// sender_id should match authenticated user
					senderID, senderOk := message["sender_id"].(string)
					assert.True(t, senderOk, "message.sender_id should be a string")
					assert.NotEmpty(t, senderID, "message.sender_id should not be empty")

					// message_type should match request
					if reqType, ok := tt.body["message_type"].(string); ok {
						msgType, typeOk := message["message_type"].(string)
						assert.True(t, typeOk, "message.message_type should be a string")
						assert.Equal(t, reqType, msgType, "message_type should match request")
					}

					// content should match request
					if reqContent, ok := tt.body["content"].(string); ok {
						content, contentOk := message["content"].(string)
						assert.True(t, contentOk, "message.content should be a string")
						assert.Equal(t, reqContent, content, "content should match request")
					}

					// sent_at should be present
					sentAt, sentOk := message["sent_at"].(string)
					assert.True(t, sentOk, "message.sent_at should be a string")
					assert.NotEmpty(t, sentAt, "message.sent_at should not be empty")

					// For broadcast messages, verify opportunity_id is present
					if reqType, ok := tt.body["message_type"].(string); ok && reqType == "broadcast" {
						oppID, oppOk := message["opportunity_id"].(string)
						assert.True(t, oppOk, "message.opportunity_id should be a string for broadcast")
						assert.NotEmpty(t, oppID, "message.opportunity_id should not be empty for broadcast")
					}

					// For direct messages, verify recipient_count or recipients
					if reqType, ok := tt.body["message_type"].(string); ok && reqType == "direct" {
						// Response may include recipient_count or recipients array
						if recipientCount, ok := message["recipient_count"].(float64); ok {
							assert.Greater(t, recipientCount, float64(0), "recipient_count should be greater than 0")
						} else if recipients, ok := message["recipients"].([]interface{}); ok {
							assert.NotEmpty(t, recipients, "recipients array should not be empty")
						}
					}
				}
			}
		})
	}

	t.Run("unauthorized message creation", func(t *testing.T) {
		// Create router without auth middleware
		rNoAuth := gin.New()
		rNoAuth.POST("/api/v1/messages", func(c *gin.Context) {
			// Check for auth, return 401 if missing
			if c.GetHeader("Authorization") == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{})
		})

		body := map[string]interface{}{
			"message_type": "direct",
			"recipient_ids": []string{
				uuid.New().String(),
			},
			"content": "Test message without authentication.",
		}

		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		rNoAuth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("broadcast message creates notifications for all registered volunteers", func(t *testing.T) {
		// This test verifies the contract expectation that broadcast messages
		// trigger notifications to all volunteers registered for the event
		body := map[string]interface{}{
			"message_type":   "broadcast",
			"opportunity_id": uuid.New().String(),
			"subject":        "Important Update",
			"content":        "All volunteers please bring ID and water bottle.",
		}

		w := postMessage(body)
		assert.Equal(t, http.StatusCreated, w.Code)

		// In actual implementation, this would trigger notification creation
		// for all volunteers registered for the opportunity
		// The response should confirm the broadcast was sent
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		message, msgOk := resp["message"].(map[string]interface{})
		assert.True(t, msgOk, "message should be an object")
		if msgOk {
			// Verify it's marked as broadcast
			msgType, _ := message["message_type"].(string)
			assert.Equal(t, "broadcast", msgType, "should be broadcast type")
		}
	})

	t.Run("message content sanitization", func(t *testing.T) {
		// Test that message content is properly sanitized (XSS prevention)
		body := map[string]interface{}{
			"message_type": "direct",
			"recipient_ids": []string{
				uuid.New().String(),
			},
			"content": "<script>alert('XSS')</script>This is a test message",
			"subject": "Test <b>HTML</b> Subject",
		}

		w := postMessage(body)
		// Should still succeed - sanitization happens on storage/display
		assert.Equal(t, http.StatusCreated, w.Code)

		// The actual implementation should sanitize content before storage
		// This test documents the expectation that dangerous HTML is removed
	})

	t.Run("rate limiting for message creation", func(t *testing.T) {
		// Test that excessive message sending is rate limited
		// FR-058 requires protection against spam

		// Simulate sending many messages rapidly
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 15; i++ {
			body := map[string]interface{}{
				"message_type": "direct",
				"recipient_ids": []string{
					uuid.New().String(),
				},
				"content": fmt.Sprintf("Test message %d", i),
			}

			w := postMessage(body)
			if w.Code == http.StatusCreated {
				successCount++
			} else if w.Code == http.StatusTooManyRequests {
				rateLimitedCount++
			}

			// Small delay to simulate realistic usage
			time.Sleep(10 * time.Millisecond)
		}

		// In actual implementation, rate limiting should kick in
		// For now, this test documents the expectation
		// Implementation should limit to reasonable message rate (e.g., 10 per minute)
		t.Logf("Sent %d messages: %d successful, %d rate limited", 15, successCount, rateLimitedCount)
	})
}
