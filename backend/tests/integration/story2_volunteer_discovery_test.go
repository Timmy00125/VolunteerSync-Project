package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Timmy00125/VolunteerSync-Project/backend/tests/integration/helpers"
)

// TestStory2_VolunteerDiscovery tests the volunteer discovery and registration flow
// Scenarios:
// 2.1: Volunteer registration
// 2.2: Complete volunteer profile (geocode, skills, interests)
// 2.3: Search for opportunities (<2s performance)
// 2.4: Register for opportunity (notification, capacity update)
func TestStory2_VolunteerDiscovery(t *testing.T) {
	ctx := helpers.SetupTestEnvironment(t)
	defer helpers.TeardownTestEnvironment(t, ctx)

	// Prerequisites: Create organization and opportunity
	var oppID string
	var orgID string

	t.Run("Prerequisites: Setup Organization and Opportunity", func(t *testing.T) {
		// Register org admin
		ctx.RegisterUser(t, "admin@greenearth.org", "SecurePass123", "Jane", "Smith", "organization_admin")

		// Create organization
		orgResult := ctx.CreateOrganization(t, "Green Earth Initiative", "Protecting the environment")
		orgData := orgResult["data"].(map[string]interface{})
		orgID = orgData["id"].(string)

		// Create opportunity
		startDate := time.Now().Add(10 * 24 * time.Hour)
		oppResult := ctx.CreateOpportunity(t, orgID, "Beach Cleanup - Ocean Park", startDate, 20)
		oppData := oppResult["data"].(map[string]interface{})
		oppID = oppData["id"].(string)

		require.NotEmpty(t, oppID, "Opportunity should be created")
	})

	var volunteerUserID string
	var volunteerProfileID string

	t.Run("Scenario 2.1: Volunteer Registration", func(t *testing.T) {
		// Register volunteer
		result := ctx.RegisterUser(
			t,
			"john.volunteer@example.com",
			"VolunteerPass123",
			"John",
			"Doe",
			"volunteer",
		)

		// Assert: Registration successful
		require.NotNil(t, result)
		assert.NotEmpty(t, ctx.AccessToken)
		assert.NotEmpty(t, ctx.RefreshToken)

		// Assert: User account is immediately active
		user := ctx.GetUserByEmail(t, "john.volunteer@example.com")
		volunteerUserID = user["id"].(string)
		assert.Equal(t, "John", user["first_name"])
		assert.Equal(t, "Doe", user["last_name"])
		assert.Equal(t, "volunteer", user["user_type"])
		assert.Equal(t, "active", user["account_status"])

		// Assert: Volunteer profile is auto-created
		var profile map[string]interface{}
		err := ctx.DB.Table("volunteer_profiles").Where("user_id = ?", volunteerUserID).First(&profile).Error
		require.NoError(t, err, "Volunteer profile should be auto-created")
		volunteerProfileID = profile["id"].(string)

		// Profile should have default values
		assert.Equal(t, 0.0, profile["total_hours"])
		assert.Equal(t, 0, profile["total_events"])
	})

	t.Run("Scenario 2.2: Complete Volunteer Profile", func(t *testing.T) {
		// Get skill and cause IDs
		var environmentCauseID string
		var eventPlanningSkillID string

		var cause map[string]interface{}
		ctx.DB.Table("cause_categories").Where("name = ?", "Environment").First(&cause)
		environmentCauseID = cause["id"].(string)

		var skill map[string]interface{}
		ctx.DB.Table("skills").Where("name = ?", "Event Planning").First(&skill)
		eventPlanningSkillID = skill["id"].(string)

		// Update volunteer profile
		payload := map[string]interface{}{
			"location":  "San Francisco, CA",
			"biography": "Passionate about environmental conservation",
			"skills":    []string{eventPlanningSkillID},
			"interests": []string{environmentCauseID},
			"availability": map[string]interface{}{
				"saturday": true,
				"sunday":   true,
			},
			"preferred_time": "morning",
		}

		resp := ctx.MakeRequest(t, "PATCH", "/api/v1/volunteers/me", payload)
		helpers.AssertResponseStatus(t, resp, 200)

		// Assert: Profile updated in database
		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").Where("id = ?", volunteerProfileID).First(&profile)

		assert.Equal(t, "Passionate about environmental conservation", profile["biography"])
		assert.Equal(t, true, profile["availability_saturday"])
		assert.Equal(t, true, profile["availability_sunday"])
		assert.Equal(t, "morning", profile["preferred_time"])

		// Assert: Location is geocoded
		assert.NotNil(t, profile["latitude"], "Latitude should be set from geocoding")
		assert.NotNil(t, profile["longitude"], "Longitude should be set from geocoding")

		lat := profile["latitude"].(float64)
		lng := profile["longitude"].(float64)
		assert.InDelta(t, 37.7749, lat, 0.5, "Latitude should be near San Francisco")
		assert.InDelta(t, -122.4194, lng, 0.5, "Longitude should be near San Francisco")

		// Assert: Skills and interests are linked
		var skillCount int64
		ctx.DB.Table("volunteer_skills").Where("volunteer_profile_id = ?", volunteerProfileID).Count(&skillCount)
		assert.Equal(t, int64(1), skillCount, "Should have 1 skill")

		var interestCount int64
		ctx.DB.Table("volunteer_interests").Where("volunteer_profile_id = ?", volunteerProfileID).Count(&interestCount)
		assert.Equal(t, int64(1), interestCount, "Should have 1 interest")
	})

	t.Run("Scenario 2.3: Search for Opportunities", func(t *testing.T) {
		// Measure search performance (NFR-002: <2 seconds)
		startTime := time.Now()

		// Search with filters
		searchPath := fmt.Sprintf(
			"/api/v1/opportunities?location=San+Francisco,CA&radius=25&cause=environment&start_date=%s&end_date=%s",
			time.Now().Format("2006-01-02"),
			time.Now().Add(30*24*time.Hour).Format("2006-01-02"),
		)

		resp := ctx.MakeRequest(t, "GET", searchPath, nil)
		searchDuration := time.Since(startTime)

		helpers.AssertResponseStatus(t, resp, 200)

		var searchResult map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &searchResult)

		// Assert: Performance requirement met (NFR-002)
		assert.Less(t, searchDuration.Seconds(), 2.0, "Search should return results within 2 seconds")

		// Assert: Results contain our opportunity
		data, ok := searchResult["data"].([]interface{})
		require.True(t, ok, "Search should return data array")
		assert.GreaterOrEqual(t, len(data), 1, "Should find at least 1 opportunity")

		// Find "Beach Cleanup - Ocean Park" in results
		found := false
		var foundOpp map[string]interface{}
		for _, item := range data {
			oppData := item.(map[string]interface{})
			if oppData["title"] == "Beach Cleanup - Ocean Park" {
				found = true
				foundOpp = oppData
				break
			}
		}

		require.True(t, found, "Beach Cleanup opportunity should be in search results")

		// Assert: Opportunity details are correct
		assert.Equal(t, "Beach Cleanup - Ocean Park", foundOpp["title"])
		assert.Equal(t, 0, int(foundOpp["current_registrations"].(float64)))
		assert.Equal(t, 20, int(foundOpp["capacity"].(float64)))
		assert.Equal(t, "published", foundOpp["status"])

		// Assert: Opportunity has geocoded location for map display
		assert.NotNil(t, foundOpp["latitude"])
		assert.NotNil(t, foundOpp["longitude"])
	})

	var registrationID string

	t.Run("Scenario 2.4: Register for Opportunity", func(t *testing.T) {
		// Measure registration performance (NFR-006: <1 second)
		startTime := time.Now()

		// Register for opportunity
		payload := map[string]interface{}{
			"opportunity_id": oppID,
		}

		resp := ctx.MakeRequest(t, "POST", "/api/v1/registrations", payload)
		registrationDuration := time.Since(startTime)

		helpers.AssertResponseStatus(t, resp, 201)

		var result map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &result)

		// Assert: Performance requirement met (NFR-006)
		assert.Less(t, registrationDuration.Seconds(), 1.0, "Registration should complete within 1 second")

		// Extract registration ID
		data := result["data"].(map[string]interface{})
		registrationID = data["id"].(string)

		// Assert: Registration is confirmed
		assert.Equal(t, "confirmed", data["status"], "Registration should be immediately confirmed")
		assert.NotNil(t, data["registered_at"], "Registration timestamp should be set")

		// Assert: Registration exists in database
		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		assert.Equal(t, volunteerProfileID, registration["volunteer_profile_id"])
		assert.Equal(t, oppID, registration["opportunity_id"])
		assert.Equal(t, "confirmed", registration["status"])
		assert.NotNil(t, registration["registered_at"])

		// Assert: Opportunity capacity updated (FR-031)
		opp := ctx.GetOpportunityByTitle(t, "Beach Cleanup - Ocean Park")
		assert.Equal(t, 1, opp["current_registrations"], "Current registrations should increment to 1")

		// Assert: Notification created (FR-055)
		helpers.WaitForCondition(t, func() bool {
			var notifCount int64
			ctx.DB.Table("notifications").
				Where("user_id = ? AND notification_type = ?", volunteerUserID, "registration_confirmed").
				Count(&notifCount)
			return notifCount > 0
		}, 5*time.Second, "Registration confirmation notification should be created")

		var notification map[string]interface{}
		ctx.DB.Table("notifications").
			Where("user_id = ? AND notification_type = ?", volunteerUserID, "registration_confirmed").
			First(&notification)

		assert.Equal(t, "registration_confirmed", notification["notification_type"])
		assert.False(t, notification["read_at"] != nil, "Notification should be unread initially")

		// Assert: Can download calendar file (FR-064)
		calendarPath := fmt.Sprintf("/api/v1/registrations/%s/calendar.ics", registrationID)
		calResp := ctx.MakeRequest(t, "GET", calendarPath, nil)
		helpers.AssertResponseStatus(t, calResp, 200)
		assert.Equal(t, "text/calendar", calResp.Header.Get("Content-Type"), "Should return .ics file")
	})

	// Final assertions: End-to-end validation
	t.Run("End-to-End Validation", func(t *testing.T) {
		// Verify complete volunteer journey
		user := ctx.GetUserByEmail(t, "john.volunteer@example.com")

		var profile map[string]interface{}
		ctx.DB.Table("volunteer_profiles").Where("user_id = ?", user["id"]).First(&profile)

		var registration map[string]interface{}
		ctx.DB.Table("registrations").Where("id = ?", registrationID).First(&registration)

		var opportunity map[string]interface{}
		ctx.DB.Table("opportunities").Where("id = ?", oppID).First(&opportunity)

		// Verify relationships
		assert.Equal(t, user["id"], profile["user_id"])
		assert.Equal(t, profile["id"], registration["volunteer_profile_id"])
		assert.Equal(t, oppID, registration["opportunity_id"])

		// Verify capacity tracking
		assert.Equal(t, 1, opportunity["current_registrations"])
		assert.Equal(t, 20, opportunity["capacity"])

		// Verify volunteer can see their registration in API
		resp := ctx.MakeRequest(t, "GET", "/api/v1/volunteers/me/registrations", nil)
		helpers.AssertResponseStatus(t, resp, 200)

		var regList map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &regList)

		registrations := regList["data"].([]interface{})
		assert.GreaterOrEqual(t, len(registrations), 1, "Volunteer should have at least 1 registration")
	})
}
