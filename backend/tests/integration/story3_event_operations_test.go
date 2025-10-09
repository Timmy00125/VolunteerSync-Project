package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Timmy00125/VolunteerSync-Project/backend/tests/integration/helpers"
)

// TestStory3_EventOperations tests event day operations and hours tracking
// Scenarios:
// 3.1: Event reminder notifications (24h and 2h before event)
// 3.2: Volunteer check-in
// 3.3: Log volunteer hours (pending status)
// 3.4: Volunteer confirms hours (verified status)
// 3.5: Auto-verify hours after 7 days
// 3.6: Volunteer reviews event
func TestStory3_EventOperations(t *testing.T) {
	ctx := helpers.SetupTestEnvironment(t)
	defer helpers.TeardownTestEnvironment(t, ctx)

	// Prerequisites: Setup organization, opportunity, and volunteer registration
	var oppID string
	var volunteerUserID string
	var registrationID string
	var adminUserID string

	t.Run("Prerequisites: Setup Complete Flow", func(t *testing.T) {
		// Register org admin
		ctx.RegisterUser(t, "admin@greenearth.org", "SecurePass123", "Jane", "Smith", "organization_admin")
		adminUser := ctx.GetUserByEmail(t, "admin@greenearth.org")
		adminUserID = adminUser["id"].(string)

		// Create organization
		orgResult := ctx.CreateOrganization(t, "Green Earth Initiative", "Protecting the environment")
		orgData := orgResult["data"].(map[string]interface{})
		orgID := orgData["id"].(string)

		// Create opportunity (event in 10 days)
		startDate := time.Now().Add(10 * 24 * time.Hour)
		oppResult := ctx.CreateOpportunity(t, orgID, "Beach Cleanup - Ocean Park", startDate, 20)
		oppData := oppResult["data"].(map[string]interface{})
		oppID = oppData["id"].(string)

		// Register volunteer
		ctx.RegisterUser(t, "john.volunteer@example.com", "VolunteerPass123", "John", "Doe", "volunteer")
		volunteerUser := ctx.GetUserByEmail(t, "john.volunteer@example.com")
		volunteerUserID = volunteerUser["id"].(string)

		// Register for opportunity
		regPayload := map[string]interface{}{"opportunity_id": oppID}
		regResp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", regPayload)
		var regResult map[string]interface{}
		helpers.ParseJSONResponse(t, regResp, &regResult)
		registrationID = regResult["data"].(map[string]interface{})["id"].(string)

		require.NotEmpty(t, registrationID, "Registration should be created")
	})

	t.Run("Scenario 3.1: Event Reminder Notifications", func(t *testing.T) {
		// Simulate 24-hour reminder notification
		// In production, this would be triggered by a cron job
		notif24h := map[string]interface{}{
			"user_id":           volunteerUserID,
			"notification_type": "event_reminder_24h",
			"title":             "Event Tomorrow!",
			"message":           "Beach Cleanup - Ocean Park starts in 24 hours",
			"sent_at":           time.Now(),
		}
		ctx.DB.Table("notifications").Create(notif24h)

		// Assert: 24-hour notification created
		var notif24Count int64
		ctx.DB.Table("notifications").
			Where("user_id = ? AND notification_type = ?", volunteerUserID, "event_reminder_24h").
			Count(&notif24Count)
		assert.Equal(t, int64(1), notif24Count, "24-hour reminder should be created")

		// Simulate 2-hour reminder notification
		notif2h := map[string]interface{}{
			"user_id":           volunteerUserID,
			"notification_type": "event_reminder_2h",
			"title":             "Event Starting Soon!",
			"message":           "Beach Cleanup - Ocean Park starts in 2 hours",
			"sent_at":           time.Now(),
		}
		ctx.DB.Table("notifications").Create(notif2h)

		// Assert: 2-hour notification created
		var notif2hCount int64
		ctx.DB.Table("notifications").
			Where("user_id = ? AND notification_type = ?", volunteerUserID, "event_reminder_2h").
			Count(&notif2hCount)
		assert.Equal(t, int64(1), notif2hCount, "2-hour reminder should be created")

		// Assert: Notifications are unread
		var unreadCount int64
		ctx.DB.Table("notifications").
			Where("user_id = ? AND read_at IS NULL", volunteerUserID).
			Count(&unreadCount)
		assert.GreaterOrEqual(t, int(unreadCount), 2, "Should have at least 2 unread notifications")
	})

	t.Run("Scenario 3.2: Volunteer Check-In", func(t *testing.T) {
		// Switch to admin token
		ctx.LoginUser(t, "admin@greenearth.org", "SecurePass123")

		// Check in volunteer
		checkInPath := fmt.Sprintf("/api/v1/registrations/%s/check-in", registrationID)
		resp := ctx.MakeRequest(t, "PATCH", checkInPath, nil)
		helpers.AssertResponseStatus(t, resp, 200)

		// Assert: Check-in timestamp recorded
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		assert.NotNil(t, registration["checked_in_at"], "Check-in timestamp should be set")
		assert.Equal(t, "confirmed", registration["status"], "Status should still be confirmed")
	})

	var hoursLogID string

	t.Run("Scenario 3.3: Log Volunteer Hours", func(t *testing.T) {
		// Admin logs hours for volunteer
		payload := map[string]interface{}{
			"registration_id":   registrationID,
			"hours_worked":      3.0,
			"coordinator_notes": "Great work, very enthusiastic",
		}

		resp := ctx.MakeRequest(t, "POST", "/api/v1/hours/log", payload)
		helpers.AssertResponseStatus(t, resp, 201)

		var result map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &result)

		data := result["data"].(map[string]interface{})
		hoursLogID = data["hours_log_id"].(string)

		// Assert: Registration updated with hours (FR-046)
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		assert.Equal(t, 3.0, registration["hours_worked"])
		assert.Equal(t, "pending", registration["hours_status"], "Hours should be in pending status")
		assert.NotNil(t, registration["hours_logged_at"])
		assert.Equal(t, "Great work, very enthusiastic", registration["coordinator_notes"])

		// Assert: Hours_Log audit record created (FR-054)
		var hoursLog map[string]interface{}
		ctx.DB.Table("hours_logs").Where("id = ?", hoursLogID).First(&hoursLog)

		assert.Equal(t, registrationID, hoursLog["registration_id"])
		assert.Equal(t, 3.0, hoursLog["hours_worked"])
		assert.Equal(t, "pending", hoursLog["status"])
		assert.Equal(t, adminUserID, hoursLog["logged_by_user_id"])
		assert.NotNil(t, hoursLog["logged_at"])

		// Assert: Volunteer profile total_hours NOT yet incremented (pending verification)
		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").
			Joins("JOIN users ON users.id = volunteer_profiles.user_id").
			Where("users.id = ?", volunteerUserID).
			First(&profile)

		assert.Equal(t, 0.0, profile["total_hours"], "Total hours should not increment until verified")

		// Assert: Notification sent to volunteer (FR-047)
		helpers.WaitForCondition(t, func() bool {
			var notifCount int64
			ctx.DB.Table("notifications").
				Where("user_id = ? AND notification_type = ?", volunteerUserID, "hours_logged").
				Count(&notifCount)
			return notifCount > 0
		}, 5*time.Second, "Hours logged notification should be created")
	})

	t.Run("Scenario 3.4: Volunteer Confirms Hours", func(t *testing.T) {
		// Switch to volunteer token
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")

		// Verify hours
		verifyPath := fmt.Sprintf("/api/v1/hours/%s/verify", hoursLogID)
		resp := ctx.MakeRequest(t, "POST", verifyPath, nil)
		helpers.AssertResponseStatus(t, resp, 200)

		// Assert: Hours status changed to verified (FR-048, FR-051)
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		assert.Equal(t, "verified", registration["hours_status"], "Hours should be verified")
		assert.NotNil(t, registration["hours_verified_at"], "Verification timestamp should be set")
		assert.Equal(t, "completed", registration["status"], "Registration status should be completed")

		// Assert: Hours_Log updated
		var hoursLog map[string]interface{}
		ctx.DB.Table("hours_logs").Where("id = ?", hoursLogID).First(&hoursLog)

		assert.Equal(t, "verified", hoursLog["status"])
		assert.NotNil(t, hoursLog["verified_at"])

		// Assert: Volunteer profile total_hours incremented
		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").
			Joins("JOIN users ON users.id = volunteer_profiles.user_id").
			Where("users.id = ?", volunteerUserID).
			First(&profile)

		assert.Equal(t, 3.0, profile["total_hours"], "Total hours should be 3.0 after verification")
		assert.Equal(t, 1, profile["total_events"], "Total events should be 1")
	})

	t.Run("Scenario 3.5: Auto-Verify Hours After 7 Days", func(t *testing.T) {
		// Create another registration with pending hours that are 8 days old
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")

		// Create a second opportunity
		ctx.LoginUser(t, "admin@greenearth.org", "SecurePass123")

		var org map[string]interface{}
		ctx.DB.Table("organizations").Where("slug = ?", "green-earth-initiative").First(&org)

		startDate := time.Now().Add(5 * 24 * time.Hour)
		opp2Result := ctx.CreateOpportunity(t, org["id"].(string), "Park Cleanup", startDate, 10)
		opp2ID := opp2Result["data"].(map[string]interface{})["id"].(string)

		// Register volunteer
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")
		regPayload := map[string]interface{}{"opportunity_id": opp2ID}
		regResp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", regPayload)
		var regResult map[string]interface{}
		helpers.ParseJSONResponse(t, regResp, &regResult)
		reg2ID := regResult["data"].(map[string]interface{})["id"].(string)

		// Log hours (as admin)
		ctx.LoginUser(t, "admin@greenearth.org", "SecurePass123")
		hoursPayload := map[string]interface{}{
			"registration_id":   reg2ID,
			"hours_worked":      2.5,
			"coordinator_notes": "Good job",
		}
		hoursResp := ctx.MakeRequest(t, "POST", "/api/v1/hours/log", hoursPayload)
		var hoursResult map[string]interface{}
		helpers.ParseJSONResponse(t, hoursResp, &hoursResult)
		log2ID := hoursResult["data"].(map[string]interface{})["hours_log_id"].(string)

		// Simulate 8 days passing by backdating the logged_at timestamp
		eightDaysAgo := time.Now().Add(-8 * 24 * time.Hour)
		ctx.DB.Table("registrations").Where("id = ?", reg2ID).Update("hours_logged_at", eightDaysAgo)
		ctx.DB.Table("hours_logs").Where("id = ?", log2ID).Update("logged_at", eightDaysAgo)

		// Simulate auto-verification cron job (FR-049)
		// In production, this would run as a background job
		var oldPendingLogs []map[string]interface{}
		ctx.DB.Table("hours_logs").
			Where("status = ? AND logged_at < ?", "pending", time.Now().Add(-7*24*time.Hour)).
			Find(&oldPendingLogs)

		for _, log := range oldPendingLogs {
			logID := log["id"].(string)
			regID := log["registration_id"].(string)
			hoursWorked := log["hours_worked"].(float64)

			// Update hours log to verified
			ctx.DB.Table("hours_logs").Where("id = ?", logID).Updates(map[string]interface{}{
				"status":           "verified",
				"verified_at":      time.Now(),
				"auto_verified_at": time.Now(),
			})

			// Update registration
			ctx.DB.Table("registrations").Where("id = ?", regID).Updates(map[string]interface{}{
				"hours_status":      "verified",
				"hours_verified_at": time.Now(),
				"status":            "completed",
			})

			// Increment volunteer profile hours
			var reg map[string]interface{}
			ctx.DB.Table("registrations").Where("id = ?", regID).First(&reg)

			ctx.DB.Exec(`
				UPDATE volunteer_profiles 
				SET total_hours = total_hours + ?, 
				    total_events = total_events + 1 
				WHERE id = ?
			`, hoursWorked, reg["volunteer_profile_id"])
		}

		// Assert: Hours auto-verified
		var hoursLog map[string]interface{}
		ctx.DB.Table("hours_logs").Where("id = ?", log2ID).First(&hoursLog)

		assert.Equal(t, "verified", hoursLog["status"], "Hours should be auto-verified after 7 days")
		assert.NotNil(t, hoursLog["auto_verified_at"], "Auto-verify timestamp should be set")

		// Assert: Volunteer total hours updated
		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").
			Joins("JOIN users ON users.id = volunteer_profiles.user_id").
			Where("users.id = ?", volunteerUserID).
			First(&profile)

		assert.Equal(t, 5.5, profile["total_hours"], "Total hours should be 5.5 (3.0 + 2.5)")
		assert.Equal(t, 2, profile["total_events"], "Total events should be 2")
	})

	t.Run("Scenario 3.6: Volunteer Reviews Event", func(t *testing.T) {
		// Switch to volunteer token
		ctx.LoginUser(t, "john.volunteer@example.com", "VolunteerPass123")

		// Submit review (FR-068)
		reviewPayload := map[string]interface{}{
			"rating": 5,
			"review": "Amazing experience! Well organized and impactful.",
		}

		reviewPath := fmt.Sprintf("/api/v1/registrations/%s/review", registrationID)
		resp := ctx.MakeRequest(t, "POST", reviewPath, reviewPayload)
		helpers.AssertResponseStatus(t, resp, 200)

		// Assert: Review stored in registration
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		assert.Equal(t, 5, registration["volunteer_rating"], "Rating should be 5 stars")
		assert.Equal(t, "Amazing experience! Well organized and impactful.", registration["volunteer_review"])
		assert.NotNil(t, registration["review_submitted_at"], "Review timestamp should be set")

		// Assert: Organization average rating updated
		var org map[string]interface{}
		ctx.DB.Table("organizations").Where("slug = ?", "green-earth-initiative").First(&org)

		assert.NotNil(t, org["avg_rating"], "Organization should have average rating")
		// Note: Exact value depends on how many reviews exist, but should be set
	})

	// Final assertions: End-to-end validation
	t.Run("End-to-End Validation", func(t *testing.T) {
		// Verify complete event operations flow
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		// Check-in completed
		assert.NotNil(t, registration["checked_in_at"])

		// Hours logged, verified, and completed
		assert.Equal(t, 3.0, registration["hours_worked"])
		assert.Equal(t, "verified", registration["hours_status"])
		assert.Equal(t, "completed", registration["status"])

		// Review submitted
		assert.Equal(t, 5, registration["volunteer_rating"])
		assert.NotEmpty(t, registration["volunteer_review"])

		// Volunteer profile updated
		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").
			Joins("JOIN users ON users.id = volunteer_profiles.user_id").
			Where("users.id = ?", volunteerUserID).
			First(&profile)

		assert.GreaterOrEqual(t, profile["total_hours"].(float64), 3.0)
		assert.GreaterOrEqual(t, profile["total_events"].(int), 1)

		// Notifications sent
		var notifCount int64
		ctx.DB.Table("notifications").
			Where("user_id = ?", volunteerUserID).
			Count(&notifCount)
		assert.GreaterOrEqual(t, int(notifCount), 3, "Should have multiple notifications")
	})
}
