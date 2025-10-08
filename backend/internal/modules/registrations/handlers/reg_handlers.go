package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// RegistrationHandler exposes HTTP handlers for registration management
type RegistrationHandler struct {
	service services.RegistrationService
	log     *logger.Logger
}

// NewRegistrationHandler constructs a RegistrationHandler with required dependencies
func NewRegistrationHandler(service services.RegistrationService, log *logger.Logger) (*RegistrationHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("registration handler requires registration service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &RegistrationHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires registration routes under the provided router group
// All routes require authentication
func (h *RegistrationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// All registration routes require authentication
	// POST /registrations - Register for an opportunity
	rg.POST("", h.CreateRegistration)

	// GET /registrations/:id - Get a specific registration
	rg.GET("/:id", h.GetRegistration)

	// PATCH /registrations/:id/cancel - Cancel a registration
	rg.PATCH("/:id/cancel", h.CancelRegistration)

	// PATCH /registrations/:id/check-in - Check in for an event
	rg.PATCH("/:id/check-in", h.CheckInRegistration)

	// GET /registrations/:id/calendar.ics - Download calendar file
	rg.GET("/:id/calendar.ics", h.DownloadCalendar)
}

// Request/Response DTOs

type createRegistrationRequest struct {
	OpportunityID string `json:"opportunity_id" binding:"required"`
}

type cancelRegistrationRequest struct {
	Reason *string `json:"reason"`
}

type registrationResponse struct {
	ID                 string     `json:"id"`
	OpportunityID      string     `json:"opportunity_id"`
	VolunteerProfileID string     `json:"volunteer_profile_id"`
	Status             string     `json:"status"`
	RegisteredAt       time.Time  `json:"registered_at"`
	CheckedInAt        *time.Time `json:"checked_in_at,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason *string    `json:"cancellation_reason,omitempty"`
	HoursWorked        *float64   `json:"hours_worked,omitempty"`
	HoursStatus        *string    `json:"hours_status,omitempty"`
	HoursLoggedAt      *time.Time `json:"hours_logged_at,omitempty"`
	HoursVerifiedAt    *time.Time `json:"hours_verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateRegistration handles POST /registrations
// Registers an authenticated volunteer for an opportunity
func (h *RegistrationHandler) CreateRegistration(c *gin.Context) {
	var req createRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Get authenticated user's volunteer profile ID from context
	// This should be set by auth middleware
	volunteerProfileID, exists := c.Get("volunteer_profile_id")
	if !exists {
		h.respondWithError(c, apperrors.NewUnauthorizedError("volunteer profile required"))
		return
	}

	volunteerProfileUUID, ok := volunteerProfileID.(uuid.UUID)
	if !ok {
		h.respondWithError(c, apperrors.NewUnauthorizedError("invalid volunteer profile ID"))
		return
	}

	// Parse opportunity ID
	opportunityUUID, err := uuid.Parse(req.OpportunityID)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid opportunity ID"))
		return
	}

	ctx := c.Request.Context()

	// Register volunteer
	input := services.RegisterVolunteerInput{
		OpportunityID:      opportunityUUID,
		VolunteerProfileID: volunteerProfileUUID,
	}

	registration, err := h.service.RegisterVolunteer(ctx, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Convert to response DTO
	response := h.toRegistrationResponse(registration)

	c.JSON(http.StatusCreated, gin.H{
		"data": response,
	})
}

// GetRegistration handles GET /registrations/:id
// Retrieves a specific registration by ID
func (h *RegistrationHandler) GetRegistration(c *gin.Context) {
	// Parse registration ID from path
	registrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration ID"))
		return
	}

	ctx := c.Request.Context()

	// Get registration
	registration, err := h.service.GetRegistration(ctx, registrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Verify ownership (volunteer can only see their own registrations)
	volunteerProfileID, exists := c.Get("volunteer_profile_id")
	if exists {
		volunteerProfileUUID := volunteerProfileID.(uuid.UUID)
		if registration.VolunteerProfileID != volunteerProfileUUID {
			h.respondWithError(c, apperrors.NewForbiddenError("access denied"))
			return
		}
	}

	// Convert to response DTO
	response := h.toRegistrationResponse(registration)

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// CancelRegistration handles PATCH /registrations/:id/cancel
// Cancels a volunteer's registration
func (h *RegistrationHandler) CancelRegistration(c *gin.Context) {
	var req cancelRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Parse registration ID from path
	registrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration ID"))
		return
	}

	// Get authenticated volunteer profile ID
	volunteerProfileID, exists := c.Get("volunteer_profile_id")
	if !exists {
		h.respondWithError(c, apperrors.NewUnauthorizedError("volunteer profile required"))
		return
	}

	volunteerProfileUUID := volunteerProfileID.(uuid.UUID)

	ctx := c.Request.Context()

	// Cancel registration
	err = h.service.CancelRegistration(ctx, registrationID, volunteerProfileUUID, req.Reason)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration cancelled successfully",
	})
}

// CheckInRegistration handles PATCH /registrations/:id/check-in
// Checks in a volunteer for an event (coordinator action)
func (h *RegistrationHandler) CheckInRegistration(c *gin.Context) {
	// Parse registration ID from path
	registrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration ID"))
		return
	}

	// Get authenticated user ID (coordinator)
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, apperrors.NewUnauthorizedError("authentication required"))
		return
	}

	coordinatorUUID := userID.(uuid.UUID)

	ctx := c.Request.Context()

	// Check in volunteer
	err = h.service.CheckInVolunteer(ctx, registrationID, coordinatorUUID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Volunteer checked in successfully",
	})
}

