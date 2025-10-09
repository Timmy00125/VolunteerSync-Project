package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestContext holds the test environment setup
type TestContext struct {
	DB           *gorm.DB
	HTTPClient   *http.Client
	BaseURL      string
	Server       *httptest.Server
	AccessToken  string
	RefreshToken string
}

// SetupTestEnvironment initializes a clean test environment with database and HTTP server
func SetupTestEnvironment(t *testing.T) *TestContext {
	// TODO: Initialize testcontainers for PostgreSQL and Redis
	// For now, we'll use a placeholder connection string
	dsn := "host=localhost user=volunteersync_test password=test dbname=volunteersync_test port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Clean database state
	CleanDatabase(t, db)

	// Seed lookup data
	SeedLookupData(t, db)

	// TODO: Initialize HTTP server with Gin router
	// For now, we'll use a test server placeholder
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}))

	return &TestContext{
		DB:         db,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    server.URL,
		Server:     server,
	}
}

// TeardownTestEnvironment cleans up the test environment
func TeardownTestEnvironment(t *testing.T, ctx *TestContext) {
	if ctx.Server != nil {
		ctx.Server.Close()
	}

	if ctx.DB != nil {
		sqlDB, err := ctx.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// CleanDatabase removes all data from tables
func CleanDatabase(t *testing.T, db *gorm.DB) {
	// Order matters due to foreign key constraints
	tables := []string{
		"volunteer_achievements",
		"hours_logs",
		"notifications",
		"message_recipients",
		"messages",
		"registrations",
		"opportunity_skills",
		"opportunity_causes",
		"opportunities",
		"volunteer_interests",
		"volunteer_skills",
		"volunteer_profiles",
		"organization_causes",
		"organization_members",
		"organizations",
		"users",
		// Don't clean lookup tables: skills, cause_categories, achievements
	}

	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err, "Failed to clean table: %s", table)
	}
}

// SeedLookupData populates cause categories and skills
func SeedLookupData(t *testing.T, db *gorm.DB) {
	// Seed Cause Categories
	causes := []map[string]interface{}{
		{"name": "Environment", "description": "Environmental protection and conservation"},
		{"name": "Education", "description": "Teaching and learning opportunities"},
		{"name": "Healthcare", "description": "Health and wellness services"},
		{"name": "Social Services", "description": "Community support services"},
	}

	for _, cause := range causes {
		// Check if already exists
		var count int64
		db.Table("cause_categories").Where("name = ?", cause["name"]).Count(&count)
		if count == 0 {
			err := db.Table("cause_categories").Create(cause).Error
			require.NoError(t, err, "Failed to seed cause category")
		}
	}

	// Seed Skills
	skills := []map[string]interface{}{
		{"name": "Web Development", "category": "Technical"},
		{"name": "Teaching", "category": "Education"},
		{"name": "Event Planning", "category": "Administrative"},
		{"name": "First Aid", "category": "Healthcare"},
		{"name": "Graphic Design", "category": "Creative"},
	}

	for _, skill := range skills {
		// Check if already exists
		var count int64
		db.Table("skills").Where("name = ?", skill["name"]).Count(&count)
		if count == 0 {
			err := db.Table("skills").Create(skill).Error
			require.NoError(t, err, "Failed to seed skill")
		}
	}

	// Seed Achievements
	achievements := []map[string]interface{}{
		{
			"name":        "First Event",
			"description": "Completed your first volunteer event",
			"badge_icon":  "🎉",
			"criteria":    "Complete 1 event",
		},
		{
			"name":        "10 Hours",
			"description": "Volunteered for 10 hours",
			"badge_icon":  "⏰",
			"criteria":    "Log 10 verified hours",
		},
		{
			"name":        "5 Organizations",
			"description": "Volunteered with 5 different organizations",
			"badge_icon":  "🤝",
			"criteria":    "Work with 5 organizations",
		},
	}

	for _, achievement := range achievements {
		var count int64
		db.Table("achievements").Where("name = ?", achievement["name"]).Count(&count)
		if count == 0 {
			err := db.Table("achievements").Create(achievement).Error
			require.NoError(t, err, "Failed to seed achievement")
		}
	}
}

// MakeRequest sends an HTTP request to the test server
func (ctx *TestContext) MakeRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, ctx.BaseURL+path, reqBody)
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	if ctx.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.AccessToken)
	}

	resp, err := ctx.HTTPClient.Do(req)
	require.NoError(t, err, "Failed to send request")

	return resp
}

// ParseJSONResponse parses the response body into the provided interface
func ParseJSONResponse(t *testing.T, resp *http.Response, v interface{}) {
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	err = json.Unmarshal(bodyBytes, v)
	require.NoError(t, err, "Failed to unmarshal response: %s", string(bodyBytes))
}

// AssertResponseStatus checks if the response has the expected status code
func AssertResponseStatus(t *testing.T, resp *http.Response, expectedStatus int) {
	require.Equal(t, expectedStatus, resp.StatusCode,
		"Expected status %d, got %d", expectedStatus, resp.StatusCode)
}

