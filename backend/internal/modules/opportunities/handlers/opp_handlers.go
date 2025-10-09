package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/middleware"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// OpportunityHandler exposes HTTP handlers for opportunity management
type OpportunityHandler struct {
	service services.OpportunityService
	log     *logger.Logger
}

// NewOpportunityHandler constructs an OpportunityHandler with required dependencies
func NewOpportunityHandler(service services.OpportunityService, log *logger.Logger) (*OpportunityHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("opportunity handler requires opportunity service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &OpportunityHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires opportunity routes under the provided router group
func (h *OpportunityHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// Public routes
	rg.GET("", h.ListOpportunities)  // GET /opportunities (with search/filters)
	rg.GET("/:id", h.GetOpportunity) // GET /opportunities/:id

	// Protected routes (require authentication)
	// authenticated := rg.Group("")
	// authenticated.Use(authMiddleware)
	// authenticated.POST("", h.CreateOpportunity)            // POST /opportunities
	// authenticated.PATCH("/:id", h.UpdateOpportunity)       // PATCH /opportunities/:id
	// authenticated.DELETE("/:id", h.CancelOpportunity)      // DELETE /opportunities/:id (cancel)
	// authenticated.POST("/:id/complete", h.CompleteOpportunity) // POST /opportunities/:id/complete
}

// Request/Response DTOs

type createOpportunityRequest struct {
	OrganizationID     string     `json:"organization_id" binding:"required"`
	Title              string     `json:"title" binding:"required"`
	Description        string     `json:"description" binding:"required"`
	PublishImmediately bool       `json:"publish_immediately"`
	StartDate          string     `json:"start_date" binding:"required"` // YYYY-MM-DD
	StartTime          string     `json:"start_time" binding:"required"` // HH:MM
	EndDate            string     `json:"end_date" binding:"required"`
	EndTime            string     `json:"end_time" binding:"required"`
	Timezone           string     `json:"timezone" binding:"required"`
	Address            addressDTO `json:"address" binding:"required"`
	Capacity           int        `json:"capacity" binding:"required,min=1"`
	MinAge             *int       `json:"min_age"`
	IsRecurring        bool       `json:"is_recurring"`
	RecurrencePattern  *string    `json:"recurrence_pattern"` // daily, weekly, monthly
	RecurrenceEndDate  *string    `json:"recurrence_end_date"`
	SkillIDs           []string   `json:"skill_ids"`
	CauseIDs           []string   `json:"cause_ids"`
	DocumentIDs        []string   `json:"document_ids"`
}

type updateOpportunityRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	Status      *string     `json:"status"` // draft, published, cancelled, completed
	StartDate   *string     `json:"start_date"`
	StartTime   *string     `json:"start_time"`
	EndDate     *string     `json:"end_date"`
	EndTime     *string     `json:"end_time"`
	Timezone    *string     `json:"timezone"`
	Address     *addressDTO `json:"address"`
	Capacity    *int        `json:"capacity"`
	MinAge      *int        `json:"min_age"`
	SkillIDs    []string    `json:"skill_ids"`
	CauseIDs    []string    `json:"cause_ids"`
	DocumentIDs []string    `json:"document_ids"`
}

type addressDTO struct {
	AddressLine1 string  `json:"address_line_1" binding:"required"`
	AddressLine2 *string `json:"address_line_2"`
	City         string  `json:"city" binding:"required"`
	State        string  `json:"state" binding:"required"`
	PostalCode   string  `json:"postal_code" binding:"required"`
	Country      string  `json:"country"`
}

type cancelOpportunityRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// CreateOpportunity handles POST /opportunities
func (h *OpportunityHandler) CreateOpportunity(c *gin.Context) {
	var req createOpportunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Parse organization ID
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid organization ID"))
		return
	}

	// Parse dates and times
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid start date format, use YYYY-MM-DD", nil))
		return
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid start time format, use HH:MM", nil))
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid end date format, use YYYY-MM-DD", nil))
		return
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid end time format, use HH:MM", nil))
		return
	}

	// Parse optional fields
	var recurrencePattern *models.RecurrencePattern
	if req.RecurrencePattern != nil {
		pattern := models.RecurrencePattern(*req.RecurrencePattern)
		recurrencePattern = &pattern
	}

	var recurrenceEndDate *time.Time
	if req.RecurrenceEndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.RecurrenceEndDate)
		if err != nil {
			h.respondWithError(c, apperrors.NewValidationError("invalid recurrence end date format", nil))
			return
		}
		recurrenceEndDate = &parsed
	}

	// Parse UUID arrays
	skillIDs, err := h.parseUUIDs(req.SkillIDs)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid skill IDs"))
		return
	}

	causeIDs, err := h.parseUUIDs(req.CauseIDs)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid cause IDs"))
		return
	}

	documentIDs, err := h.parseUUIDs(req.DocumentIDs)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid document IDs"))
		return
	}

	// Set default country if not provided
	country := "United States"
	if req.Address.Country != "" {
		country = req.Address.Country
	}

	// Build service input
	input := services.CreateOpportunityInput{
		OrganizationID:     orgID,
		Title:              req.Title,
		Description:        req.Description,
		PublishImmediately: req.PublishImmediately,
		StartDate:          startDate,
		StartTime:          startTime,
		EndDate:            endDate,
		EndTime:            endTime,
		Timezone:           req.Timezone,
		AddressLine1:       req.Address.AddressLine1,
		AddressLine2:       req.Address.AddressLine2,
		City:               req.Address.City,
		State:              req.Address.State,
		PostalCode:         req.Address.PostalCode,
		Country:            country,
		Capacity:           req.Capacity,
		MinAge:             req.MinAge,
		IsRecurring:        req.IsRecurring,
		RecurrencePattern:  recurrencePattern,
		RecurrenceEndDate:  recurrenceEndDate,
		SkillIDs:           skillIDs,
		CauseIDs:           causeIDs,
		DocumentIDs:        documentIDs,
	}

	// Call service
	opportunity, err := h.service.CreateOpportunity(ctx, input, userUUID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": opportunity,
	})
}

