package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// OrganizationHandler exposes HTTP handlers for organization management
type OrganizationHandler struct {
	service services.OrganizationService
	log     *logger.Logger
}

// NewOrganizationHandler constructs an OrganizationHandler with required dependencies
func NewOrganizationHandler(service services.OrganizationService, log *logger.Logger) (*OrganizationHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("organization handler requires organization service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &OrganizationHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires organization routes under the provided router group
func (h *OrganizationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// Public routes
	rg.GET("", h.ListOrganizations)   // GET /organizations
	rg.GET("/:id", h.GetOrganization) // GET /organizations/:id

	// Protected routes (require authentication)
	// These would be added to a separate group with auth middleware
	// authenticated := rg.Group("")
	// authenticated.Use(authMiddleware)
	// authenticated.POST("", h.CreateOrganization)           // POST /organizations
	// authenticated.PATCH("/:id", h.UpdateOrganization)      // PATCH /organizations/:id
	// authenticated.DELETE("/:id", h.DeleteOrganization)     // DELETE /organizations/:id
}

// Request/Response DTOs

type createOrganizationRequest struct {
	Name             string      `json:"name"`
	MissionStatement *string     `json:"mission_statement"`
	Description      *string     `json:"description"`
	Website          *string     `json:"website"`
	Email            string      `json:"email"`
	Phone            *string     `json:"phone"`
	Address          *addressDTO `json:"address"`
	LogoURL          *string     `json:"logo_url"`
	BannerURL        *string     `json:"banner_url"`
	CauseIDs         []string    `json:"cause_ids"` // UUID strings
}

type updateOrganizationRequest struct {
	Name             *string     `json:"name"`
	MissionStatement *string     `json:"mission_statement"`
	Description      *string     `json:"description"`
	Website          *string     `json:"website"`
	Email            *string     `json:"email"`
	Phone            *string     `json:"phone"`
	Address          *addressDTO `json:"address"`
	LogoURL          *string     `json:"logo_url"`
	BannerURL        *string     `json:"banner_url"`
	CauseIDs         []string    `json:"cause_ids"` // UUID strings
}

type addressDTO struct {
	AddressLine1 *string  `json:"address_line_1"`
	AddressLine2 *string  `json:"address_line_2"`
	City         *string  `json:"city"`
	State        *string  `json:"state"`
	PostalCode   *string  `json:"postal_code"`
	Country      *string  `json:"country"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

// CreateOrganization handles POST /organizations
// Creates a new organization with the authenticated user as admin
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req createOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Get authenticated user ID from context
	// This would be set by the authentication middleware
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

	// Parse cause IDs from strings to UUIDs
	var causeIDs []uuid.UUID
	if req.CauseIDs != nil {
		for _, idStr := range req.CauseIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				h.respondWithError(c, apperrors.NewValidationError("invalid cause ID format", map[string]interface{}{
					"cause_id": idStr,
				}))
				return
			}
			causeIDs = append(causeIDs, id)
		}
	}

	// Build service input
	input := services.CreateOrganizationInput{
		Name:             req.Name,
		MissionStatement: req.MissionStatement,
		Description:      req.Description,
		Website:          req.Website,
		Email:            req.Email,
		Phone:            req.Phone,
		LogoURL:          req.LogoURL,
		BannerURL:        req.BannerURL,
		CauseIDs:         causeIDs,
	}

	// Handle address if provided
	if req.Address != nil {
		input.AddressLine1 = req.Address.AddressLine1
		input.AddressLine2 = req.Address.AddressLine2
		input.City = req.Address.City
		input.State = req.Address.State
		input.PostalCode = req.Address.PostalCode
		if req.Address.Country != nil {
			input.Country = *req.Address.Country
		} else {
			input.Country = "United States"
		}
	} else {
		input.Country = "United States"
	}

	// Create organization
	org, err := h.service.CreateOrganization(ctx, input, userUUID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, org)
}

// GetOrganization handles GET /organizations/:id
// Retrieves organization details (public endpoint)
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	idParam := c.Param("id")
	orgID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid organization ID", map[string]interface{}{
			"id": idParam,
		}))
		return
	}

	ctx := c.Request.Context()

	org, err := h.service.GetOrganization(ctx, orgID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, org)
}

// ListOrganizations handles GET /organizations
// Retrieves a paginated list of organizations with optional filters
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	city := c.Query("city")
	state := c.Query("state")

	// Parse cause IDs if provided
	var causeIDs []uuid.UUID
	if causeParam := c.Query("cause"); causeParam != "" {
		id, err := uuid.Parse(causeParam)
		if err != nil {
			h.respondWithError(c, apperrors.NewValidationError("invalid cause ID format", map[string]interface{}{
				"cause_id": causeParam,
			}))
			return
		}
		causeIDs = append(causeIDs, id)
	}

	// Build filters
	filters := services.OrganizationListFilters{
		Search:   search,
		City:     city,
		State:    state,
		CauseIDs: causeIDs,
		Page:     page,
		Limit:    limit,
	}

	// Get organizations
	result, err := h.service.ListOrganizations(ctx, filters)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateOrganization handles PATCH /organizations/:id
// Updates organization details (admin only)
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	idParam := c.Param("id")
	orgID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid organization ID", map[string]interface{}{
			"id": idParam,
		}))
		return
	}

	var req updateOrganizationRequest
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

	// Parse cause IDs from strings to UUIDs
	var causeIDs []uuid.UUID
	if req.CauseIDs != nil {
		for _, idStr := range req.CauseIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				h.respondWithError(c, apperrors.NewValidationError("invalid cause ID format", map[string]interface{}{
					"cause_id": idStr,
				}))
				return
			}
			causeIDs = append(causeIDs, id)
		}
	}

	// Build service input
	input := services.UpdateOrganizationInput{
		Name:             req.Name,
		MissionStatement: req.MissionStatement,
		Description:      req.Description,
		Website:          req.Website,
		Email:            req.Email,
		Phone:            req.Phone,
		LogoURL:          req.LogoURL,
		BannerURL:        req.BannerURL,
		CauseIDs:         causeIDs,
	}

	// Handle address if provided
	if req.Address != nil {
		input.AddressLine1 = req.Address.AddressLine1
		input.AddressLine2 = req.Address.AddressLine2
		input.City = req.Address.City
		input.State = req.Address.State
		input.PostalCode = req.Address.PostalCode
		input.Country = req.Address.Country
	}

	// Update organization
	org, err := h.service.UpdateOrganization(ctx, orgID, input, userUUID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization handles DELETE /organizations/:id
// Soft deletes an organization (admin only)
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	idParam := c.Param("id")
	orgID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(c, apperrors.NewValidationError("invalid organization ID", map[string]interface{}{
			"id": idParam,
		}))
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

	// Delete organization
	if err := h.service.DeleteOrganization(ctx, orgID, userUUID); err != nil {
		h.respondWithError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// respondWithError is a helper to send error responses
func (h *OrganizationHandler) respondWithError(c *gin.Context, err error) {
	apperrors.HandleError(c, err)
}
