# RBAC Middleware Usage Guide

This document explains how to use the Role-Based Access Control (RBAC) middleware in the VolunteerSync backend.

## Overview

The RBAC middleware provides several functions to control access to routes based on user roles and organization membership.

## Role Constants

```go
const (
    RoleSuperAdmin  = "super_admin"
    RoleOrgAdmin    = "org_admin"
    RoleCoordinator = "coordinator"
    RoleVolunteer   = "volunteer"
)
```

## Basic Role Checks

### RequireRole

Check if a user has one of the specified roles:

```go
router.GET("/admin/dashboard",
    middleware.AuthMiddleware(jwtManager),
    middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleOrgAdmin),
    handler.AdminDashboard,
)
```

### Convenience Middleware

```go
// Requires super admin only
middleware.RequireSuperAdmin()

// Requires org admin or super admin
middleware.RequireOrgAdmin()

// Requires coordinator, org admin, or super admin
middleware.RequireCoordinator()

// Requires any staff role
middleware.RequireStaff()

// Requires any authenticated user with a role
middleware.RequireAnyRole()
```

## Organization Membership Checks

### RequireOrgMembership

**New in v1.1**: Now supports actual database verification of organization membership!

#### Basic Usage (Deferred Verification)

If you want to defer membership checks to the handler level:

```go
router.GET("/organizations/:org_id/opportunities",
    middleware.AuthMiddleware(jwtManager),
    middleware.RequireOrgMembership("org_id", nil), // nil = defer to handler
    handler.GetOrgOpportunities,
)
```

#### With Database Verification

To verify membership at the middleware level (recommended for security):

```go
// In main.go, when setting up routes:
orgRepo := orgRepos.NewOrganizationRepository(dbConn.DB)

// Apply to route groups that require org membership
orgProtectedRoutes := authenticated.Group("/organizations/:org_id")
orgProtectedRoutes.Use(middleware.RequireOrgMembership("org_id", orgRepo))
{
    orgProtectedRoutes.POST("/opportunities", handler.CreateOpportunity)
    orgProtectedRoutes.PUT("/settings", handler.UpdateOrgSettings)
    orgProtectedRoutes.DELETE("/members/:user_id", handler.RemoveMember)
}
```

#### How It Works

1. **Super Admin Bypass**: Users with `super_admin` role automatically pass the check
2. **UUID Validation**: Validates organization ID and user ID formats
3. **Database Lookup**: Calls `IsMember()` on the provided checker
4. **Error Handling**: Returns appropriate HTTP errors (400, 401, 403, 500)
5. **Context Storage**: Stores `org_id` in context for downstream handlers

#### Interface Implementation

The middleware uses the `OrganizationMembershipChecker` interface:

```go
type OrganizationMembershipChecker interface {
    IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
}
```

Any repository that implements this interface can be used. The `OrganizationRepository` already implements it via the `IsMember` method.

## Resource Ownership Checks

### RequireResourceOwnership

Verifies a user owns a specific resource (or is staff):

```go
router.GET("/users/:user_id/profile",
    middleware.AuthMiddleware(jwtManager),
    middleware.RequireResourceOwnership("user_id"),
    handler.GetUserProfile,
)
```

**How it works**:

- Super admins and org admins can access any resource
- Regular users can only access their own resources
- Compares authenticated user ID with resource owner ID from URL

## Helper Functions

### Context Getters

```go
// Get organization ID from context (set by RequireOrgMembership)
orgID := middleware.GetOrgID(c)

// Check user roles
isSuperAdmin := middleware.IsSuperAdmin(c)
isOrgAdmin := middleware.IsOrgAdmin(c)
isCoordinator := middleware.IsCoordinator(c)
isVolunteer := middleware.IsVolunteer(c)
isStaff := middleware.IsStaff(c)
```

## Best Practices

1. **Always use AuthMiddleware first**: RBAC middleware depends on authentication context
2. **Use specific role checks**: Prefer `RequireStaff()` over `RequireRole(...)` for clarity
3. **Verify org membership at middleware level**: Use database checker for security-critical routes
4. **Defer to handlers for complex logic**: Use `nil` checker when authorization logic is complex
5. **Log all authorization failures**: The middleware logs warnings/errors automatically

## Example: Complete Route Setup

```go
func setupRoutes(router *gin.Engine, jwtManager *jwt.Manager, orgRepo orgRepos.OrganizationRepository) {
    v1 := router.Group("/api/v1")

    // Public routes
    publicGroup := v1.Group("/public")
    publicGroup.GET("/opportunities", handler.ListPublicOpportunities)

    // Authenticated routes
    authenticated := v1.Group("")
    authenticated.Use(middleware.AuthMiddleware(jwtManager))
    authenticated.Use(middleware.ContextEnrichmentMiddleware())
    authenticated.Use(middleware.RequireAnyRole())

    // User-owned resources
    userRoutes := authenticated.Group("/users/:user_id")
    userRoutes.Use(middleware.RequireResourceOwnership("user_id"))
    {
        userRoutes.GET("/profile", handler.GetProfile)
        userRoutes.PUT("/profile", handler.UpdateProfile)
    }

    // Organization member routes (with DB verification)
    orgRoutes := authenticated.Group("/organizations/:org_id")
    orgRoutes.Use(middleware.RequireOrgMembership("org_id", orgRepo))
    {
        orgRoutes.GET("/dashboard", handler.GetOrgDashboard)
        orgRoutes.POST("/opportunities", handler.CreateOpportunity)
    }

    // Organization admin routes
    orgAdminRoutes := authenticated.Group("/organizations/:org_id/admin")
    orgAdminRoutes.Use(middleware.RequireOrgMembership("org_id", orgRepo))
    orgAdminRoutes.Use(middleware.RequireOrgAdmin())
    {
        orgAdminRoutes.PUT("/settings", handler.UpdateOrgSettings)
        orgAdminRoutes.DELETE("", handler.DeleteOrganization)
    }

    // Super admin routes
    adminRoutes := authenticated.Group("/admin")
    adminRoutes.Use(middleware.RequireSuperAdmin())
    {
        adminRoutes.GET("/analytics", handler.GetPlatformAnalytics)
        adminRoutes.POST("/achievements", handler.CreatePlatformAchievement)
    }
}
```

## Migration Guide

If you're using the old `RequireOrgMembership` (without database verification):

**Before:**

```go
router.Use(middleware.RequireOrgMembership("org_id"))
```

**After (with DB verification):**

```go
// Option 1: Add database verification (recommended)
router.Use(middleware.RequireOrgMembership("org_id", orgRepo))

// Option 2: Keep old behavior (defer to handler)
router.Use(middleware.RequireOrgMembership("org_id", nil))
```

## Testing

When writing tests, you can mock the `OrganizationMembershipChecker`:

```go
type mockMembershipChecker struct {
    shouldBeMember bool
    err            error
}

func (m *mockMembershipChecker) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
    return m.shouldBeMember, m.err
}

// In test:
checker := &mockMembershipChecker{shouldBeMember: true}
router.Use(middleware.RequireOrgMembership("org_id", checker))
```

## Security Considerations

1. **Always verify membership for sensitive operations** (create, update, delete)
2. **Super admins bypass all checks** - ensure this role is tightly controlled
3. **Use HTTPS in production** to protect authentication tokens
4. **Log all authorization failures** for security auditing
5. **Consider rate limiting** on authentication/authorization endpoints

## Related Documentation

- [Authentication Middleware](./auth.go)
- [Context Enrichment](./context_enrichment.go)
- [Organization Repository](../modules/organizations/repositories/org_repository.go)
