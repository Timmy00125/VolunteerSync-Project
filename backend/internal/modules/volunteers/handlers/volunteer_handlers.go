package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// VolunteerHandler exposes HTTP handlers for volunteer profile management
type VolunteerHandler struct {
	service services.VolunteerService
	log     *logger.Logger
}

// NewVolunteerHandler constructs a VolunteerHandler with required dependencies
func NewVolunteerHandler(service services.VolunteerService, log *logger.Logger) (*VolunteerHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("volunteer handler requires volunteer service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &VolunteerHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires volunteer routes under the provided router group
// All routes require authentication
func (h *VolunteerHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// All volunteer routes require authentication
	// PATCH /volunteers/me - Update volunteer profile
	rg.PATCH("/me", h.UpdateVolunteerProfile)

	// GET /volunteers/me/dashboard - Get dashboard metrics
	rg.GET("/me/dashboard", h.GetDashboard)

	// GET /volunteers/me/analytics - Get analytics data
	rg.GET("/me/analytics", h.GetAnalytics)

	// GET /volunteers/me/report - Download impact report (PDF)
	rg.GET("/me/report", h.GenerateImpactReport)
}

// Request/Response DTOs

type updateVolunteerProfileRequest struct {
	ProfilePhotoURL       *string                  `json:"profile_photo_url"`
	Biography             *string                  `json:"biography"`
	Location              *string                  `json:"location"`
	Availability          *availabilityDTO         `json:"availability"`
	PreferredTime         *models.PreferredTime    `json:"preferred_time"`
	EmergencyContactName  *string                  `json:"emergency_contact_name"`
	EmergencyContactPhone *string                  `json:"emergency_contact_phone"`
	PrivacySettings       *privacySettingsDTO      `json:"privacy_settings"`
	NotificationSettings  *notificationSettingsDTO `json:"notification_settings"`
	SkillIDs              []string                 `json:"skill_ids"`    // UUID strings
	InterestIDs           []string                 `json:"interest_ids"` // UUID strings (cause categories)
}

type availabilityDTO struct {
	Monday    *bool `json:"monday"`
	Tuesday   *bool `json:"tuesday"`
	Wednesday *bool `json:"wednesday"`
	Thursday  *bool `json:"thursday"`
	Friday    *bool `json:"friday"`
	Saturday  *bool `json:"saturday"`
	Sunday    *bool `json:"sunday"`
}

type privacySettingsDTO struct {
	ShowHours         *bool `json:"show_hours"`
	ShowEvents        *bool `json:"show_events"`
	ShowOrganizations *bool `json:"show_organizations"`
}

type notificationSettingsDTO struct {
	InApp       *bool `json:"in_app"`
	BrowserPush *bool `json:"browser_push"`
}

// UpdateVolunteerProfile handles PATCH /volunteers/me
// Updates the authenticated user's volunteer profile
func (h *VolunteerHandler) UpdateVolunteerProfile(c *gin.Context) {
	var req updateVolunteerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Get authenticated user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, apperrors.NewUnauthorizedError("authentication required"))
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, apperrors.NewUnauthorizedError("invalid user ID"))
		return
	}

	ctx := c.Request.Context()

	// Parse skill IDs from strings to UUIDs
	var skillIDs []uuid.UUID
	if req.SkillIDs != nil {
		for _, idStr := range req.SkillIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				h.respondWithError(c, apperrors.NewValidationError("invalid skill ID format", map[string]interface{}{
					"skill_id": idStr,
				}))
				return
			}
			skillIDs = append(skillIDs, id)
		}
	}

	// Parse interest IDs from strings to UUIDs
	var interestIDs []uuid.UUID
	if req.InterestIDs != nil {
		for _, idStr := range req.InterestIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				h.respondWithError(c, apperrors.NewValidationError("invalid interest ID format", map[string]interface{}{
					"interest_id": idStr,
				}))
				return
			}
			interestIDs = append(interestIDs, id)
		}
	}

	// Build service input
	input := services.UpdateVolunteerProfileInput{
		ProfilePhotoURL:       req.ProfilePhotoURL,
		Biography:             req.Biography,
		Location:              req.Location,
		PreferredTime:         req.PreferredTime,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		SkillIDs:              skillIDs,
		InterestIDs:           interestIDs,
	}

	// Map availability if provided
	if req.Availability != nil {
		input.AvailabilityMonday = req.Availability.Monday
		input.AvailabilityTuesday = req.Availability.Tuesday
		input.AvailabilityWednesday = req.Availability.Wednesday
		input.AvailabilityThursday = req.Availability.Thursday
		input.AvailabilityFriday = req.Availability.Friday
		input.AvailabilitySaturday = req.Availability.Saturday
		input.AvailabilitySunday = req.Availability.Sunday
	}

	// Map privacy settings if provided
	if req.PrivacySettings != nil {
		input.PrivacyShowHours = req.PrivacySettings.ShowHours
		input.PrivacyShowEvents = req.PrivacySettings.ShowEvents
		input.PrivacyShowOrganizations = req.PrivacySettings.ShowOrganizations
	}

	// Map notification settings if provided
	if req.NotificationSettings != nil {
		input.NotificationInApp = req.NotificationSettings.InApp
		input.NotificationBrowserPush = req.NotificationSettings.BrowserPush
	}

	// Call service
	profile, err := h.service.UpdateVolunteerProfile(ctx, userUUID, input)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})
}

// GetDashboard handles GET /volunteers/me/dashboard
// Retrieves dashboard metrics for the authenticated volunteer
func (h *VolunteerHandler) GetDashboard(c *gin.Context) {
	// Get authenticated user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, apperrors.NewUnauthorizedError("authentication required"))
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, apperrors.NewUnauthorizedError("invalid user ID"))
		return
	}

	ctx := c.Request.Context()

	// Call service
	dashboard, err := h.service.GetDashboard(ctx, userUUID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dashboard,
	})
}

// GetAnalytics handles GET /volunteers/me/analytics
// Retrieves analytics data for the authenticated volunteer
func (h *VolunteerHandler) GetAnalytics(c *gin.Context) {
	// Get authenticated user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, apperrors.NewUnauthorizedError("authentication required"))
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, apperrors.NewUnauthorizedError("invalid user ID"))
		return
	}

	ctx := c.Request.Context()

	// Parse query parameters for date range (optional)
	// startDate := c.Query("start_date")
	// endDate := c.Query("end_date")
	// For now, use default date range (all time)
	dateRange := services.DateRange{}

	// Call service
	analytics, err := h.service.GetAnalytics(ctx, userUUID, dateRange)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": analytics,
	})
}

// GenerateImpactReport handles GET /volunteers/me/report
// Generates and downloads a PDF impact report for the authenticated volunteer
func (h *VolunteerHandler) GenerateImpactReport(c *gin.Context) {
	// Get authenticated user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, apperrors.NewUnauthorizedError("authentication required"))
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, apperrors.NewUnauthorizedError("invalid user ID"))
		return
	}

	ctx := c.Request.Context()

	// Call service
	pdfData, err := h.service.GenerateImpactReport(ctx, userUUID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Set headers for PDF download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=impact-report-%s.pdf", userUUID.String()))
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// respondWithError is a helper method to send standardized error responses
func (h *VolunteerHandler) respondWithError(c *gin.Context, err error) {
	apperrors.HandleError(c, err)
}
