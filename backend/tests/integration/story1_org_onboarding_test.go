package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Timmy00125/VolunteerSync-Project/backend/tests/integration/helpers"
)

// TestStory1_OrgOnboarding tests the complete organization onboarding flow
// Scenarios:
// 1.1: Organization admin registration
// 1.2: Create organization profile (auto-verify, geocode)
// 1.3: Create and publish volunteer opportunity
func TestStory1_OrgOnboarding(t *testing.T) {
	ctx := helpers.SetupTestEnvironment(t)
	defer helpers.TeardownTestEnvironment(t, ctx)

	t.Run("Scenario 1.1: Organization Admin Registration", func(t *testing.T) {
		// Register organization admin
		result := ctx.RegisterUser(
			t,
			"admin@greenearth.org",
			"SecurePass123",
			"Jane",
			"Smith",
			"organization_admin",
		)

		// Assert: Registration successful
		require.NotNil(t, result)
		assert.NotEmpty(t, ctx.AccessToken, "Access token should be provided")
		assert.NotEmpty(t, ctx.RefreshToken, "Refresh token should be provided")

		// Assert: User exists in database with active status
		user := ctx.GetUserByEmail(t, "admin@greenearth.org")
		assert.Equal(t, "Jane", user["first_name"])
		assert.Equal(t, "Smith", user["last_name"])
		assert.Equal(t, "organization_admin", user["user_type"])
		assert.Equal(t, "active", user["account_status"], "Account should be immediately active (FR-002)")
		assert.NotNil(t, user["last_login_at"], "Last login timestamp should be set")
		assert.NotEmpty(t, user["security_question1"], "Security questions should be stored")
		assert.NotEmpty(t, user["security_question2"])
		assert.NotEmpty(t, user["security_question3"])

		// Assert: Security answers are hashed (not plaintext)
		assert.NotEqual(t, "Johnson", user["security_answer1_hash"], "Security answers should be hashed")
		assert.NotEqual(t, "Boston", user["security_answer2_hash"])
		assert.NotEqual(t, "Max", user["security_answer3_hash"])

		// Assert: Password is hashed with Argon2
		assert.NotEqual(t, "SecurePass123", user["password_hash"], "Password should be hashed")
		assert.Contains(t, user["password_hash"], "$argon2", "Should use Argon2 hashing")
	})

	var orgID string
	var orgSlug string

	t.Run("Scenario 1.2: Create Organization Profile", func(t *testing.T) {
		// Create organization
		result := ctx.CreateOrganization(
			t,
			"Green Earth Initiative",
			"Protecting the environment through action",
		)

		// Extract organization ID from response
		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "Response should have data field")
		orgID = data["id"].(string)
		orgSlug = data["slug"].(string)

		// Assert: Organization has auto-generated slug
		assert.Equal(t, "green-earth-initiative", orgSlug, "Slug should be auto-generated from name")

		// Assert: Organization is auto-verified (FR-015)
		assert.Equal(t, "verified", data["verification_status"], "Organization should be auto-verified")
		assert.NotNil(t, data["verified_at"], "Verified timestamp should be set")

		// Assert: Organization exists in database with geocoding
		org := ctx.GetOrganizationBySlug(t, orgSlug)
		assert.Equal(t, "Green Earth Initiative", org["name"])
		assert.Equal(t, "Protecting the environment through action", org["mission_statement"])
		assert.Equal(t, "verified", org["verification_status"])
		assert.NotNil(t, org["latitude"], "Latitude should be set from geocoding")
		assert.NotNil(t, org["longitude"], "Longitude should be set from geocoding")

		// Assert: Coordinates are approximately Portland, Oregon
		lat := org["latitude"].(float64)
		lng := org["longitude"].(float64)
		assert.InDelta(t, 45.5202, lat, 0.5, "Latitude should be near Portland")
		assert.InDelta(t, -122.6742, lng, 0.5, "Longitude should be near Portland")

		// Assert: Organization member record created
		var memberCount int64
		ctx.DB.Table("organization_members").
			Where("organization_id = ? AND role = ?", orgID, "admin").
			Count(&memberCount)
		assert.Equal(t, int64(1), memberCount, "Admin should be added as organization member")

		// Assert: Cause category junction created
		var causeCount int64
		ctx.DB.Table("organization_causes").
			Where("organization_id = ?", orgID).
			Count(&causeCount)
		assert.Equal(t, int64(1), causeCount, "Organization should have Environment cause")

		// Performance: Organization creation should complete within 5 minutes (NFR-017)
		// This is tested implicitly by the test timeout
	})

	var oppID string

	t.Run("Scenario 1.3: Create Volunteer Opportunity", func(t *testing.T) {
		// Create opportunity for future date
		startDate := time.Now().Add(10 * 24 * time.Hour) // 10 days from now

		result := ctx.CreateOpportunity(
			t,
			orgID,
			"Beach Cleanup - Ocean Park",
			startDate,
			20, // capacity
		)

		// Extract opportunity ID
		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "Response should have data field")
		oppID = data["id"].(string)

		// Assert: Opportunity is published immediately
		assert.Equal(t, "published", data["status"], "Opportunity should be published immediately")
		assert.NotNil(t, data["published_at"], "Published timestamp should be set")
		assert.Equal(t, 0, int(data["current_registrations"].(float64)), "Should have 0 registrations initially")
		assert.Equal(t, 20, int(data["capacity"].(float64)), "Capacity should be 20")

		// Assert: Opportunity exists in database with geocoding
		opp := ctx.GetOpportunityByTitle(t, "Beach Cleanup - Ocean Park")
		assert.Equal(t, "Beach Cleanup - Ocean Park", opp["title"])
		assert.Equal(t, "published", opp["status"])
		assert.Equal(t, 20, opp["capacity"])
		assert.Equal(t, 0, opp["current_registrations"])
		assert.NotNil(t, opp["latitude"], "Latitude should be set from geocoding")
		assert.NotNil(t, opp["longitude"], "Longitude should be set from geocoding")

		// Assert: Coordinates are approximately San Francisco
		lat := opp["latitude"].(float64)
		lng := opp["longitude"].(float64)
		assert.InDelta(t, 37.7749, lat, 0.5, "Latitude should be near San Francisco")
		assert.InDelta(t, -122.4194, lng, 0.5, "Longitude should be near San Francisco")

		// Assert: Cause category junction created
		var causeCount int64
		ctx.DB.Table("opportunity_causes").
			Where("opportunity_id = ?", oppID).
			Count(&causeCount)
		assert.Equal(t, int64(1), causeCount, "Opportunity should have Environment cause")

		// Assert: Created by user is set
		assert.NotEmpty(t, opp["created_by_user_id"], "Created by user should be set")

		// Performance: Opportunity creation + geocoding should complete within 5 seconds
		duration := ctx.MeasureResponseTime(t, "POST", "/api/v1/opportunities", map[string]interface{}{
			"organization_id":  orgID,
			"title":            "Performance Test Opportunity",
			"description":      "Test",
			"start_date":       startDate.Format(time.RFC3339),
			"end_date":         startDate.Add(3 * time.Hour).Format(time.RFC3339),
			"timezone":         "America/Los_Angeles",
			"address_line1":    "123 Test St",
			"city":             "San Francisco",
			"state":            "California",
			"postal_code":      "94121",
			"capacity":         10,
			"min_age":          16,
			"cause_categories": []string{"Environment"},
			"publish":          true,
		})
		assert.Less(t, duration.Seconds(), 5.0, "Opportunity creation should complete within 5 seconds")
	})

	// Final assertions: End-to-end validation
	t.Run("End-to-End Validation", func(t *testing.T) {
		// Verify complete data flow
		user := ctx.GetUserByEmail(t, "admin@greenearth.org")
		org := ctx.GetOrganizationBySlug(t, orgSlug)
		opp := ctx.GetOpportunityByTitle(t, "Beach Cleanup - Ocean Park")

		assert.NotEmpty(t, user["id"])
		assert.NotEmpty(t, org["id"])
		assert.NotEmpty(t, opp["id"])

		// Verify relationships
		assert.Equal(t, org["id"], opp["organization_id"], "Opportunity should belong to organization")

		// Verify opportunity appears in search results
		resp := ctx.MakeRequest(t, "GET", "/api/v1/opportunities?status=published", nil)
		helpers.AssertResponseStatus(t, resp, 200)

		var searchResult map[string]interface{}
		helpers.ParseJSONResponse(t, resp, &searchResult)

		data, ok := searchResult["data"].([]interface{})
		require.True(t, ok, "Search should return data array")
		assert.GreaterOrEqual(t, len(data), 1, "Should have at least 1 published opportunity")

		// Find our opportunity in results
		found := false
		for _, item := range data {
			oppData := item.(map[string]interface{})
			if oppData["title"] == "Beach Cleanup - Ocean Park" {
				found = true
				break
			}
		}
		assert.True(t, found, "Published opportunity should appear in search results immediately")
	})
}
