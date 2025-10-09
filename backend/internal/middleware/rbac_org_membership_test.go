package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/middleware"
)

// mockMembershipChecker is a mock implementation for testing
type mockMembershipChecker struct {
	shouldBeMember bool
	err            error
}

func (m *mockMembershipChecker) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return m.shouldBeMember, m.err
}

func TestRequireOrgMembership_WithChecker_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()
	
	// Create a mock checker that returns true (user is a member)
	checker := &mockMembershipChecker{shouldBeMember: true}
	
	// Set up test route with middleware
	testUserID := uuid.New()
	testOrgID := uuid.New()
	
	router.GET("/organizations/:org_id/test", func(c *gin.Context) {
		// Simulate auth middleware setting user context
		c.Set("user_id", testUserID.String())
		c.Set("user_role", middleware.RoleVolunteer)
		c.Next()
	}, middleware.RequireOrgMembership("org_id", checker), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/organizations/"+testOrgID.String()+"/test", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireOrgMembership_WithChecker_NotMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	
	// Create a mock checker that returns false (user is not a member)
	checker := &mockMembershipChecker{shouldBeMember: false}
	
	testUserID := uuid.New()
	testOrgID := uuid.New()
	
	router.GET("/organizations/:org_id/test", func(c *gin.Context) {
		c.Set("user_id", testUserID.String())
		c.Set("user_role", middleware.RoleVolunteer)
		c.Next()
	}, middleware.RequireOrgMembership("org_id", checker), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest(http.MethodGet, "/organizations/"+testOrgID.String()+"/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireOrgMembership_SuperAdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	
	// Create a mock checker that would return false, but super admin should bypass
	checker := &mockMembershipChecker{shouldBeMember: false}
	
	testUserID := uuid.New()
	testOrgID := uuid.New()
	
	router.GET("/organizations/:org_id/test", func(c *gin.Context) {
		c.Set("user_id", testUserID.String())
		c.Set("user_role", middleware.RoleSuperAdmin) // Super admin role
		c.Next()
	}, middleware.RequireOrgMembership("org_id", checker), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest(http.MethodGet, "/organizations/"+testOrgID.String()+"/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Super admin should bypass the check
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireOrgMembership_NilChecker_DeferToHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	
	testUserID := uuid.New()
	testOrgID := uuid.New()
	
	router.GET("/organizations/:org_id/test", func(c *gin.Context) {
		c.Set("user_id", testUserID.String())
		c.Set("user_role", middleware.RoleVolunteer)
		c.Next()
	}, middleware.RequireOrgMembership("org_id", nil), func(c *gin.Context) {
		// Handler can access org_id from context
		orgID := middleware.GetOrgID(c)
		assert.Equal(t, testOrgID.String(), orgID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest(http.MethodGet, "/organizations/"+testOrgID.String()+"/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should pass through to handler when checker is nil
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireOrgMembership_InvalidOrgID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	checker := &mockMembershipChecker{shouldBeMember: true}
	
	testUserID := uuid.New()
	
	router.GET("/organizations/:org_id/test", func(c *gin.Context) {
		c.Set("user_id", testUserID.String())
		c.Set("user_role", middleware.RoleVolunteer)
		c.Next()
	}, middleware.RequireOrgMembership("org_id", checker), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Use invalid UUID format
	req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequireOrgMembership_MissingAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	checker := &mockMembershipChecker{shouldBeMember: true}
	
	testOrgID := uuid.New()
	
	// Don't set user context - simulating missing auth middleware
	router.GET("/organizations/:org_id/test", 
		middleware.RequireOrgMembership("org_id", checker), 
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	
	req := httptest.NewRequest(http.MethodGet, "/organizations/"+testOrgID.String()+"/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
