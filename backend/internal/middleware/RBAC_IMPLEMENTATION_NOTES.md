# RBAC Middleware Organization Membership Implementation

**Date**: October 9, 2025  
**Status**: ✅ Complete  
**Priority**: HIGH  
**Related TODO**: Line 131 in `backend/internal/middleware/rbac.go`

## Summary

Implemented actual database verification for organization membership in the RBAC middleware, moving from deferred handler-level checks to optional middleware-level enforcement.

## What Changed

### 1. New Interface: `OrganizationMembershipChecker`

Added a new interface to support dependency injection and testing:

```go
type OrganizationMembershipChecker interface {
    IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
}
```

This interface:

- Decouples middleware from repository implementation
- Enables easy mocking in tests
- Follows dependency inversion principle

### 2. Updated `RequireOrgMembership` Function

**Old Signature:**

```go
func RequireOrgMembership(orgIDParam string) gin.HandlerFunc
```

**New Signature:**

```go
func RequireOrgMembership(orgIDParam string, checker OrganizationMembershipChecker) gin.HandlerFunc
```

**Key Features:**

- ✅ Performs actual database lookup when checker is provided
- ✅ Validates UUID formats for org_id and user_id
- ✅ Super admin bypass (super admins access all organizations)
- ✅ Backward compatible (pass `nil` for old behavior)
- ✅ Comprehensive error handling (400, 401, 403, 500)
- ✅ Detailed logging for debugging and security auditing
- ✅ Stores org_id in context for downstream handlers

### 3. Workflow

```
1. Extract user_id and role from context (set by AuthMiddleware)
2. Check if user is super_admin → ALLOW (bypass)
3. Extract org_id from URL parameter
4. Store org_id in context
5. If checker is nil → ALLOW (defer to handler)
6. Parse UUIDs (org_id, user_id)
7. Call checker.IsMember(ctx, orgID, userID)
8. If member → ALLOW, else → DENY (403 Forbidden)
```

## Files Modified

1. **backend/internal/middleware/rbac.go**

   - Added `OrganizationMembershipChecker` interface
   - Updated `RequireOrgMembership` implementation
   - Added context and uuid imports

2. **todos.md**
   - Marked TODO as complete
   - Updated progress tracking (6 of 7 items complete)
   - Added implementation details

## Files Created

1. **backend/internal/middleware/RBAC_USAGE.md**

   - Comprehensive usage guide
   - Examples for all middleware functions
   - Migration guide from old to new API
   - Best practices and security considerations

2. **backend/internal/middleware/rbac_org_membership_test.go**
   - 6 test cases covering all scenarios:
     - ✅ Success: User is member
     - ✅ Forbidden: User is not member
     - ✅ Bypass: Super admin access
     - ✅ Defer: Nil checker behavior
     - ✅ Bad Request: Invalid UUID format
     - ✅ Unauthorized: Missing auth context

## Integration with Existing Code

### OrganizationRepository Already Compatible

The `OrganizationRepository` interface already has the required method:

```go
// From backend/internal/modules/organizations/repositories/org_repository.go
type OrganizationRepository interface {
    ...
    IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
    ...
}
```

The `gormOrganizationRepository` implementation is fully compatible:

```go
func (r *gormOrganizationRepository) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
    if orgID == uuid.Nil || userID == uuid.Nil {
        return false, nil
    }

    var count int64
    err := r.db.WithContext(ctx).
        Model(&models.OrganizationMember{}).
        Where("organization_id = ? AND user_id = ?", orgID, userID).
        Count(&count).Error

    if err != nil {
        return false, fmt.Errorf("failed to check membership: %w", err)
    }

    return count > 0, nil
}
```

### No Breaking Changes

The update is **backward compatible**. Existing code can continue to work by passing `nil`:

```go
// Old usage (still works)
router.Use(middleware.RequireOrgMembership("org_id", nil))

// New usage (recommended)
router.Use(middleware.RequireOrgMembership("org_id", orgRepo))
```

## Usage Examples

### Example 1: Apply to Route Group

