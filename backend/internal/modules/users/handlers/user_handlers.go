package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/middleware"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/users/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// UserHandler exposes HTTP handlers for user profile management.
// Handles operations on the current authenticated user's profile.
type UserHandler struct {
	service services.UserService
	log     *logger.Logger
}

// NewUserHandler constructs a UserHandler with required dependencies.
// The service parameter is required and cannot be nil.
// If log is nil, the global logger will be used.
func NewUserHandler(service services.UserService, log *logger.Logger) (*UserHandler, error) {
	if service == nil {
		return nil, apperrors.NewInternalServerError("user handler requires user service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &UserHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires user profile routes under the provided router group.
// Expected to be mounted at /users and requires authentication middleware.
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// All routes require authentication
	rg.GET("/me", h.GetCurrentUser)
	rg.PATCH("/me", h.UpdateCurrentUser)
	rg.DELETE("/me/delete", h.DeleteCurrentUser)
}

// updateUserRequest represents the request body for updating user profile.
// All fields are optional to support partial updates per REST best practices.
type updateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty"`
}

// GetCurrentUser handles GET /users/me.
// Retrieves the authenticated user's profile information.
// Requires the user ID to be set in the Gin context by auth middleware.
//
// Success Response (200):
//
//	{
//	  "id": "uuid",
//	  "email": "user@example.com",
//	  "first_name": "John",
//	  "last_name": "Doe",
//	  "phone": "+1234567890",
//	  "account_status": "active",
//	  "email_verified": true,
//	  "last_login_at": "2024-01-15T10:30:00Z",
//	  "created_at": "2024-01-01T00:00:00Z",
//	  "updated_at": "2024-01-15T10:30:00Z"
//	}
//
// Error Responses:
// - 401 Unauthorized: Missing or invalid authentication
// - 403 Forbidden: Account is not active
// - 404 Not Found: User not found
// - 500 Internal Server Error: Server error
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// Extract user UUID from context (set by auth and context enrichment middleware)
	userID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Retrieve user profile from service
	user, err := h.service.GetCurrentUser(ctx, userID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Return user profile
	// The User model's json tags control serialization
	// Sensitive fields like password_hash are excluded via json:"-"
	c.JSON(http.StatusOK, user)
}

// UpdateCurrentUser handles PATCH /users/me.
// Updates the authenticated user's profile information.
// Supports partial updates - only provided fields are updated.
// Email changes require reverification per FR-007.
//
// Request Body (all fields optional):
//
//	{
//	  "first_name": "Jane",
//	  "last_name": "Smith",
//	  "phone": "+1987654321",
//	  "email": "newemail@example.com"
//	}
//
// Success Response (200): Returns updated user object (same format as GET /users/me)
//
// Error Responses:
// - 400 Bad Request: Invalid request payload or validation error
// - 401 Unauthorized: Missing or invalid authentication
// - 403 Forbidden: Account is not active
// - 404 Not Found: User not found
// - 409 Conflict: Email already in use by another user
// - 500 Internal Server Error: Server error
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	// Extract user UUID from context (set by auth and context enrichment middleware)
	userID := middleware.MustGetUserUUID(c)

	// Bind and validate request body
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("Invalid request payload").WithError(err))
		return
	}

	// Validate that at least one field is provided
	if req.FirstName == nil && req.LastName == nil && req.Phone == nil && req.Email == nil {
		h.respondWithError(c, apperrors.NewBadRequestError("At least one field must be provided for update"))
		return
	}

	ctx := c.Request.Context()

	// Convert request to service input
	updates := &services.UpdateUserProfileRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Email:     req.Email,
	}

	// Update user profile via service
	user, err := h.service.UpdateUserProfile(ctx, userID, updates)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Return updated user profile
	c.JSON(http.StatusOK, user)
}

// DeleteCurrentUser handles DELETE /users/me/delete.
// Performs a soft delete of the authenticated user's account.
// Implements data retention compliance per FR-009 and FR-107.
//
// The account is not physically deleted but marked as deleted (soft delete).
// This allows for:
// - Data retention for legal/compliance purposes
// - Potential account recovery
// - Audit trail maintenance
//
// Success Response (200):
//
//	{
//	  "message": "Account deletion requested"
//	}
//
// Error Responses:
// - 401 Unauthorized: Missing or invalid authentication
// - 404 Not Found: User not found
// - 500 Internal Server Error: Server error
//
// Note: In a production system, this operation might:
// - Invalidate all active sessions/tokens
// - Schedule deletion of associated data per retention policy
// - Send confirmation email
// - Trigger cleanup of related resources
func (h *UserHandler) DeleteCurrentUser(c *gin.Context) {
	// Extract user UUID from context (set by auth and context enrichment middleware)
	userID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Delete user account via service
	if err := h.service.DeleteUserAccount(ctx, userID); err != nil {
		h.respondWithError(c, err)
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Account deletion requested",
	})
}

// respondWithError is a helper method to handle error responses consistently.
// It uses the application error package to convert errors to appropriate HTTP responses.
func (h *UserHandler) respondWithError(c *gin.Context, err error) {
	apperrors.AbortWithError(c, err)
}
