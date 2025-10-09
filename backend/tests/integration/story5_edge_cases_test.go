package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Timmy00125/VolunteerSync-Project/backend/tests/integration/helpers"
)

// TestStory5_EdgeCases tests edge cases and error handling
// Scenarios:
// 5.1: Event at capacity (waitlist functionality)
// 5.2: Late cancellation (warning within 24 hours)
// 5.3: Overlapping events (warning with override)
// 5.4: Hours dispute workflow
func TestStory5_EdgeCases(t *testing.T) {
	ctx := helpers.SetupTestEnvironment(t)
	defer helpers.TeardownTestEnvironment(t, ctx)

	// Prerequisites: Setup organization and opportunities
	var orgID string
	var oppID string
	var adminUserID string

	t.Run("Prerequisites: Setup Organization", func(t *testing.T) {
		// Register org admin
		ctx.RegisterUser(t, "admin@greenearth.org", "SecurePass123", "Jane", "Smith", "organization_admin")
		adminUser := ctx.GetUserByEmail(t, "admin@greenearth.org")
		adminUserID = adminUser["id"].(string)

		// Create organization
		orgResult := ctx.CreateOrganization(t, "Green Earth Initiative", "Protecting the environment")
		orgData := orgResult["data"].(map[string]interface{})
		orgID = orgData["id"].(string)

		require.NotEmpty(t, orgID)
	})

	t.Run("Scenario 5.1: Event at Capacity", func(t *testing.T) {
		// Create opportunity with capacity of 2
		startDate := time.Now().Add(10 * 24 * time.Hour)
		oppResult := ctx.CreateOpportunity(t, orgID, "Small Beach Cleanup", startDate, 2)
		oppData := oppResult["data"].(map[string]interface{})
		oppID = oppData["id"].(string)

		// Register first volunteer
		ctx.RegisterUser(t, "volunteer1@example.com", "Pass123", "Alice", "Smith", "volunteer")
		reg1Payload := map[string]interface{}{"opportunity_id": oppID}
		reg1Resp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", reg1Payload)
		helpers.AssertResponseStatus(t, reg1Resp, 201)

		var reg1Result map[string]interface{}
		helpers.ParseJSONResponse(t, reg1Resp, &reg1Result)
		assert.Equal(t, "confirmed", reg1Result["data"].(map[string]interface{})["status"])

		// Register second volunteer
		ctx.RegisterUser(t, "volunteer2@example.com", "Pass123", "Bob", "Jones", "volunteer")
		reg2Payload := map[string]interface{}{"opportunity_id": oppID}
		reg2Resp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", reg2Payload)
		helpers.AssertResponseStatus(t, reg2Resp, 201)

		var reg2Result map[string]interface{}
		helpers.ParseJSONResponse(t, reg2Resp, &reg2Result)
		assert.Equal(t, "confirmed", reg2Result["data"].(map[string]interface{})["status"])

		// Assert: Opportunity is now at capacity
		opp := ctx.GetOpportunityByTitle(t, "Small Beach Cleanup")
		assert.Equal(t, 2, opp["current_registrations"], "Should have 2 registrations")
		assert.Equal(t, 2, opp["capacity"], "Capacity is 2")

		// Register third volunteer (should be waitlisted - FR-030, FR-035)
		ctx.RegisterUser(t, "volunteer3@example.com", "Pass123", "Charlie", "Brown", "volunteer")
		volunteer3User := ctx.GetUserByEmail(t, "volunteer3@example.com")
		volunteer3UserID := volunteer3User["id"].(string)

		reg3Payload := map[string]interface{}{"opportunity_id": oppID}
		reg3Resp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", reg3Payload)

		// Should succeed but with waitlisted status
		helpers.AssertResponseStatus(t, reg3Resp, 201)

		var reg3Result map[string]interface{}
		helpers.ParseJSONResponse(t, reg3Resp, &reg3Result)
		reg3Data := reg3Result["data"].(map[string]interface{})

		// Assert: Third volunteer is waitlisted
		assert.Equal(t, "waitlisted", reg3Data["status"], "Third volunteer should be waitlisted")
		assert.NotNil(t, reg3Data["waitlisted_at"], "Waitlist timestamp should be set")

		// Assert: Opportunity still shows 2 registrations (confirmed only)
		opp = ctx.GetOpportunityByTitle(t, "Small Beach Cleanup")
		assert.Equal(t, 2, opp["current_registrations"], "Should still show 2 confirmed registrations")

		// Assert: Waitlist notification sent
		helpers.WaitForCondition(t, func() bool {
			var notifCount int64
			ctx.DB.Table("notifications").
				Where("user_id = ? AND notification_type = ?", volunteer3UserID, "waitlisted").
				Count(&notifCount)
			return notifCount > 0
		}, 5*time.Second, "Waitlist notification should be created")

		// Assert: Display shows "Full (2 of 2)"
		resp := ctx.MakeRequest(t, "GET", fmt.Sprintf("/api/v1/opportunities/%s", oppID), nil)
		helpers.AssertResponseStatus(t, resp, 200)

		var oppDetail map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &oppDetail)
		oppDetailData := oppDetail["data"].(map[string]interface{})

		assert.Equal(t, true, oppDetailData["is_full"], "Opportunity should be marked as full")
		assert.Equal(t, float64(2), oppDetailData["current_registrations"])
		assert.Equal(t, float64(2), oppDetailData["capacity"])
	})

	t.Run("Scenario 5.2: Late Cancellation", func(t *testing.T) {
		// Create opportunity starting in 12 hours
		startDate := time.Now().Add(12 * time.Hour)
		opp2Result := ctx.CreateOpportunity(t, orgID, "Urgent Cleanup", startDate, 10)
		opp2ID := opp2Result["data"].(map[string]interface{})["id"].(string)

		// Register volunteer
		ctx.RegisterUser(t, "latecancel@example.com", "Pass123", "Late", "Cancel", "volunteer")
		// lateCancelUser := ctx.GetUserByEmail(t, "latecancel@example.com")
		// lateCancelUserID := lateCancelUser["id"].(string)

		regPayload := map[string]interface{}{"opportunity_id": opp2ID}
		regResp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", regPayload)
		var regResult map[string]interface{}
		helpers.ParseJSONResponse(t, regResp, &regResult)
		registrationID := regResult["data"].(map[string]interface{})["id"].(string)

		// Cancel registration (FR-034: late cancellation warning)
		cancelPayload := map[string]interface{}{
			"reason": "Personal emergency",
		}
		cancelPath := fmt.Sprintf("/api/v1/registrations/%s/cancel", registrationID)
		cancelResp := ctx.MakeRequest(t, "PATCH", cancelPath, cancelPayload)
		helpers.AssertResponseStatus(t, cancelResp, 200)

		var cancelResult map[string]interface{}
		helpers.ParseJSONResponse(t, cancelResp, &cancelResult)

		// Assert: Late cancellation warning returned
		assert.Contains(t, cancelResult, "warning", "Should include late cancellation warning")
		warning := cancelResult["warning"].(map[string]interface{})
		assert.Equal(t, "late_cancellation", warning["type"])
		assert.Contains(t, warning["message"], "within 24 hours", "Warning should mention 24 hour policy")

		// Assert: Cancellation recorded
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		assert.Equal(t, "cancelled", registration["status"])
		assert.NotNil(t, registration["cancelled_at"])
		assert.Equal(t, "Personal emergency", registration["cancellation_reason"])

		// Assert: Opportunity capacity decremented
		opp := ctx.GetOpportunityByTitle(t, "Urgent Cleanup")
		assert.Equal(t, 0, opp["current_registrations"], "Should decrement back to 0")

		// Assert: Organization notified immediately
		helpers.WaitForCondition(t, func() bool {
			var notifCount int64
			ctx.DB.Table("notifications").
				Where("user_id = ? AND notification_type = ?", adminUserID, "registration_cancelled").
				Count(&notifCount)
			return notifCount > 0
		}, 5*time.Second, "Org admin should receive cancellation notification")

		// Assert: Late cancellation flag set
		assert.True(t, registration["is_late_cancellation"].(bool), "Should be flagged as late cancellation")
	})

	t.Run("Scenario 5.3: Overlapping Events", func(t *testing.T) {
		// Register volunteer
		ctx.RegisterUser(t, "overlap@example.com", "Pass123", "Over", "Lap", "volunteer")

		// Create Event A: 9:00 AM - 12:00 PM
		eventAStart := time.Now().Add(10 * 24 * time.Hour).Truncate(24 * time.Hour).Add(9 * time.Hour)
		eventAResult := ctx.CreateOpportunity(t, orgID, "Morning Event", eventAStart, 10)
		eventAID := eventAResult["data"].(map[string]interface{})["id"].(string)

		// Register for Event A
		regAPayload := map[string]interface{}{"opportunity_id": eventAID}
		regAResp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", regAPayload)
		helpers.AssertResponseStatus(t, regAResp, 201)

		// Create Event B: 11:00 AM - 2:00 PM (overlaps with Event A)
		eventBStart := time.Now().Add(10 * 24 * time.Hour).Truncate(24 * time.Hour).Add(11 * time.Hour)
		eventBResult := ctx.CreateOpportunity(t, orgID, "Midday Event", eventBStart, 10)
		eventBID := eventBResult["data"].(map[string]interface{})["id"].(string)

		// Attempt to register for Event B (FR-032: overlapping warning)
		regBPayload := map[string]interface{}{"opportunity_id": eventBID}
		regBResp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", regBPayload)

		var regBResult map[string]interface{}
		helpers.ParseJSONResponse(t, regBResp, &regBResult)

		// Assert: Overlap warning returned
		assert.Contains(t, regBResult, "warning", "Should include overlap warning")
		warning := regBResult["warning"].(map[string]interface{})
		assert.Equal(t, "overlapping_events", warning["type"])
		assert.Contains(t, warning["message"], "overlap", "Warning should mention overlap")
		assert.Contains(t, warning, "conflicting_event", "Should show conflicting event details")

		// Assert: Registration still created (user can override)
		regBData := regBResult["data"].(map[string]interface{})
		assert.Equal(t, "confirmed", regBData["status"], "Should still be confirmed with warning")

		// Verify both registrations exist
		var regCount int64
		var volunteerProfile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").
			Joins("JOIN users ON users.id = volunteer_profiles.user_id").
			Where("users.email = ?", "overlap@example.com").
			First(&volunteerProfile)

		ctx.DB.Table("registrations").
			Where("volunteer_profile_id = ? AND status = ?", volunteerProfile["id"], "confirmed").
			Count(&regCount)
		assert.Equal(t, int64(2), regCount, "Should have 2 confirmed registrations despite overlap")
	})

	t.Run("Scenario 5.4: Hours Dispute", func(t *testing.T) {
		// Create opportunity and complete flow up to hours logging
		startDate := time.Now().Add(5 * 24 * time.Hour)
		opp3Result := ctx.CreateOpportunity(t, orgID, "Dispute Test Event", startDate, 10)
		opp3ID := opp3Result["data"].(map[string]interface{})["id"].(string)

		// Register volunteer
		ctx.RegisterUser(t, "dispute@example.com", "Pass123", "Dis", "Pute", "volunteer")
		disputeUser := ctx.GetUserByEmail(t, "dispute@example.com")
		disputeUserID := disputeUser["id"].(string)

		regPayload := map[string]interface{}{"opportunity_id": opp3ID}
		regResp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", regPayload)
		var regResult map[string]interface{}
		helpers.ParseJSONResponse(t, regResp, &regResult)
		registrationID := regResult["data"].(map[string]interface{})["id"].(string)

		// Admin logs hours (3.0 hours)
		ctx.LoginUser(t, "admin@greenearth.org", "SecurePass123")
		hoursPayload := map[string]interface{}{
			"registration_id":   registrationID,
			"hours_worked":      3.0,
			"coordinator_notes": "Arrived late, left early",
		}
		hoursResp := ctx.MakeRequest(t, "POST", "/api/v1/hours/log", hoursPayload)
		var hoursResult map[string]interface{}
		helpers.ParseJSONResponse(t, hoursResp, &hoursResult)
		hoursLogID := hoursResult["data"].(map[string]interface{})["hours_log_id"].(string)

		// Volunteer disputes hours (FR-050)
		ctx.LoginUser(t, "dispute@example.com", "Pass123")
		disputePayload := map[string]interface{}{
			"reason":        "I worked 4 hours, not 3. I arrived on time and stayed until the end.",
			"claimed_hours": 4.0,
		}
		disputePath := fmt.Sprintf("/api/v1/hours/%s/dispute", hoursLogID)
		disputeResp := ctx.MakeRequest(t, "POST", disputePath, disputePayload)
		helpers.AssertResponseStatus(t, disputeResp, 200)

		// Assert: Hours status changed to disputed
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		assert.Equal(t, "disputed", registration["hours_status"], "Hours should be in disputed status")
		assert.NotNil(t, registration["hours_disputed_at"], "Dispute timestamp should be set")
		assert.Equal(t, "I worked 4 hours, not 3. I arrived on time and stayed until the end.",
			registration["dispute_reason"])
		assert.Equal(t, 4.0, registration["claimed_hours"])

		// Assert: Hours_Log updated
		var hoursLog map[string]interface{}
		ctx.DB.Table("hours_logs").Where("id = ?", hoursLogID).First(&hoursLog)

		assert.Equal(t, "disputed", hoursLog["status"])
		assert.NotNil(t, hoursLog["disputed_at"])

		// Assert: Hours NOT added to volunteer profile (dispute pending)
		var volunteerProfile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").
			Joins("JOIN users ON users.id = volunteer_profiles.user_id").
			Where("users.id = ?", disputeUserID).
			First(&volunteerProfile)

		assert.Equal(t, 0.0, volunteerProfile["total_hours"], "Total hours should remain 0 during dispute")

		// Assert: Coordinator notified of dispute
		helpers.WaitForCondition(t, func() bool {
			var notifCount int64
			ctx.DB.Table("notifications").
				Where("user_id = ? AND notification_type = ?", adminUserID, "hours_disputed").
				Count(&notifCount)
			return notifCount > 0
		}, 5*time.Second, "Coordinator should receive dispute notification")

		var notification map[string]interface{}
		ctx.DB.Table("notifications").
			Where("user_id = ? AND notification_type = ?", adminUserID, "hours_disputed").
			First(&notification)

		assert.Contains(t, notification["message"], "dispute", "Notification should mention dispute")
		assert.Contains(t, notification, "dispute_reason", "Should include dispute reason")

		// Assert: Dispute resolution workflow initiated
		var disputeRecord map[string]interface{}
		ctx.DB.Table("hours_logs").Where("id = ?", hoursLogID).First(&disputeRecord)
		assert.Equal(t, "disputed", disputeRecord["status"])
		assert.NotEmpty(t, registration["dispute_reason"], "Dispute reason should be stored")

		// Simulate coordinator resolving dispute (updating hours to 4.0)
		ctx.LoginUser(t, "admin@greenearth.org", "SecurePass123")
		resolvePayload := map[string]interface{}{
			"resolution":       "verified",
			"final_hours":      4.0,
			"resolution_notes": "After review, volunteer is correct - worked full 4 hours",
		}
		resolvePath := fmt.Sprintf("/api/v1/hours/%s/resolve-dispute", hoursLogID)
		resolveResp := ctx.MakeRequest(t, "POST", resolvePath, resolvePayload)
		helpers.AssertResponseStatus(t, resolveResp, 200)

		// Assert: Dispute resolved
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)
		assert.Equal(t, "verified", registration["hours_status"], "Should be verified after resolution")
		assert.Equal(t, 4.0, registration["hours_worked"], "Hours should be updated to 4.0")
		assert.NotNil(t, registration["hours_verified_at"])

		// Assert: Hours added to profile
		ctx.DB.Table("volunteer_profiles").
			Joins("JOIN users ON users.id = volunteer_profiles.user_id").
			Where("users.id = ?", disputeUserID).
			First(&volunteerProfile)

		assert.Equal(t, 4.0, volunteerProfile["total_hours"], "Total hours should now be 4.0")
	})

	// Final assertions: End-to-end validation
	t.Run("End-to-End Validation: All Edge Cases Handled", func(t *testing.T) {
		// Verify waitlist functionality
		var waitlistedCount int64
		ctx.DB.Table("registrations").Where("status = ?", "waitlisted").Count(&waitlistedCount)
		assert.GreaterOrEqual(t, int(waitlistedCount), 1, "Should have at least 1 waitlisted registration")

		// Verify late cancellations tracked
		var lateCancelCount int64
		ctx.DB.Table("registrations").
			Where("status = ? AND is_late_cancellation = ?", "cancelled", true).
			Count(&lateCancelCount)
		assert.GreaterOrEqual(t, int(lateCancelCount), 1, "Should have at least 1 late cancellation")

		// Verify overlapping event warnings work
		var overlapUser map[string]interface{}
		ctx.DB.Table("users").Where("email = ?", "overlap@example.com").First(&overlapUser)

		var overlapProfile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").Where("user_id = ?", overlapUser["id"]).First(&overlapProfile)

		var overlapRegCount int64
		ctx.DB.Table("registrations").
			Where("volunteer_profile_id = ? AND status = ?", overlapProfile["id"], "confirmed").
			Count(&overlapRegCount)
		assert.Equal(t, int64(2), overlapRegCount, "Overlapping registrations should both exist")

		// Verify dispute resolution works
		var resolvedDisputeCount int64
		ctx.DB.Table("hours_logs").
			Where("status = ? AND disputed_at IS NOT NULL", "verified").
			Count(&resolvedDisputeCount)
		assert.GreaterOrEqual(t, int(resolvedDisputeCount), 1, "Should have at least 1 resolved dispute")

		// Verify all edge case notifications sent
		var edgeCaseNotifCount int64
		ctx.DB.Table("notifications").
			Where("notification_type IN ?", []string{
				"waitlisted",
				"registration_cancelled",
				"hours_disputed",
			}).
			Count(&edgeCaseNotifCount)
		assert.GreaterOrEqual(t, int(edgeCaseNotifCount), 3, "Should have edge case notifications")
	})
}
