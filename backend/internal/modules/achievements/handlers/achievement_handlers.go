package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// AchievementHandler exposes HTTP handlers for achievement flows
type AchievementHandler struct {
	service services.AchievementService
	log     *logger.Logger
}

// NewAchievementHandler constructs an AchievementHandler with required dependencies
func NewAchievementHandler(service services.AchievementService, log *logger.Logger) (*AchievementHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("achievement handler requires achievement service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &AchievementHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires achievement routes under the provided router group
func (h *AchievementHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// Public routes
	rg.GET("/achievements", h.ListAchievements)
	rg.GET("/achievements/:id", h.GetAchievement)

	// Volunteer routes
	rg.GET("/volunteers/:id/achievements", h.GetVolunteerAchievements)

	// Organization routes (for custom achievements)
	rg.POST("/organizations/:org_id/achievements", h.CreateCustomAchievement)
	rg.POST("/achievements/:id/award", h.AwardCustomAchievement)
}

// listAchievementsQueryParams captures query parameters for listing achievements
type listAchievementsQueryParams struct {
	OrganizationID *uuid.UUID `form:"organization_id"`
}

// ListAchievements handles GET /achievements
// Lists all available achievements, optionally filtered by organization
// Query params:
//   - organization_id (optional): UUID of organization to include custom badges
func (h *AchievementHandler) ListAchievements(c *gin.Context) {
	var query listAchievementsQueryParams

	// Parse organization_id from query param if provided
	orgIDStr := c.Query("organization_id")
	if orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid organization_id"))
			return
		}
		query.OrganizationID = &orgID
	}

	ctx := c.Request.Context()

	achievements, err := h.service.GetAllAchievements(ctx, query.OrganizationID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"achievements": achievements,
		"count":        len(achievements),
	})
}

// GetAchievement handles GET /achievements/:id
// Retrieves details of a specific achievement
func (h *AchievementHandler) GetAchievement(c *gin.Context) {
	achievementIDStr := c.Param("id")
	achievementID, err := uuid.Parse(achievementIDStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid achievement ID"))
		return
	}

	ctx := c.Request.Context()

	achievement, err := h.service.GetAchievementByID(ctx, achievementID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"achievement": achievement,
	})
}

// GetVolunteerAchievements handles GET /volunteers/:id/achievements
// Retrieves all achievements earned by a specific volunteer
func (h *AchievementHandler) GetVolunteerAchievements(c *gin.Context) {
	volunteerIDStr := c.Param("id")
	volunteerID, err := uuid.Parse(volunteerIDStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid volunteer ID"))
		return
	}

	ctx := c.Request.Context()

	// Get all achievements earned by this volunteer
	volunteerAchievements, err := h.service.GetVolunteerAchievements(ctx, volunteerID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Get total count
	count, err := h.service.CountVolunteerAchievements(ctx, volunteerID)
	if err != nil {
		h.log.WithFields(map[string]interface{}{
			"volunteer_id": volunteerID,
		}).Warn("Failed to count volunteer achievements")
		count = int64(len(volunteerAchievements))
	}

	c.JSON(http.StatusOK, gin.H{
		"volunteer_profile_id": volunteerID,
		"achievements":         volunteerAchievements,
		"total_count":          count,
	})
}

// createCustomAchievementRequest captures the request body for creating a custom achievement
type createCustomAchievementRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	IconURL     *string `json:"icon_url"`
}

// CreateCustomAchievement handles POST /organizations/:org_id/achievements
// Creates a new organization-specific custom achievement
// This is used by organization coordinators to create custom badges (FR-075)
func (h *AchievementHandler) CreateCustomAchievement(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid organization ID"))
		return
	}

	var req createCustomAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	// TODO: Check if user is authorized (org admin/coordinator) using middleware

	achievement, err := h.service.CreateCustomAchievement(ctx, services.CreateCustomAchievementInput{
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		IconURL:        req.IconURL,
	})
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	h.log.WithFields(map[string]interface{}{
		"achievement_id":  achievement.ID,
		"organization_id": orgID,
		"name":            req.Name,
	}).Info("Custom achievement created")

	c.JSON(http.StatusCreated, gin.H{
		"achievement": achievement,
		"message":     "Custom achievement created successfully",
	})
}

// awardCustomAchievementRequest captures the request body for awarding a custom achievement
type awardCustomAchievementRequest struct {
	VolunteerProfileID uuid.UUID `json:"volunteer_profile_id" binding:"required"`
}

// AwardCustomAchievement handles POST /achievements/:id/award
// Manually awards a custom achievement to a volunteer
// This is used by coordinators to award organization-specific badges (FR-075)
func (h *AchievementHandler) AwardCustomAchievement(c *gin.Context) {
	achievementIDStr := c.Param("id")
	achievementID, err := uuid.Parse(achievementIDStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid achievement ID"))
		return
	}

	var req awardCustomAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	ctx := c.Request.Context()

	// TODO: Get current user from context (authenticated user)
	// For now, we'll use a placeholder - this should be set by auth middleware
	awardedByUserID := uuid.New() // Placeholder - should come from c.GetString("user_id") or similar

	volunteerAchievement, err := h.service.AwardCustomAchievement(ctx, services.AwardCustomAchievementInput{
		VolunteerProfileID: req.VolunteerProfileID,
		AchievementID:      achievementID,
		AwardedByUserID:    awardedByUserID,
	})
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	h.log.WithFields(map[string]interface{}{
		"achievement_id":       achievementID,
		"volunteer_profile_id": req.VolunteerProfileID,
		"awarded_by_user_id":   awardedByUserID,
	}).Info("Custom achievement awarded")

	c.JSON(http.StatusCreated, gin.H{
		"volunteer_achievement": volunteerAchievement,
		"message":               "Achievement awarded successfully",
	})
}

// respondWithError is a helper to send error responses in a consistent format
func (h *AchievementHandler) respondWithError(c *gin.Context, err error) {
	if c == nil || err == nil {
		return
	}

	// Handle application errors with proper status codes
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error":   appErr.Code,
			"message": appErr.Message,
			"details": appErr.Details,
		})
		return
	}

	// Log unexpected errors for debugging
	h.log.WithFields(map[string]interface{}{
		"error": err.Error(),
	}).Error("Unexpected error in achievement handler")

	// Return detailed error response for unexpected errors
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal_server_error",
		"message": err.Error(),
	})
}