// GetOpportunity handles GET /opportunities/:id
func (h *OpportunityHandler) GetOpportunity(c *gin.Context) {
	idStr := c.Param("id")
	oppID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid opportunity ID"))
		return
	}

	ctx := c.Request.Context()

	opportunity, err := h.service.GetOpportunity(ctx, oppID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": opportunity,
	})
}

// UpdateOpportunity handles PATCH /opportunities/:id
func (h *OpportunityHandler) UpdateOpportunity(c *gin.Context) {
	idStr := c.Param("id")
	oppID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid opportunity ID"))
		return
	}

	var req updateOpportunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Parse optional date/time fields
	var startDate, endDate *time.Time
	var startTime, endTime *time.Time

	if req.StartDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			h.respondWithError(c, apperrors.NewValidationError("invalid start date format", nil))
			return
		}
		startDate = &parsed
	}

	if req.StartTime != nil {
		parsed, err := time.Parse("15:04", *req.StartTime)
		if err != nil {
			h.respondWithError(c, apperrors.NewValidationError("invalid start time format", nil))
			return
		}
		startTime = &parsed
	}

	if req.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			h.respondWithError(c, apperrors.NewValidationError("invalid end date format", nil))
			return
		}
		endDate = &parsed
	}

	if req.EndTime != nil {
		parsed, err := time.Parse("15:04", *req.EndTime)
		if err != nil {
			h.respondWithError(c, apperrors.NewValidationError("invalid end time format", nil))
			return
		}
		endTime = &parsed
	}

	// Parse status if provided
	var status *models.OpportunityStatus
	if req.Status != nil {
		s := models.OpportunityStatus(*req.Status)
		status = &s
	}

	// Parse UUID arrays
	skillIDs, err := h.parseUUIDs(req.SkillIDs)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid skill IDs"))
		return
	}

	causeIDs, err := h.parseUUIDs(req.CauseIDs)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid cause IDs"))
		return
	}

	documentIDs, err := h.parseUUIDs(req.DocumentIDs)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid document IDs"))
		return
	}

	// Build service input
	input := services.UpdateOpportunityInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		StartDate:   startDate,
		StartTime:   startTime,
		EndDate:     endDate,
		EndTime:     endTime,
		Timezone:    req.Timezone,
		Capacity:    req.Capacity,
		MinAge:      req.MinAge,
		SkillIDs:    skillIDs,
		CauseIDs:    causeIDs,
		DocumentIDs: documentIDs,
	}

	// Handle address if provided
	if req.Address != nil {
		input.AddressLine1 = &req.Address.AddressLine1
		input.AddressLine2 = req.Address.AddressLine2
		input.City = &req.Address.City
		input.State = &req.Address.State
		input.PostalCode = &req.Address.PostalCode
		input.Country = &req.Address.Country
	}

	// Call service
	opportunity, err := h.service.UpdateOpportunity(ctx, oppID, input, userUUID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": opportunity,
	})
}

