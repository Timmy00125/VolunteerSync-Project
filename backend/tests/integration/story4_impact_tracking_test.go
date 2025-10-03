package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Timmy00125/VolunteerSync-Project/backend/tests/integration/helpers"
)

// TestStory4_ImpactTracking tests volunteer impact tracking and analytics
// Scenarios:
// 4.1: View personal dashboard with metrics
// 4.2: Achievement badge earned (First Event badge)
// 4.3: Download impact report PDF
func TestStory4_ImpactTracking(t *testing.T) {
	ctx := helpers.SetupTestEnvironment(t)
	defer helpers.TeardownTestEnvironment(t, ctx)

	// Prerequisites: Complete volunteer journey with 1 verified event
	var volunteerUserID string
	var volunteerProfileID string

	t.Run("Prerequisites: Complete Volunteer Journey", func(t *testing.T) {
		// Register org admin and create organization
		ctx.RegisterUser(t, "admin@greenearth.org", "SecurePass123", "Jane", "Smith", "organization_admin")
		orgResult := ctx.CreateOrganization(t, "Green Earth Initiative", "Protecting the environment")
		orgID := orgResult["data"].(map[string]interface{})["id"].(string)

		// Create opportunity
		startDate := time.Now().Add(10 * 24 * time.Hour)
		oppResult := ctx.CreateOpportunity(t, orgID, "Beach Cleanup - Ocean Park", startDate, 20)
		oppID := oppResult["data"].(map[string]interface{})["id"].(string)

		// Register volunteer
		ctx.RegisterUser(t, "john.volunteer@example.com", "VolunteerPass123", "John", "Doe", "volunteer")
		volunteerUser := ctx.GetUserByEmail(t, "john.volunteer@example.com")
		volunteerUserID = volunteerUser["id"].(string)

		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").Where("user_id = ?", volunteerUserID).First(&profile)
		volunteerProfileID = profile["id"].(string)

		// Register for opportunity
		regPayload := map[string]interface{}{"opportunity_id": oppID}
		regResp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", regPayload)
		var regResult map[string]interface{}
		helpers.ParseJSONResponse(t, regResp, &regResult)
		registrationID := regResult["data"].(map[string]interface{})["id"].(string)

		// Complete event: check-in, log hours, verify hours
		ctx.LoginUser(t, "admin@greenearth.org", "SecurePass123")

		// Check in
		checkInPath := fmt.Sprintf("/api/v1/registrations/%s/check-in", registrationID)
		ctx.MakeRequest(t, "PATCH", checkInPath, nil)

		// Log hours
		hoursPayload := map[string]interface{}{
			"registration_id":   registrationID,
			"hours_worked":      3.0,
			"coordinator_notes": "Excellent work",
		}
		hoursResp := ctx.MakeRequest(t, "POST", "/api/v1/hours/log", hoursPayload)
		var hoursResult map[string]interface{}
		helpers.ParseJSONResponse(t, hoursResp, &hoursResult)
		hoursLogID := hoursResult["data"].(map[string]interface{})["hours_log_id"].(string)

		// Verify hours (as volunteer)
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")
		verifyPath := fmt.Sprintf("/api/v1/hours/%s/verify", hoursLogID)
		ctx.MakeRequest(t, "POST", verifyPath, nil)

		// Verify prerequisites met
		var updatedProfile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").Where("id = ?", volunteerProfileID).First(&updatedProfile)
		require.Equal(t, 3.0, updatedProfile["total_hours"], "Should have 3.0 verified hours")
		require.Equal(t, 1, updatedProfile["total_events"], "Should have completed 1 event")
	})

	t.Run("Scenario 4.1: View Personal Dashboard", func(t *testing.T) {
		// Login as volunteer
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")

		// Get dashboard data (FR-042)
		resp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/dashboard", nil)
		helpers.AssertResponseStatus(t, resp, 200)

		var dashboard map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &dashboard)

		data := dashboard["data"].(map[string]interface{})

		// Assert: Impact metrics displayed (FR-043)
		metrics := data["metrics"].(map[string]interface{})
		assert.Equal(t, 3.0, metrics["total_hours"], "Total hours should be 3.0")
		assert.Equal(t, float64(1), metrics["events_attended"], "Events attended should be 1")
		assert.Equal(t, float64(1), metrics["organizations_supported"], "Organizations supported should be 1")

		// Assert: Recent events list
		recentEvents, ok := data["recent_events"].([]interface{})
		require.True(t, ok, "Should have recent events")
		assert.GreaterOrEqual(t, len(recentEvents), 1, "Should have at least 1 recent event")

		event := recentEvents[0].(map[string]interface{})
		assert.Equal(t, "Beach Cleanup - Ocean Park", event["title"])
		assert.Equal(t, "completed", event["status"])

		// Assert: Upcoming events list (should be empty in this test)
		upcomingEvents := data["upcoming_events"].([]interface{})
		assert.NotNil(t, upcomingEvents, "Upcoming events should not be nil")

		// Performance: Dashboard should load within 3 seconds (NFR-003)
		duration := ctx.MeasureResponseTime(t, "GET", "/api/v1/volunteers/me/dashboard", nil)
		assert.Less(t, duration.Seconds(), 3.0, "Dashboard should load within 3 seconds")
	})

	t.Run("Scenario 4.1.1: View Analytics with Chart Data", func(t *testing.T) {
		// Get analytics data for charts
		resp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/analytics", nil)
		helpers.AssertResponseStatus(t, resp, 200)

		var analytics map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &analytics)

		data := analytics["data"].(map[string]interface{})

		// Assert: Hours over time data
		hoursOverTime, ok := data["hours_over_time"].([]interface{})
		require.True(t, ok, "Should have hours over time data")
		assert.GreaterOrEqual(t, len(hoursOverTime), 1, "Should have at least 1 data point")

		// Assert: Events by cause
		eventsByCause, ok := data["events_by_cause"].(map[string]interface{})
		require.True(t, ok, "Should have events by cause data")
		assert.Contains(t, eventsByCause, "Environment", "Should have Environment cause")
	})

	t.Run("Scenario 4.2: Achievement Badge Earned", func(t *testing.T) {
		// Simulate achievement check (FR-073)
		// In production, this would be a background cron job

		// Get "First Event" achievement
		var achievement map[string]interface{}
		ctx.DB.Table("achievements").Where("name = ?", "First Event").First(&achievement)
		achievementID := achievement["id"].(string)

		// Check if volunteer qualifies (1 completed event)
		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").Where("id = ?", volunteerProfileID).First(&profile)

		if profile["total_events"].(int) >= 1 {
			// Check if already awarded
			var existingCount int64
			ctx.DB.Table("volunteer_achievements").
				Where("volunteer_profile_id = ? AND achievement_id = ?", volunteerProfileID, achievementID).
				Count(&existingCount)

			if existingCount == 0 {
				// Award achievement
				volunteerAchievement := map[string]interface{}{
					"volunteer_profile_id": volunteerProfileID,
					"achievement_id":       achievementID,
					"earned_at":            time.Now(),
				}
				ctx.DB.Table("volunteer_achievements").Create(volunteerAchievement)

				// Create notification (FR-076)
				notification := map[string]interface{}{
					"user_id":           volunteerUserID,
					"notification_type": "achievement_earned",
					"title":             "Achievement Unlocked!",
					"message":           "You earned the 'First Event' badge!",
					"sent_at":           time.Now(),
				}
				ctx.DB.Table("notifications").Create(notification)
			}
		}

		// Assert: Achievement awarded
		var awardedCount int64
		ctx.DB.Table("volunteer_achievements").
			Where("volunteer_profile_id = ? AND achievement_id = ?", volunteerProfileID, achievementID).
			Count(&awardedCount)
		assert.Equal(t, int64(1), awardedCount, "First Event badge should be awarded")

		// Assert: Achievement appears on profile (FR-074)
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")
		resp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/achievements", nil)
		helpers.AssertResponseStatus(t, resp, 200)

		var achievementsResult map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &achievementsResult)

		achievements := achievementsResult["data"].([]interface{})
		assert.GreaterOrEqual(t, len(achievements), 1, "Should have at least 1 achievement")

		firstAchievement := achievements[0].(map[string]interface{})
		assert.Equal(t, "First Event", firstAchievement["name"])
		assert.Equal(t, "🎉", firstAchievement["badge_icon"])
		assert.NotNil(t, firstAchievement["earned_at"])

		// Assert: Congratulatory notification created (FR-076)
		var notifCount int64
		ctx.DB.Table("notifications").
			Where("user_id = ? AND notification_type = ?", volunteerUserID, "achievement_earned").
			Count(&notifCount)
		assert.GreaterOrEqual(t, int(notifCount), 1, "Should have achievement notification")
	})

	t.Run("Scenario 4.3: Download Impact Report", func(t *testing.T) {
		// Download PDF impact report (FR-045)
		resp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/report?format=pdf", nil)
		helpers.AssertResponseStatus(t, resp, 200)

		// Assert: Response is PDF
		contentType := resp.Header.Get("Content-Type")
		assert.Equal(t, "application/pdf", contentType, "Should return PDF file")

		// Assert: Content-Disposition header for download
		contentDisposition := resp.Header.Get("Content-Disposition")
		assert.Contains(t, contentDisposition, "attachment", "Should be downloadable")
		assert.Contains(t, contentDisposition, "impact-report", "Filename should contain 'impact-report'")

		// Assert: PDF has content
		defer resp.Body.Close()
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		assert.Greater(t, n, 0, "PDF should have content")

		// PDF files start with %PDF
		assert.Equal(t, []byte("%PDF"), body[:4], "Should be valid PDF format")

		// Note: Full PDF content validation would require PDF parsing library
		// For now, we validate basic structure and headers
	})

	t.Run("Scenario 4.3.1: Report Contains Required Information", func(t *testing.T) {
		// Verify report data endpoint (JSON version for testing)
		resp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/report?format=json", nil)
		helpers.AssertResponseStatus(t, resp, 200)

		var report map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &report)

		data := report["data"].(map[string]interface{})

		// Assert: Report contains all required sections
		assert.Contains(t, data, "total_hours", "Report should include total hours")
		assert.Contains(t, data, "events_attended", "Report should include events attended")
		assert.Contains(t, data, "organizations_supported", "Report should include organizations")
		assert.Contains(t, data, "achievements", "Report should include achievement badges")
		assert.Contains(t, data, "date_range", "Report should include date range")

		// Assert: Values match dashboard
		assert.Equal(t, 3.0, data["total_hours"])
		assert.Equal(t, float64(1), data["events_attended"])
		assert.Equal(t, float64(1), data["organizations_supported"])

		// Assert: Achievements included
		achievements := data["achievements"].([]interface{})
		assert.GreaterOrEqual(t, len(achievements), 1, "Report should include earned achievements")
	})

	// Final assertions: End-to-end validation
	t.Run("End-to-End Validation", func(t *testing.T) {
		// Verify complete impact tracking flow
		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").Where("id = ?", volunteerProfileID).First(&profile)

		// Hours tracked accurately
		assert.Equal(t, 3.0, profile["total_hours"])
		assert.Equal(t, 1, profile["total_events"])

		// Achievement awarded
		var achievementCount int64
		ctx.DB.Table("volunteer_achievements").
			Where("volunteer_profile_id = ?", volunteerProfileID).
			Count(&achievementCount)
		assert.GreaterOrEqual(t, int(achievementCount), 1, "Should have at least 1 achievement")

		// Dashboard accessible
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")
		dashboardResp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/dashboard", nil)
		helpers.AssertResponseStatus(t, dashboardResp, 200)

		// Report downloadable
		reportResp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/report?format=pdf", nil)
		helpers.AssertResponseStatus(t, reportResp, 200)

		// Analytics available
		analyticsResp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/analytics", nil)
		helpers.AssertResponseStatus(t, analyticsResp, 200)

		// Verify user engagement metrics
		var notifCount int64
		ctx.DB.Table("notifications").
			Where("user_id = ?", volunteerUserID).
			Count(&notifCount)
		assert.GreaterOrEqual(t, int(notifCount), 2, "Should have multiple engagement notifications")
	})
}
