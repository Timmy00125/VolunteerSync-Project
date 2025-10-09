package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/middleware"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/analytics/services"
	orgrepo "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/repositories"
	volrepo "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/repositories"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// AnalyticsHandler exposes HTTP handlers for analytics and reporting
type AnalyticsHandler struct {
	service       services.AnalyticsService
	orgRepo       orgrepo.OrganizationRepository
	volunteerRepo volrepo.VolunteerRepository
	log           *logger.Logger
}

// NewAnalyticsHandler constructs an AnalyticsHandler with required dependencies
func NewAnalyticsHandler(
	service services.AnalyticsService,
	orgRepo orgrepo.OrganizationRepository,
	volunteerRepo volrepo.VolunteerRepository,
	log *logger.Logger,
) (*AnalyticsHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("analytics handler requires analytics service")
	}

	if orgRepo == nil {
		return nil, fmt.Errorf("analytics handler requires organization repository")
	}

	if volunteerRepo == nil {
		return nil, fmt.Errorf("analytics handler requires volunteer repository")
	}

	if log == nil {
		log = logger.Get()
	}

	return &AnalyticsHandler{
		service:       service,
		orgRepo:       orgRepo,
		volunteerRepo: volunteerRepo,
		log:           log,
	}, nil
}

// RegisterRoutes wires analytics routes under the provided router group
// All routes require authentication; platform analytics requires admin role
func (h *AnalyticsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// GET /analytics/volunteer/{id} - Get volunteer analytics (requires auth)
	rg.GET("/volunteer/:id", h.GetVolunteerAnalytics)

	// GET /analytics/organization/{id} - Get organization analytics (requires auth + org member)
	rg.GET("/organization/:id", h.GetOrganizationAnalytics)

	// GET /analytics/platform - Get platform-wide analytics (requires admin role)
	rg.GET("/platform", h.GetPlatformAnalytics)
}

// Request/Response DTOs

type dateRangeQueryParams struct {
	StartDate string `form:"start_date" binding:"required"` // YYYY-MM-DD format
	EndDate   string `form:"end_date" binding:"required"`   // YYYY-MM-DD format
}