// ListOpportunities handles GET /opportunities with search and filters
func (h *OpportunityHandler) ListOpportunities(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	filters := services.OpportunityListFilters{
		Search:        c.Query("search"),
		City:          c.Query("city"),
		State:         c.Query("state"),
		OnlyRecurring: c.Query("recurring") == "true",
		Page:          h.parseIntQuery(c, "page", 1),
		Limit:         h.parseIntQuery(c, "limit", 20),
		SortBy:        c.DefaultQuery("sort_by", "created_at"),
		SortOrder:     c.DefaultQuery("sort_order", "desc"),
	}

	// Parse optional UUID filters
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid organization ID"))
			return
		}
		filters.OrganizationID = &orgID
	}

	// Parse status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.OpportunityStatus(statusStr)
		filters.Status = &status
	}

	// Parse location-based search
	if latStr := c.Query("latitude"); latStr != "" {
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid latitude"))
			return
		}
		filters.Latitude = &lat
	}

	if lngStr := c.Query("longitude"); lngStr != "" {
		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid longitude"))
			return
		}
		filters.Longitude = &lng
	}

	if radiusStr := c.Query("radius_km"); radiusStr != "" {
		radius, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid radius"))
			return
		}
		filters.RadiusKm = &radius
	}

	// Parse date range filters
	if startFromStr := c.Query("start_date_from"); startFromStr != "" {
		startFrom, err := time.Parse("2006-01-02", startFromStr)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid start_date_from format"))
			return
		}
		filters.StartDateFrom = &startFrom
	}

	if startToStr := c.Query("start_date_to"); startToStr != "" {
		startTo, err := time.Parse("2006-01-02", startToStr)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid start_date_to format"))
			return
		}
		filters.StartDateTo = &startTo
	}

	// Parse min age filter
	if minAgeStr := c.Query("min_age"); minAgeStr != "" {
		minAge, err := strconv.Atoi(minAgeStr)
		if err != nil {
			h.respondWithError(c, apperrors.NewBadRequestError("invalid min_age"))
			return
		}
		filters.MinAge = &minAge
	}

	// Call service
	result, err := h.service.ListOpportunities(ctx, filters)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelOpportunity handles DELETE /opportunities/:id or POST /opportunities/:id/cancel
func (h *OpportunityHandler) CancelOpportunity(c *gin.Context) {
	idStr := c.Param("id")
	oppID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid opportunity ID"))
		return
	}

	var req cancelOpportunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("cancellation reason is required").WithError(err))
		return
	}

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Call service
	if err := h.service.CancelOpportunity(ctx, oppID, req.Reason, userUUID); err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Opportunity cancelled successfully",
	})
}

// CompleteOpportunity handles POST /opportunities/:id/complete
func (h *OpportunityHandler) CompleteOpportunity(c *gin.Context) {
	idStr := c.Param("id")
	oppID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid opportunity ID"))
		return
	}

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Call service
	if err := h.service.CompleteOpportunity(ctx, oppID, userUUID); err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Opportunity completed successfully",
	})
}

// Helper methods

func (h *OpportunityHandler) parseUUIDs(ids []string) ([]uuid.UUID, error) {
	if ids == nil {
		return nil, nil
	}

	result := make([]uuid.UUID, 0, len(ids))
	for _, idStr := range ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

func (h *OpportunityHandler) parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func (h *OpportunityHandler) respondWithError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "internal_server_error",
		"message": "An unexpected error occurred",
	})
}
