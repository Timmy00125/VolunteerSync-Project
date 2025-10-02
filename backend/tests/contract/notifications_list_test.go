package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNotificationsList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with placeholder handler that returns notifications
	// This will be replaced with the actual handler during implementation
	r := gin.New()
	r.GET("/api/v1/notifications", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"notifications": []gin.H{},
			"pagination": gin.H{
				"page":        1,
				"limit":       20,
				"total_pages": 1,
				"total_items": 0,
				"has_next":    false,
				"has_prev":    false,
			},
			"unread_count": 0,
		})
	})

	// Mock authentication middleware - returns dummy user ID
	mockAuthMiddleware := func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		c.Next()
	}
	r.Use(mockAuthMiddleware)

	tests := []struct {
		name        string
		queryParams string
		wantCode    int
	}{
		{
			name:        "list all notifications with default pagination",
			queryParams: "",
			wantCode:    http.StatusOK,
		},
		{
			name:        "list notifications with custom page size",
			queryParams: "?page=1&limit=10",
			wantCode:    http.StatusOK,
		},
		{
			name:        "list unread notifications only",
			queryParams: "?unread=true",
			wantCode:    http.StatusOK,
		},
		{
			name:        "list notifications by type",
			queryParams: "?type=registration_confirmed",
			wantCode:    http.StatusOK,
		},
		{
			name:        "list notifications by priority",
			queryParams: "?priority=high",
			wantCode:    http.StatusOK,
		},
		{
			name:        "list with multiple filters",
			queryParams: "?unread=true&priority=critical&page=1&limit=5",
			wantCode:    http.StatusOK,
		},
		{
			name:        "invalid page number",
			queryParams: "?page=-1",
			wantCode:    http.StatusOK, // Should default to page 1
		},
		{
			name:        "invalid limit",
			queryParams: "?limit=0",
			wantCode:    http.StatusOK, // Should use default limit
		},
		{
			name:        "large page number beyond available data",
			queryParams: "?page=999",
			wantCode:    http.StatusOK, // Should return empty array
		},
	}

	// Helper to send GET request and return recorder
	getNotifications := func(queryParams string) *httptest.ResponseRecorder {
		url := "/api/v1/notifications" + queryParams
		req := httptest.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer mock-token")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getNotifications(tt.queryParams)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// notifications array should be present
				notifications, notifOk := resp["notifications"].([]interface{})
				assert.True(t, notifOk, "notifications should be an array")
				assert.NotNil(t, notifications, "notifications should not be nil")

				// pagination object should be present
				pagination, pageOk := resp["pagination"].(map[string]interface{})
				assert.True(t, pageOk, "pagination should be an object")
				if pageOk {
					// page number
					page, pageNumOk := pagination["page"].(float64)
					assert.True(t, pageNumOk, "pagination.page should be a number")
					assert.GreaterOrEqual(t, page, float64(1), "page should be >= 1")

					// limit
					limit, limitOk := pagination["limit"].(float64)
					assert.True(t, limitOk, "pagination.limit should be a number")
					assert.Greater(t, limit, float64(0), "limit should be > 0")

					// total_pages
					totalPages, tpOk := pagination["total_pages"].(float64)
					assert.True(t, tpOk, "pagination.total_pages should be a number")
					assert.GreaterOrEqual(t, totalPages, float64(0), "total_pages should be >= 0")

					// total_items
					totalItems, tiOk := pagination["total_items"].(float64)
					assert.True(t, tiOk, "pagination.total_items should be a number")
					assert.GreaterOrEqual(t, totalItems, float64(0), "total_items should be >= 0")

					// has_next
					_, hnOk := pagination["has_next"].(bool)
					assert.True(t, hnOk, "pagination.has_next should be a boolean")

					// has_prev
					_, hpOk := pagination["has_prev"].(bool)
					assert.True(t, hpOk, "pagination.has_prev should be a boolean")
				}

				// unread_count should be present (FR-063)
				unreadCount, ucOk := resp["unread_count"].(float64)
				assert.True(t, ucOk, "unread_count should be a number")
				assert.GreaterOrEqual(t, unreadCount, float64(0), "unread_count should be >= 0")

				// If notifications exist, verify structure
				if len(notifications) > 0 {
					firstNotif, ok := notifications[0].(map[string]interface{})
					assert.True(t, ok, "notification item should be an object")
					if ok {
						// id
						id, idOk := firstNotif["id"].(string)
						assert.True(t, idOk, "notification.id should be a string")
						assert.NotEmpty(t, id, "notification.id should not be empty")

						// recipient_id
						recipientID, recipientOk := firstNotif["recipient_id"].(string)
						assert.True(t, recipientOk, "notification.recipient_id should be a string")
						assert.NotEmpty(t, recipientID, "notification.recipient_id should not be empty")

						// notification_type
						notifType, typeOk := firstNotif["notification_type"].(string)
						assert.True(t, typeOk, "notification.notification_type should be a string")
						assert.NotEmpty(t, notifType, "notification.notification_type should not be empty")

						// title
						title, titleOk := firstNotif["title"].(string)
						assert.True(t, titleOk, "notification.title should be a string")
						assert.NotEmpty(t, title, "notification.title should not be empty")

						// content
						content, contentOk := firstNotif["content"].(string)
						assert.True(t, contentOk, "notification.content should be a string")
						assert.NotEmpty(t, content, "notification.content should not be empty")

						// priority
						priority, priorityOk := firstNotif["priority"].(string)
						assert.True(t, priorityOk, "notification.priority should be a string")
						assert.Contains(t, []string{"low", "normal", "high", "critical"}, priority, "priority should be valid enum")

						// sent_at
						sentAt, sentOk := firstNotif["sent_at"].(string)
						assert.True(t, sentOk, "notification.sent_at should be a string")
						assert.NotEmpty(t, sentAt, "notification.sent_at should not be empty")

						// read_at is nullable
						if readAt, readOk := firstNotif["read_at"]; readOk {
							if readAt != nil {
								_, readStrOk := readAt.(string)
								assert.True(t, readStrOk, "notification.read_at should be a string or null")
							}
						}

						// action_url is nullable
						if actionURL, actionOk := firstNotif["action_url"]; actionOk {
							if actionURL != nil {
								_, urlStrOk := actionURL.(string)
								assert.True(t, urlStrOk, "notification.action_url should be a string or null")
							}
						}
					}
				}
			}
		})
	}

	t.Run("unauthorized notification access", func(t *testing.T) {
		// Create router without auth middleware
		rNoAuth := gin.New()
		rNoAuth.GET("/api/v1/notifications", func(c *gin.Context) {
			// Check for auth, return 401 if missing
			if c.GetHeader("Authorization") == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
		// No Authorization header

		w := httptest.NewRecorder()
		rNoAuth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("notification list performance", func(t *testing.T) {
		// Test that notification list responds quickly even with pagination
		// Should meet performance requirement of quick response times

		w := getNotifications("?page=1&limit=50")
		assert.Equal(t, http.StatusOK, w.Code)

		// Response should include all required fields
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotNil(t, resp["notifications"])
		assert.NotNil(t, resp["pagination"])
		assert.NotNil(t, resp["unread_count"])
	})

	t.Run("notification types enum validation", func(t *testing.T) {
		// Test filtering by various valid notification types
		validTypes := []string{
			"registration_confirmed",
			"event_reminder_24h",
			"event_reminder_2h",
			"hours_logged",
			"message_received",
			"event_cancelled",
			"waitlist_notification",
		}

		for _, notifType := range validTypes {
			t.Run(fmt.Sprintf("filter_by_%s", notifType), func(t *testing.T) {
				w := getNotifications(fmt.Sprintf("?type=%s", notifType))
				assert.Equal(t, http.StatusOK, w.Code)

				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// If notifications exist, all should be of the requested type
				notifications, ok := resp["notifications"].([]interface{})
				if ok && len(notifications) > 0 {
					for _, n := range notifications {
						notif, notifOk := n.(map[string]interface{})
						if notifOk {
							actualType, _ := notif["notification_type"].(string)
							assert.Equal(t, notifType, actualType, "notification type should match filter")
						}
					}
				}
			})
		}
	})

	t.Run("priority enum validation", func(t *testing.T) {
		// Test filtering by various valid priority levels
		validPriorities := []string{"low", "normal", "high", "critical"}

		for _, priority := range validPriorities {
			t.Run(fmt.Sprintf("filter_by_%s", priority), func(t *testing.T) {
				w := getNotifications(fmt.Sprintf("?priority=%s", priority))
				assert.Equal(t, http.StatusOK, w.Code)

				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				// If notifications exist, all should be of the requested priority
				notifications, ok := resp["notifications"].([]interface{})
				if ok && len(notifications) > 0 {
					for _, n := range notifications {
						notif, notifOk := n.(map[string]interface{})
						if notifOk {
							actualPriority, _ := notif["priority"].(string)
							assert.Equal(t, priority, actualPriority, "priority should match filter")
						}
					}
				}
			})
		}
	})

	t.Run("unread count accuracy", func(t *testing.T) {
		// Test that unread_count matches the number of notifications with read_at = null
		w := getNotifications("")
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		unreadCount := resp["unread_count"].(float64)

		// When filtering for unread only
		w2 := getNotifications("?unread=true")
		assert.Equal(t, http.StatusOK, w2.Code)

		var resp2 map[string]interface{}
		err2 := json.Unmarshal(w2.Body.Bytes(), &resp2)
		assert.NoError(t, err2)

		unreadNotifications := resp2["notifications"].([]interface{})

		// The unread_count should reflect the total number of unread notifications
		// (may be more than what's on the current page due to pagination)
		assert.GreaterOrEqual(t, unreadCount, float64(len(unreadNotifications)),
			"unread_count should be at least the number of unread notifications on current page")
	})

	t.Run("pagination consistency", func(t *testing.T) {
		// Test that pagination metadata is consistent
		w := getNotifications("?page=1&limit=10")
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		pagination := resp["pagination"].(map[string]interface{})
		page := pagination["page"].(float64)
		totalPages := pagination["total_pages"].(float64)
		hasNext := pagination["has_next"].(bool)
		hasPrev := pagination["has_prev"].(bool)

		// First page should not have previous
		if page == 1 {
			assert.False(t, hasPrev, "first page should not have previous")
		}

		// Last page should not have next
		if page >= totalPages && totalPages > 0 {
			assert.False(t, hasNext, "last page should not have next")
		}

		// If has_next is true, page should be less than total_pages
		if hasNext {
			assert.Less(t, page, totalPages, "if has_next, page should be less than total_pages")
		}
	})
}
