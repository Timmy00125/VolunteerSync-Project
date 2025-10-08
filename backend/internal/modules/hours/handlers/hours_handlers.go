package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// HoursHandler exposes HTTP handlers for volunteer hours tracking management
type HoursHandler struct {
	service services.HoursService
	log     *logger.Logger
}

// NewHoursHandler constructs a HoursHandler with required dependencies
func NewHoursHandler(service services.HoursService, log *logger.Logger) (*HoursHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("hours handler requires hours service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &HoursHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires hours tracking routes under the provided router group
// All routes require authentication
func (h *HoursHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// POST /hours/log - Log volunteer hours (coordinator only)
	rg.POST("/log", h.LogHours)

	// POST /hours/:id/verify - Verify logged hours (volunteer only)
	rg.POST("/:id/verify", h.VerifyHours)

	// POST /hours/:id/dispute - Dispute logged hours (volunteer only)
	rg.POST("/:id/dispute", h.DisputeHours)

	// POST /hours/:id/resolve - Resolve disputed hours (coordinator only)
	rg.POST("/:id/resolve", h.ResolveDispute)

	// GET /hours/:id - Get a specific hours log
	rg.GET("/:id", h.GetHoursLog)

	// GET /hours/volunteer/:volunteer_id - Get all hours logs for a volunteer (admin/coordinator)
	rg.GET("/volunteer/:volunteer_id", h.GetHoursLogsByVolunteer)

	// GET /hours/registration/:registration_id - Get all hours logs for a registration
	rg.GET("/registration/:registration_id", h.GetHoursLogsByRegistration)

	// GET /hours/volunteer/:volunteer_id/pending - Get pending hours for volunteer
	rg.GET("/volunteer/:volunteer_id/pending", h.GetPendingHours)
}

// Request/Response DTOs

type logHoursRequest struct {
	RegistrationID   string  `json:"registration_id" binding:"required"`
	Hours            float64 `json:"hours" binding:"required,gt=0"`
	CoordinatorNotes *string `json:"coordinator_notes"`
}

type verifyHoursRequest struct {
	VolunteerNotes *string `json:"volunteer_notes"`
}

type disputeHoursRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type resolveDisputeRequest struct {
	ResolutionNotes string `json:"resolution_notes" binding:"required"`
}

type hoursLogResponse struct {
	ID               string     `json:"id"`
	RegistrationID   string     `json:"registration_id"`
	Hours            float64    `json:"hours"`
	LoggedByUserID   string     `json:"logged_by_user_id"`
	Status           string     `json:"status"`
	CoordinatorNotes *string    `json:"coordinator_notes,omitempty"`
	VolunteerNotes   *string    `json:"volunteer_notes,omitempty"`
	DisputeReason    *string    `json:"dispute_reason,omitempty"`
	DisputedAt       *time.Time `json:"disputed_at,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes  *string    `json:"resolution_notes,omitempty"`
	LoggedAt         time.Time  `json:"logged_at"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	AutoVerifiedAt   *time.Time `json:"auto_verified_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// LogHours handles POST /hours/log
// Logs volunteer hours for a registration (coordinator only)
func (h *HoursHandler) LogHours(c *gin.Context) {
	var req logHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Get authenticated user ID from context (should be set by auth middleware)
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

	// Verify user is a coordinator (should be checked by RBAC middleware)
	role, exists := c.Get("role")
	if !exists || (role != "coordinator" && role != "org_admin" && role != "super_admin") {
		h.respondWithError(c, apperrors.NewForbiddenError("coordinator role required"))
		return
	}

	// Parse registration ID
	registrationUUID, err := uuid.Parse(req.RegistrationID)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration ID"))
		return
	}

	ctx := c.Request.Context()

	// Log hours
	input := services.LogHoursInput{
		RegistrationID:   registrationUUID,
		Hours:            req.Hours,
		LoggedByUserID:   userUUID,
		CoordinatorNotes: req.CoordinatorNotes,
	}

	hoursLog, err := h.service.LogHours(ctx, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Convert to response DTO
	response := h.toHoursLogResponse(hoursLog)

	c.JSON(http.StatusCreated, gin.H{
		"data":    response,
		"message": "Hours logged successfully. Volunteer has been notified.",
	})
}

// VerifyHours handles POST /hours/:id/verify
// Verifies logged hours (volunteer only)
func (h *HoursHandler) VerifyHours(c *gin.Context) {
	var req verifyHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Parse hours log ID from path
	hoursLogID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid hours log ID"))
		return
	}

	// Get authenticated volunteer profile ID from context
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

	ctx := c.Request.Context()

	// Verify hours
	if err := h.service.VerifyHours(ctx, hoursLogID, volunteerProfileUUID, req.VolunteerNotes); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hours verified successfully. Your total hours have been updated.",
	})
}

// DisputeHours handles POST /hours/:id/dispute
// Disputes logged hours (volunteer only)
func (h *HoursHandler) DisputeHours(c *gin.Context) {
	var req disputeHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Parse hours log ID from path
	hoursLogID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid hours log ID"))
		return
	}

	// Get authenticated volunteer profile ID from context
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

	ctx := c.Request.Context()

	// Dispute hours
	if err := h.service.DisputeHours(ctx, hoursLogID, volunteerProfileUUID, req.Reason); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hours disputed successfully. The coordinator has been notified.",
	})
}

// ResolveDispute handles POST /hours/:id/resolve
// Resolves disputed hours (coordinator only)
func (h *HoursHandler) ResolveDispute(c *gin.Context) {
	var req resolveDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Parse hours log ID from path
	hoursLogID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid hours log ID"))
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

	// Verify user is a coordinator (should be checked by RBAC middleware)
	role, exists := c.Get("role")
	if !exists || (role != "coordinator" && role != "org_admin" && role != "super_admin") {
		h.respondWithError(c, apperrors.NewForbiddenError("coordinator role required"))
		return
	}

	ctx := c.Request.Context()

	// Resolve dispute
	if err := h.service.ResolveDispute(ctx, hoursLogID, userUUID, req.ResolutionNotes); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dispute resolved successfully. Volunteer has been notified.",
	})
}

// GetHoursLog handles GET /hours/:id
// Retrieves a specific hours log by ID
func (h *HoursHandler) GetHoursLog(c *gin.Context) {
	// Parse hours log ID from path
	hoursLogID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid hours log ID"))
		return
	}

	ctx := c.Request.Context()

	// Get hours log
	hoursLog, err := h.service.GetHoursLog(ctx, hoursLogID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Convert to response DTO
	response := h.toHoursLogResponse(hoursLog)

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetHoursLogsByVolunteer handles GET /hours/volunteer/:volunteer_id
// Retrieves all hours logs for a volunteer
func (h *HoursHandler) GetHoursLogsByVolunteer(c *gin.Context) {
	// Parse volunteer profile ID from path
	volunteerProfileID, err := uuid.Parse(c.Param("volunteer_id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid volunteer profile ID"))
		return
	}

	ctx := c.Request.Context()

	// Get hours logs
	hoursLogs, err := h.service.GetHoursLogsByVolunteer(ctx, volunteerProfileID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Convert to response DTOs
	responses := make([]hoursLogResponse, len(hoursLogs))
	for i, log := range hoursLogs {
		responses[i] = h.toHoursLogResponse(log)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"count": len(responses),
	})
}

// GetHoursLogsByRegistration handles GET /hours/registration/:registration_id
// Retrieves all hours logs for a registration
func (h *HoursHandler) GetHoursLogsByRegistration(c *gin.Context) {
	// Parse registration ID from path
	registrationID, err := uuid.Parse(c.Param("registration_id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid registration ID"))
		return
	}

	ctx := c.Request.Context()

	// Get hours logs
	hoursLogs, err := h.service.GetHoursLogsByRegistration(ctx, registrationID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Convert to response DTOs
	responses := make([]hoursLogResponse, len(hoursLogs))
	for i, log := range hoursLogs {
		responses[i] = h.toHoursLogResponse(log)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"count": len(responses),
	})
}

// GetPendingHours handles GET /hours/volunteer/:volunteer_id/pending
// Retrieves pending hours logs for a volunteer
func (h *HoursHandler) GetPendingHours(c *gin.Context) {
	// Parse volunteer profile ID from path
	volunteerProfileID, err := uuid.Parse(c.Param("volunteer_id"))
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid volunteer profile ID"))
		return
	}

	// Verify volunteer is accessing their own pending hours
	authVolunteerProfileID, exists := c.Get("volunteer_profile_id")
	if exists {
		authVolunteerUUID := authVolunteerProfileID.(uuid.UUID)
		if authVolunteerUUID != volunteerProfileID {
			// Check if user is coordinator/admin
			role, roleExists := c.Get("role")
			if !roleExists || (role != "coordinator" && role != "org_admin" && role != "super_admin") {
				h.respondWithError(c, apperrors.NewForbiddenError("access denied"))
				return
			}
		}
	}

	ctx := c.Request.Context()

	// Get pending hours logs
	hoursLogs, err := h.service.GetPendingHoursForVolunteer(ctx, volunteerProfileID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Convert to response DTOs
	responses := make([]hoursLogResponse, len(hoursLogs))
	for i, log := range hoursLogs {
		responses[i] = h.toHoursLogResponse(log)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"count": len(responses),
	})
}

// Helper methods

// toHoursLogResponse converts a hours log model to a response DTO
func (h *HoursHandler) toHoursLogResponse(hoursLog *models.HoursLog) hoursLogResponse {
	status := string(hoursLog.Status)

	return hoursLogResponse{
		ID:               hoursLog.ID.String(),
		RegistrationID:   hoursLog.RegistrationID.String(),
		Hours:            hoursLog.Hours,
		LoggedByUserID:   hoursLog.LoggedByUserID.String(),
		Status:           status,
		CoordinatorNotes: hoursLog.CoordinatorNotes,
		VolunteerNotes:   hoursLog.VolunteerNotes,
		DisputeReason:    hoursLog.DisputeReason,
		DisputedAt:       hoursLog.DisputedAt,
		ResolvedAt:       hoursLog.ResolvedAt,
		ResolutionNotes:  hoursLog.ResolutionNotes,
		LoggedAt:         hoursLog.LoggedAt,
		VerifiedAt:       hoursLog.VerifiedAt,
		AutoVerifiedAt:   hoursLog.AutoVerifiedAt,
		CreatedAt:        hoursLog.CreatedAt,
		UpdatedAt:        hoursLog.UpdatedAt,
	}
}

// respondWithError sends a structured error response
func (h *HoursHandler) respondWithError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An internal error occurred",
		},
	})
}

// handleServiceError maps service errors to appropriate HTTP responses
func (h *HoursHandler) handleServiceError(c *gin.Context, err error) {
	switch err {
	case services.ErrHoursLogNotFound:
		h.respondWithError(c, apperrors.NewNotFoundError("hours log not found"))
	case services.ErrInvalidHoursData:
		h.respondWithError(c, apperrors.NewBadRequestError("invalid hours data").WithError(err))
	case services.ErrUnauthorized:
		h.respondWithError(c, apperrors.NewForbiddenError("access denied"))
	case services.ErrInvalidStatusTransition:
		h.respondWithError(c, apperrors.NewBadRequestError("invalid status transition").WithError(err))
	case services.ErrAlreadyVerified:
		h.respondWithError(c, apperrors.NewBadRequestError("hours are already verified"))
	case services.ErrAlreadyDisputed:
		h.respondWithError(c, apperrors.NewBadRequestError("hours are already disputed"))
	default:
		h.log.Errorf("unhandled service error: %v", err)
		h.respondWithError(c, apperrors.NewInternalServerError("failed to process request"))
	}
}