// RegisterUser creates a new user account via API
func (ctx *TestContext) RegisterUser(t *testing.T, email, password, firstName, lastName, userType string) map[string]interface{} {
	payload := map[string]interface{}{
		"email":      email,
		"password":   password,
		"first_name": firstName,
		"last_name":  lastName,
		"user_type":  userType,
		"security_questions": []map[string]string{
			{"question": "What is your mother's maiden name?", "answer": "Johnson"},
			{"question": "What city were you born in?", "answer": "Boston"},
			{"question": "What was your first pet's name?", "answer": "Max"},
		},
	}

	resp := ctx.MakeRequest(t, "POST", "/api/v1/auth/register", payload)
	AssertResponseStatus(t, resp, http.StatusCreated)

	var result map[string]interface{}
	ParseJSONResponse(t, resp, &result)

	// Store tokens for authenticated requests
	if data, ok := result["data"].(map[string]interface{}); ok {
		if accessToken, ok := data["access_token"].(string); ok {
			ctx.AccessToken = accessToken
		}
		if refreshToken, ok := data["refresh_token"].(string); ok {
			ctx.RefreshToken = refreshToken
		}
	}

	return result
}

// LoginUser authenticates a user and returns tokens
func (ctx *TestContext) LoginUser(t *testing.T, email, password string) map[string]interface{} {
	payload := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	resp := ctx.MakeRequest(t, "POST", "/api/v1/auth/login", payload)
	AssertResponseStatus(t, resp, http.StatusOK)

	var result map[string]interface{}
	ParseJSONResponse(t, resp, &result)

	// Store tokens
	if data, ok := result["data"].(map[string]interface{}); ok {
		if accessToken, ok := data["access_token"].(string); ok {
			ctx.AccessToken = accessToken
		}
		if refreshToken, ok := data["refresh_token"].(string); ok {
			ctx.RefreshToken = refreshToken
		}
	}

	return result
}

// CreateOrganization creates a new organization via API
func (ctx *TestContext) CreateOrganization(t *testing.T, name, mission string) map[string]interface{} {
	payload := map[string]interface{}{
		"name":              name,
		"mission_statement": mission,
		"description":       "Test organization description",
		"email":             fmt.Sprintf("contact@%s.org", name),
		"phone":             "+1-555-0100",
		"address_line1":     "123 Test Street",
		"city":              "Portland",
		"state":             "Oregon",
		"postal_code":       "97201",
		"website":           fmt.Sprintf("https://%s.org", name),
		"cause_categories":  []string{"Environment"},
	}

	resp := ctx.MakeRequest(t, "POST", "/api/v1/organizations", payload)
	AssertResponseStatus(t, resp, http.StatusCreated)

	var result map[string]interface{}
	ParseJSONResponse(t, resp, &result)

	return result
}

// CreateOpportunity creates a new volunteer opportunity via API
func (ctx *TestContext) CreateOpportunity(t *testing.T, orgID, title string, startDate time.Time, capacity int) map[string]interface{} {
	endDate := startDate.Add(3 * time.Hour)

	payload := map[string]interface{}{
		"organization_id":  orgID,
		"title":            title,
		"description":      "Test opportunity description",
		"start_date":       startDate.Format(time.RFC3339),
		"end_date":         endDate.Format(time.RFC3339),
		"timezone":         "America/Los_Angeles",
		"address_line1":    "Ocean Park Beach",
		"city":             "San Francisco",
		"state":            "California",
		"postal_code":      "94121",
		"capacity":         capacity,
		"min_age":          16,
		"cause_categories": []string{"Environment"},
		"publish":          true,
	}

	resp := ctx.MakeRequest(t, "POST", "/api/v1/opportunities", payload)
	AssertResponseStatus(t, resp, http.StatusCreated)

	var result map[string]interface{}
	ParseJSONResponse(t, resp, &result)

	return result
}

// GetUserByEmail retrieves a user from the database by email
func (ctx *TestContext) GetUserByEmail(t *testing.T, email string) map[string]interface{} {
	var user map[string]interface{}
	err := ctx.DB.Table("users").Where("email = ?", email).First(&user).Error
	require.NoError(t, err, "Failed to get user by email")
	return user
}

// GetOrganizationBySlug retrieves an organization from database by slug
func (ctx *TestContext) GetOrganizationBySlug(t *testing.T, slug string) map[string]interface{} {
	var org map[string]interface{}
	err := ctx.DB.Table("organizations").Where("slug = ?", slug).First(&org).Error
	require.NoError(t, err, "Failed to get organization by slug")
	return org
}

// GetOpportunityByTitle retrieves an opportunity from database by title
func (ctx *TestContext) GetOpportunityByTitle(t *testing.T, title string) map[string]interface{} {
	var opp map[string]interface{}
	err := ctx.DB.Table("opportunities").Where("title = ?", title).First(&opp).Error
	require.NoError(t, err, "Failed to get opportunity by title")
	return opp
}

// WaitForCondition polls a condition function until it returns true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "Timeout waiting for condition: "+message)
			return
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// MeasureResponseTime measures the time taken for an HTTP request
func (ctx *TestContext) MeasureResponseTime(t *testing.T, method, path string, body interface{}) time.Duration {
	start := time.Now()
	resp := ctx.MakeRequest(t, method, path, body)
	duration := time.Since(start)
	resp.Body.Close()
	return duration
}