// DownloadCalendar handles GET /registrations/:id/calendar.ics
// Generates and downloads an iCalendar file for a registration
func (h *RegistrationHandler) DownloadCalendar(c *gin.Context) {
	// Parse registration ID from path
	registrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration ID"))
		return
	}

	ctx := c.Request.Context()

	// Generate calendar file
	icsContent, err := h.service.GenerateCalendarFile(ctx, registrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Set headers for file download
	c.Header("Content-Type", "text/calendar")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=event_%s.ics", registrationID.String()))

	c.Data(http.StatusOK, "text/calendar", icsContent)
}

// Helper methods

// toRegistrationResponse converts a registration model to response DTO
func (h *RegistrationHandler) toRegistrationResponse(registration *models.Registration) *registrationResponse {
	response := &registrationResponse{
		ID:                 registration.ID.String(),
		OpportunityID:      registration.OpportunityID.String(),
		VolunteerProfileID: registration.VolunteerProfileID.String(),
		Status:             string(registration.Status),
		RegisteredAt:       registration.RegisteredAt,
		CheckedInAt:        registration.CheckedInAt,
		CancelledAt:        registration.CancelledAt,
		CancellationReason: registration.CancellationReason,
		HoursWorked:        registration.HoursWorked,
		HoursLoggedAt:      registration.HoursLoggedAt,
		HoursVerifiedAt:    registration.HoursVerifiedAt,
		CreatedAt:          registration.CreatedAt,
		UpdatedAt:          registration.UpdatedAt,
	}

	if registration.HoursStatus != nil {
		status := string(*registration.HoursStatus)
		response.HoursStatus = &status
	}

	return response
}

// handleServiceError maps service errors to HTTP responses
func (h *RegistrationHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case services.ErrRegistrationNotFound:
		h.respondWithError(c, apperrors.NewNotFoundError("registration not found"))
	case services.ErrRegistrationAlreadyExists:
		h.respondWithError(c, apperrors.NewConflictError("already registered for this opportunity"))
	case services.ErrOpportunityAtCapacity:
		h.respondWithError(c, apperrors.NewConflictError("opportunity is at capacity"))
	case services.ErrRegistrationOverlap:
		h.respondWithError(c, apperrors.NewConflictError("registration conflicts with another event"))
	case services.ErrLateCancellation:
		// This is informational, not an error
		h.respondWithError(c, apperrors.NewBadRequestError("late cancellation warning"))
	case services.ErrUnauthorized:
		h.respondWithError(c, apperrors.NewForbiddenError("access denied"))
	case services.ErrInvalidStatus:
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration status"))
	case services.ErrInvalidRegistrationData:
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration data"))
	default:
		h.respondWithError(c, apperrors.NewInternalServerError("an unexpected error occurred").WithError(err))
	}
}

// respondWithError sends an error response
func (h *RegistrationHandler) respondWithError(c *gin.Context, err *apperrors.AppError) {
	c.JSON(err.HTTPStatus, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}
