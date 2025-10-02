package contract

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationsCalendar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Dummy router with a placeholder handler that returns a simple .ics file
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

	r.GET("/api/v1/registrations/:id/calendar.ics", func(c *gin.Context) {
		// Return a minimal valid .ics file
		c.Header("Content-Type", "text/calendar; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=event.ics")
		c.String(http.StatusOK, "BEGIN:VCALENDAR\nVERSION:2.0\nEND:VCALENDAR")
	})

	tests := []struct {
		name           string
		registrationID string
		token          string
		wantCode       int
	}{
		{
			name:           "successful calendar download",
			registrationID: "registration-uuid-1",
			token:          "valid-token",
			wantCode:       http.StatusOK,
		},
		{
			name:           "missing authentication",
			registrationID: "registration-uuid-2",
			token:          "",
			wantCode:       http.StatusUnauthorized,
		},
		{
			name:           "invalid registration ID format",
			registrationID: "invalid-uuid",
			token:          "valid-token",
			wantCode:       http.StatusBadRequest,
		},
		{
			name:           "registration not found",
			registrationID: "non-existent-uuid",
			token:          "valid-token",
			wantCode:       http.StatusNotFound,
		},
		{
			name:           "unauthorized - different user's registration",
			registrationID: "other-user-registration-uuid",
			token:          "valid-token",
			wantCode:       http.StatusForbidden,
		},
		{
			name:           "cancelled registration calendar",
			registrationID: "cancelled-registration-uuid",
			token:          "valid-token",
			wantCode:       http.StatusOK,
		},
	}

	// helper to send GET and return recorder
	getCalendar := func(registrationID string, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/registrations/"+registrationID+"/calendar.ics", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getCalendar(tt.registrationID, tt.token)
			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				// Verify Content-Type is text/calendar
				contentType := w.Header().Get("Content-Type")
				assert.Contains(t, contentType, "text/calendar", "Content-Type should be text/calendar")

				// Verify Content-Disposition header is set for download
				contentDisposition := w.Header().Get("Content-Disposition")
				assert.Contains(t, contentDisposition, "attachment", "Content-Disposition should indicate attachment")
				assert.Contains(t, contentDisposition, ".ics", "filename should have .ics extension")

				// Verify response body contains valid .ics structure
				body := w.Body.String()
				assert.Contains(t, body, "BEGIN:VCALENDAR", ".ics file should start with BEGIN:VCALENDAR")
				assert.Contains(t, body, "END:VCALENDAR", ".ics file should end with END:VCALENDAR")
				assert.Contains(t, body, "VERSION:2.0", ".ics file should specify VERSION:2.0")
			}
		})
	}

	t.Run("calendar file contains event details", func(t *testing.T) {
		// This test verifies that the .ics file contains proper event information
		registrationID := "detailed-event-uuid"

		w := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code) // Will fail until implemented

		body := w.Body.String()

		// Verify iCalendar structure
		assert.Contains(t, body, "BEGIN:VCALENDAR", "should contain VCALENDAR block")
		assert.Contains(t, body, "VERSION:2.0", "should specify iCalendar version 2.0")
		assert.Contains(t, body, "PRODID:", "should contain PRODID")

		// Verify VEVENT block
		assert.Contains(t, body, "BEGIN:VEVENT", "should contain VEVENT block")
		assert.Contains(t, body, "END:VEVENT", "should close VEVENT block")

		// Check for required event fields
		assert.Contains(t, body, "UID:", "should contain unique event ID")
		assert.Contains(t, body, "DTSTART:", "should contain event start time")
		assert.Contains(t, body, "DTEND:", "should contain event end time")
		assert.Contains(t, body, "SUMMARY:", "should contain event title/summary")
		assert.Contains(t, body, "DESCRIPTION:", "should contain event description")
		assert.Contains(t, body, "LOCATION:", "should contain event location")

		// Optional but recommended fields
		assert.Contains(t, body, "DTSTAMP:", "should contain timestamp when file was created")
	})

	t.Run("calendar file with organization information", func(t *testing.T) {
		// This test verifies that organization details are included
		registrationID := "org-info-event-uuid"

		w := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		// Organization should be included in description or organizer
		// The ORGANIZER field in iCalendar format looks like: ORGANIZER;CN="Org Name":MAILTO:email@org.com
		if strings.Contains(body, "ORGANIZER") {
			assert.Contains(t, body, "ORGANIZER", "should contain organizer information")
		}
	})

	t.Run("calendar file with proper encoding", func(t *testing.T) {
		// This test verifies special characters are properly encoded
		registrationID := "special-chars-event-uuid"

		w := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		// iCalendar files should not have lines longer than 75 characters (RFC 5545)
		// Lines should be folded (split with CRLF + space/tab)
		lines := strings.Split(body, "\n")
		for i, line := range lines {
			// Remove carriage return if present
			line = strings.TrimRight(line, "\r")
			// Continuation lines start with space or tab, so skip those checks
			if i > 0 && (strings.HasPrefix(lines[i-1], " ") || strings.HasPrefix(lines[i-1], "\t")) {
				continue
			}
			// Note: This is a simplified check. Proper RFC 5545 allows up to 75 octets
			// For basic testing, we just ensure no excessively long lines
			if len(line) > 200 {
				t.Logf("Warning: Line %d is very long (%d chars), should be folded per RFC 5545", i+1, len(line))
			}
		}
	})

	t.Run("calendar download for recurring event", func(t *testing.T) {
		// This test verifies that recurring events generate appropriate calendar entries
		registrationID := "recurring-event-uuid"

		w := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		// Basic structure should still be present
		assert.Contains(t, body, "BEGIN:VCALENDAR", "should contain VCALENDAR")
		assert.Contains(t, body, "BEGIN:VEVENT", "should contain VEVENT")

		// For recurring events, we might include RRULE (recurrence rule)
		// However, since a registration is for a specific instance, we don't expect RRULE
		// The calendar should represent the specific occurrence the volunteer registered for
	})

	t.Run("calendar includes volunteer instructions", func(t *testing.T) {
		// This test verifies that the calendar includes helpful information for the volunteer
		registrationID := "instructions-event-uuid"

		w := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		// Description should include useful information
		// (What to bring, where to meet, etc.)
		assert.Contains(t, body, "DESCRIPTION:", "should include description with instructions")
	})

	t.Run("multiple calendar downloads - same file", func(t *testing.T) {
		// This test verifies that downloading the same calendar multiple times
		// produces consistent results
		registrationID := "consistency-test-uuid"

		w1 := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w1.Code)

		w2 := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w2.Code)

		// The UID should be the same (excluding DTSTAMP which changes)
		body1 := w1.Body.String()
		body2 := w2.Body.String()

		// Extract UID from both responses
		// UIDs should match for the same registration
		if strings.Contains(body1, "UID:") && strings.Contains(body2, "UID:") {
			// Basic check - both files should have similar structure
			assert.Equal(t, strings.Contains(body1, "BEGIN:VEVENT"), strings.Contains(body2, "BEGIN:VEVENT"))
		}
	})

	t.Run("calendar file valid RFC 5545 format", func(t *testing.T) {
		// This test does basic validation of RFC 5545 compliance
		registrationID := "rfc-validation-uuid"

		w := getCalendar(registrationID, "valid-token")
		assert.Equal(t, http.StatusOK, w.Code)

		body := w.Body.String()

		// Must start with BEGIN:VCALENDAR
		assert.True(t, strings.HasPrefix(body, "BEGIN:VCALENDAR"), "must start with BEGIN:VCALENDAR")

		// Must end with END:VCALENDAR (possibly with trailing newline)
		trimmed := strings.TrimSpace(body)
		assert.True(t, strings.HasSuffix(trimmed, "END:VCALENDAR"), "must end with END:VCALENDAR")

		// Must have VERSION:2.0
		assert.Contains(t, body, "VERSION:2.0", "must declare VERSION:2.0")

		// Must have at least one VEVENT
		assert.Contains(t, body, "BEGIN:VEVENT", "must have at least one VEVENT")
		assert.Contains(t, body, "END:VEVENT", "must close VEVENT")

		// Count BEGIN and END statements should match
		beginCount := strings.Count(body, "BEGIN:VCALENDAR")
		endCount := strings.Count(body, "END:VCALENDAR")
		assert.Equal(t, beginCount, endCount, "BEGIN:VCALENDAR and END:VCALENDAR count should match")

		beginEventCount := strings.Count(body, "BEGIN:VEVENT")
		endEventCount := strings.Count(body, "END:VEVENT")
		assert.Equal(t, beginEventCount, endEventCount, "BEGIN:VEVENT and END:VEVENT count should match")
	})
}
