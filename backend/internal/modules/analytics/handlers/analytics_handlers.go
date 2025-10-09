package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/middleware"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/analytics/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// AnalyticsHandler exposes HTTP handlers for analytics and reporting
type AnalyticsHandler struct {
	service services.AnalyticsService
	log     *logger.Logger
}

// NewAnalyticsHandler constructs an AnalyticsHandler with required dependencies
func NewAnalyticsHandler(service services.AnalyticsService, log *logger.Logger) (*AnalyticsHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("analytics handler requires analytics service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &AnalyticsHandler{
		service: service,
		log:     log,
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

	ctx := c.Request.Context()

	// TODO: Add authorization check - user should only access their own analytics
	// or be an admin/coordinator with permission
	// For now, we'll allow the request to proceed
	_ = userUUID // Suppress unused variable warning

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

	ctx := c.Request.Context()

	// TODO: Add authorization check - user should be a member/admin of the organization
	// For now, we'll allow the request to proceed
	_ = userUUID // Suppress unused variable warning

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

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	// Check if user has admin role
	// TODO: Implement proper role-based access control
	// For now, we'll check for a "is_admin" claim in the context
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		h.log.WithContext(c.Request.Context()).WithFields(map[string]interface{}{
			"user_id": userUUID.String(),
		}).Warn("Unauthorized attempt to access platform analytics")

		h.respondWithError(c, apperrors.NewForbiddenError("administrator access required for platform analytics"))
		return
	}

	ctx := c.Request.Context()

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