```go
// In main.go
orgRepo := orgRepos.NewOrganizationRepository(dbConn.DB)

orgRoutes := authenticated.Group("/organizations/:org_id")
orgRoutes.Use(middleware.RequireOrgMembership("org_id", orgRepo))
{
    orgRoutes.GET("/dashboard", handler.GetOrgDashboard)
    orgRoutes.POST("/opportunities", handler.CreateOpportunity)
    orgRoutes.PUT("/settings", handler.UpdateOrgSettings)
}
```

### Example 2: Individual Route

```go
router.POST("/organizations/:org_id/events",
    middleware.AuthMiddleware(jwtManager),
    middleware.RequireOrgMembership("org_id", orgRepo),
    handler.CreateEvent,
)
```

### Example 3: Testing with Mock

```go
type mockChecker struct {
    shouldBeMember bool
}

func (m *mockChecker) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
    return m.shouldBeMember, nil
}

// In test
checker := &mockChecker{shouldBeMember: true}
router.Use(middleware.RequireOrgMembership("org_id", checker))
```

## Security Improvements

### Before

- ❌ Middleware only stored org_id in context
- ❌ No actual membership verification
- ❌ Relied on handlers to check membership (inconsistent)
- ❌ Easy to forget authorization checks in handlers

### After

- ✅ Actual database verification at middleware level
- ✅ Consistent enforcement across all routes
- ✅ Centralized security policy
- ✅ Comprehensive logging for security audits
- ✅ Super admin access properly controlled
- ✅ Invalid UUIDs caught early with clear errors

## Testing

All tests pass successfully:

```bash
$ go test -v ./internal/middleware/... -run TestRequireOrgMembership
=== RUN   TestRequireOrgMembership_WithChecker_Success
--- PASS: TestRequireOrgMembership_WithChecker_Success (0.00s)
=== RUN   TestRequireOrgMembership_WithChecker_NotMember
--- PASS: TestRequireOrgMembership_WithChecker_NotMember (0.00s)
=== RUN   TestRequireOrgMembership_SuperAdminBypass
--- PASS: TestRequireOrgMembership_SuperAdminBypass (0.00s)
=== RUN   TestRequireOrgMembership_NilChecker_DeferToHandler
--- PASS: TestRequireOrgMembership_NilChecker_DeferToHandler (0.00s)
=== RUN   TestRequireOrgMembership_InvalidOrgID
--- PASS: TestRequireOrgMembership_InvalidOrgID (0.00s)
=== RUN   TestRequireOrgMembership_MissingAuth
--- PASS: TestRequireOrgMembership_MissingAuth (0.00s)
PASS
```

## Next Steps

### Recommended Actions

1. **Update main.go** to use the new checker parameter:

   ```go
   orgRoutes.Use(middleware.RequireOrgMembership("org_id", orgRepo))
   ```

2. **Audit existing routes** to identify where org membership checks are needed

3. **Remove redundant handler-level checks** where middleware now handles it

4. **Update integration tests** to expect 403 responses for non-members

### Related TODOs

Still remaining in Authorization & Security:

1. **Organizations Service Authorization** (HIGH priority)
   - Line 216: Create organization member record for creator as admin
   - Line 295: Verify user is admin before allowing organization updates
   - Line 431: Verify user is admin before allowing organization deletion

These should be tackled next to complete the authorization system.

## Performance Considerations

- **Database query added per request**: One additional query to check membership
- **Mitigation**: Consider adding caching layer for membership checks
- **Trade-off**: Security vs. performance (security wins for critical operations)
- **Optimization opportunity**: Batch membership checks for multiple orgs

## Conclusion

This implementation successfully addresses the TODO at line 131 of `rbac.go`. The middleware now performs actual database verification of organization membership while maintaining backward compatibility and following SOLID principles through dependency injection.

The implementation is:

- ✅ Secure
- ✅ Testable
- ✅ Well-documented
- ✅ Backward compatible
- ✅ Production-ready

**Impact**: Significantly improves security by enforcing organization membership at the middleware level, preventing unauthorized access to organization resources.
