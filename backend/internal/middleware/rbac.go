package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// Role constants for RBAC
const (
	RoleSuperAdmin  = "super_admin"
	RoleOrgAdmin    = "org_admin"
	RoleCoordinator = "coordinator"
	RoleVolunteer   = "volunteer"
)

// OrganizationMembershipChecker defines the interface for checking organization membership
// This allows the RBAC middleware to verify user membership without tight coupling to the repository
type OrganizationMembershipChecker interface {
	// IsMember checks if a user is a member of an organization (any role)
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
}

// RequireRole creates a middleware that checks if the user has one of the required roles
// This middleware should be used after AuthMiddleware
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Get user role from context
		userRole := GetUserRole(c)
		if userRole == "" {
			log.Warn("User role not found in context - authentication middleware not applied?")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Authentication required"))
			return
		}

		// Check if user has one of the allowed roles
		hasRole := false
		for _, role := range allowedRoles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			log.WithField("user_role", userRole).
				WithField("allowed_roles", allowedRoles).
				Warn("User does not have required role")
			errors.AbortWithError(c, errors.NewForbiddenError("Insufficient permissions"))
			return
		}

		log.WithField("user_role", userRole).
			WithField("allowed_roles", allowedRoles).
			Debug("Role check passed")

		c.Next()
	}
}

// RequireSuperAdmin is a convenience middleware that requires super admin role
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(RoleSuperAdmin)
}

// RequireOrgAdmin is a convenience middleware that requires org admin or super admin role
func RequireOrgAdmin() gin.HandlerFunc {
	return RequireRole(RoleOrgAdmin, RoleSuperAdmin)
}

// RequireCoordinator is a convenience middleware that requires coordinator, org admin, or super admin role
func RequireCoordinator() gin.HandlerFunc {
	return RequireRole(RoleCoordinator, RoleOrgAdmin, RoleSuperAdmin)
}

// RequireStaff is a convenience middleware that requires any staff role (coordinator, org admin, or super admin)
func RequireStaff() gin.HandlerFunc {
	return RequireRole(RoleCoordinator, RoleOrgAdmin, RoleSuperAdmin)
}

// RequireAnyRole is a convenience middleware that requires any authenticated user
// (essentially checks if user has any role assigned)
func RequireAnyRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Get user role from context
		userRole := GetUserRole(c)
		if userRole == "" {
			log.Warn("User role not found in context - authentication middleware not applied?")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Authentication required"))
			return
		}

		log.WithField("user_role", userRole).Debug("User has valid role")
		c.Next()
	}
}

// RequireOrgMembership creates a middleware that verifies the user belongs to the specified organization
// The organization ID is typically extracted from the URL parameter (e.g., /organizations/:org_id/...)
// This middleware should be used after AuthMiddleware
//
// Parameters:
//   - orgIDParam: The name of the URL parameter containing the organization ID
//   - checker: An implementation of OrganizationMembershipChecker for database lookups
//
// If checker is nil, the middleware will only store the org_id in context without verification
// (useful for backward compatibility or when verification is done at handler level)
func RequireOrgMembership(orgIDParam string, checker OrganizationMembershipChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Get user ID from context
		userID := GetUserID(c)
		if userID == "" {
			log.Warn("User ID not found in context - authentication middleware not applied?")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Authentication required"))
			return
		}

		// Get user role from context
		userRole := GetUserRole(c)

		// Super admins have access to all organizations
		if userRole == RoleSuperAdmin {
			log.WithField("user_id", userID).
				WithField("role", userRole).
				Debug("Super admin accessing organization")
			c.Next()
			return
		}

		// Get organization ID from URL parameter
		orgID := c.Param(orgIDParam)
		if orgID == "" {
			log.Warn("Organization ID not found in URL parameter")
			errors.AbortWithError(c, errors.NewBadRequestError("Organization ID required"))
			return
		}

		// Store org_id in context for handler use
		c.Set("org_id", orgID)

		// If no checker is provided, defer verification to handler level
		if checker == nil {
			log.WithField("user_id", userID).
				WithField("org_id", orgID).
				Debug("Organization membership check deferred to handler")
			c.Next()
			return
		}

		// Parse organization ID as UUID
		orgUUID, err := uuid.Parse(orgID)
		if err != nil {
			log.WithField("org_id", orgID).
				Warn("Invalid organization ID format")
			errors.AbortWithError(c, errors.NewBadRequestError("Invalid organization ID"))
			return
		}

		// Parse user ID as UUID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			log.WithField("user_id", userID).
				Error("Invalid user ID format in context")
			errors.AbortWithError(c, errors.NewInternalServerError("Internal server error"))
			return
		}

		// Check if user is a member of the organization
		isMember, err := checker.IsMember(c.Request.Context(), orgUUID, userUUID)
		if err != nil {
			log.WithField("user_id", userID).
				WithField("org_id", orgID).
				WithField("error", err.Error()).
				Error("Failed to check organization membership")
			errors.AbortWithError(c, errors.NewInternalServerError("Failed to verify organization membership"))
			return
		}

		if !isMember {
			log.WithField("user_id", userID).
				WithField("org_id", orgID).
				Warn("User is not a member of the organization")
			errors.AbortWithError(c, errors.NewForbiddenError("You are not a member of this organization"))
			return
		}

		log.WithField("user_id", userID).
			WithField("org_id", orgID).
			Debug("Organization membership verified")

		c.Next()
	}
}