// parseAndValidateDateRange parses and validates date range query parameters
func (h *AnalyticsHandler) parseAndValidateDateRange(c *gin.Context) (services.DateRange, error) {
	var params dateRangeQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		return services.DateRange{}, apperrors.NewValidationError("invalid date range parameters", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", params.StartDate)
	if err != nil {
		return services.DateRange{}, apperrors.NewValidationError("invalid start_date format, expected YYYY-MM-DD", map[string]interface{}{
			"start_date": params.StartDate,
		})
	}

	// Parse end date
	endDate, err := time.Parse("2006-01-02", params.EndDate)
	if err != nil {
		return services.DateRange{}, apperrors.NewValidationError("invalid end_date format, expected YYYY-MM-DD", map[string]interface{}{
			"end_date": params.EndDate,
		})
	}

	// Set times to cover the full day range
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	return services.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// GetVolunteerAnalytics handles GET /analytics/volunteer/{id}
// Retrieves analytics data for a volunteer profile (FR-078)
// Query params: start_date (YYYY-MM-DD), end_date (YYYY-MM-DD)
func (h *AnalyticsHandler) GetVolunteerAnalytics(c *gin.Context) {
	// Parse volunteer profile ID from URL
	volunteerIDStr := c.Param("id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid volunteer ID format", map[string]interface{}{
			"volunteer_id": volunteerIDStr,
		}))
		return
	}

	// Parse and validate date range
	dateRange, err := h.parseAndValidateDateRange(c)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)
	userRole := middleware.GetUserRole(c)

	ctx := c.Request.Context()

	// Authorization check: User should only access their own analytics or be an admin/coordinator
	// Fetch the volunteer profile to verify ownership
	volunteerProfile, err := h.volunteerRepo.FindVolunteerProfileByID(ctx, volunteerID)
	if err != nil {
		h.log.WithContext(ctx).ErrorWithErr("Failed to fetch volunteer profile for authorization", err)
		h.respondWithError(c, apperrors.NewNotFoundError("volunteer profile not found"))
		return
	}

	// Check if user owns this volunteer profile or is an admin/coordinator
	isOwner := volunteerProfile.UserID == userUUID
	isStaff := userRole == middleware.RoleSuperAdmin ||
		userRole == middleware.RoleOrgAdmin ||
		userRole == middleware.RoleCoordinator

	if !isOwner && !isStaff {
		h.log.WithContext(ctx).WithFields(map[string]interface{}{
			"user_id":       userUUID.String(),
			"volunteer_id":  volunteerID.String(),
			"profile_owner": volunteerProfile.UserID.String(),
		}).Warn("Unauthorized attempt to access volunteer analytics")

		h.respondWithError(c, apperrors.NewForbiddenError("you can only access your own analytics"))
		return
	}

	// Get volunteer analytics from service
	analytics, err := h.service.GetVolunteerAnalytics(ctx, volunteerID, dateRange)
	if err != nil {
		h.log.WithContext(ctx).ErrorWithErr("Failed to get volunteer analytics", err)

		// Map service errors to HTTP errors
		if err == services.ErrVolunteerNotFound {
			h.respondWithError(c, apperrors.NewNotFoundError("volunteer profile not found"))
			return
		}
		if err == services.ErrInvalidDateRange {
			h.respondWithError(c, apperrors.NewValidationError(err.Error(), nil))
			return
		}

		h.respondWithError(c, apperrors.NewInternalServerError("failed to retrieve volunteer analytics"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetOrganizationAnalytics handles GET /analytics/organization/{id}
// Retrieves analytics data for an organization (FR-077)
// Query params: start_date (YYYY-MM-DD), end_date (YYYY-MM-DD)
func (h *AnalyticsHandler) GetOrganizationAnalytics(c *gin.Context) {
	// Parse organization ID from URL
	orgIDStr := c.Param("id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid organization ID format", map[string]interface{}{
			"organization_id": orgIDStr,
		}))
		return
	}

	// Parse and validate date range
	dateRange, err := h.parseAndValidateDateRange(c)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)
	userRole := middleware.GetUserRole(c)

	ctx := c.Request.Context()

	// Authorization check: User should be a member/admin of the organization or be a super admin
	// Super admins can access all organization analytics
	if userRole != middleware.RoleSuperAdmin {
		// Check if user is a member of this organization
		isMember, err := h.orgRepo.IsMember(ctx, orgID, userUUID)
		if err != nil {
			h.log.WithContext(ctx).ErrorWithErr("Failed to check organization membership", err)
			h.respondWithError(c, apperrors.NewInternalServerError("failed to verify organization membership"))
			return
		}

		if !isMember {
			h.log.WithContext(ctx).WithFields(map[string]interface{}{
				"user_id": userUUID.String(),
				"org_id":  orgID.String(),
			}).Warn("Unauthorized attempt to access organization analytics")

			h.respondWithError(c, apperrors.NewForbiddenError("you must be a member of the organization to access its analytics"))
			return
		}
	}

	// Get organization analytics from service
	analytics, err := h.service.GetOrganizationAnalytics(ctx, orgID, dateRange)
	if err != nil {
		h.log.WithContext(ctx).ErrorWithErr("Failed to get organization analytics", err)

		// Map service errors to HTTP errors
		if err == services.ErrOrganizationNotFound {
			h.respondWithError(c, apperrors.NewNotFoundError("organization not found"))
			return
		}
		if err == services.ErrInvalidDateRange {
			h.respondWithError(c, apperrors.NewValidationError(err.Error(), nil))
			return
		}

		h.respondWithError(c, apperrors.NewInternalServerError("failed to retrieve organization analytics"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetPlatformAnalytics handles GET /analytics/platform
// Retrieves platform-wide analytics (admin only) (FR-080)
// Query params: start_date (YYYY-MM-DD), end_date (YYYY-MM-DD)
func (h *AnalyticsHandler) GetPlatformAnalytics(c *gin.Context) {
	// Parse and validate date range
	dateRange, err := h.parseAndValidateDateRange(c)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Get authenticated user UUID and role from context
	userUUID := middleware.MustGetUserUUID(c)
	userRole := middleware.GetUserRole(c)

	ctx := c.Request.Context()

	// Authorization check: Only super admins can access platform-wide analytics
	if userRole != middleware.RoleSuperAdmin {
		h.log.WithContext(ctx).WithFields(map[string]interface{}{
			"user_id":   userUUID.String(),
			"user_role": userRole,
		}).Warn("Unauthorized attempt to access platform analytics")

		h.respondWithError(c, apperrors.NewForbiddenError("super administrator access required for platform analytics"))
		return
	}

	// Get platform analytics from service
	analytics, err := h.service.GetPlatformAnalytics(ctx, dateRange)
	if err != nil {
		h.log.WithContext(ctx).ErrorWithErr("Failed to get platform analytics", err)

		// Map service errors to HTTP errors
		if err == services.ErrInvalidDateRange {
			h.respondWithError(c, apperrors.NewValidationError(err.Error(), nil))
			return
		}

		h.respondWithError(c, apperrors.NewInternalServerError("failed to retrieve platform analytics"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// respondWithError sends an error response using the application's error structure
func (h *AnalyticsHandler) respondWithError(c *gin.Context, err error) {
	// Check if it's an application error
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.HTTPStatus, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"details": appErr.Details,
			},
		})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	})
}