// RequireResourceOwnership creates a middleware that verifies the user owns the specified resource
// The resource owner ID is extracted from the URL parameter (e.g., /users/:user_id/profile)
// This allows users to access their own resources, while staff can access any resource
func RequireResourceOwnership(userIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get().WithContext(c.Request.Context())

		// Get authenticated user ID from context
		authUserID := GetUserID(c)
		if authUserID == "" {
			log.Warn("User ID not found in context - authentication middleware not applied?")
			errors.AbortWithError(c, errors.NewUnauthorizedError("Authentication required"))
			return
		}

		// Get user role from context
		userRole := GetUserRole(c)

		// Super admins and org admins have access to all resources
		if userRole == RoleSuperAdmin || userRole == RoleOrgAdmin {
			log.WithField("auth_user_id", authUserID).
				WithField("role", userRole).
				Debug("Staff accessing resource")
			c.Next()
			return
		}

		// Get resource owner ID from URL parameter
		resourceUserID := c.Param(userIDParam)
		if resourceUserID == "" {
			log.Warn("Resource user ID not found in URL parameter")
			errors.AbortWithError(c, errors.NewBadRequestError("User ID required"))
			return
		}

		// Check if authenticated user owns the resource
		if authUserID != resourceUserID {
			log.WithField("auth_user_id", authUserID).
				WithField("resource_user_id", resourceUserID).
				Warn("User does not own resource")
			errors.AbortWithError(c, errors.NewForbiddenError("You can only access your own resources"))
			return
		}

		log.WithField("auth_user_id", authUserID).
			WithField("resource_user_id", resourceUserID).
			Debug("Resource ownership verified")

		c.Next()
	}
}

// GetOrgID extracts the organization ID from the Gin context
// Returns empty string if not set by RequireOrgMembership middleware
func GetOrgID(c *gin.Context) string {
	orgID, exists := c.Get("org_id")
	if !exists {
		return ""
	}
	return orgID.(string)
}

// IsSuperAdmin checks if the current user is a super admin
func IsSuperAdmin(c *gin.Context) bool {
	return GetUserRole(c) == RoleSuperAdmin
}

// IsOrgAdmin checks if the current user is an org admin
func IsOrgAdmin(c *gin.Context) bool {
	return GetUserRole(c) == RoleOrgAdmin
}

// IsCoordinator checks if the current user is a coordinator
func IsCoordinator(c *gin.Context) bool {
	return GetUserRole(c) == RoleCoordinator
}

// IsVolunteer checks if the current user is a volunteer
func IsVolunteer(c *gin.Context) bool {
	return GetUserRole(c) == RoleVolunteer
}

// IsStaff checks if the current user is staff (coordinator, org admin, or super admin)
func IsStaff(c *gin.Context) bool {
	role := GetUserRole(c)
	return role == RoleCoordinator || role == RoleOrgAdmin || role == RoleSuperAdmin
}
