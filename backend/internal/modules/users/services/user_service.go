package services

import (
	"context"
	"errors"

	authModels "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/models"
	authRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/repositories"
	appErrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService defines the interface for user profile management operations
// This service handles user profile operations (get, update, delete)
// as distinct from authentication operations (handled by AuthService)
type UserService interface {
	// GetCurrentUser retrieves the authenticated user's profile by user ID
	// Returns the user profile or an error if not found
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*authModels.User, error)

	// UpdateUserProfile updates the authenticated user's profile information
	// Supports updating: first_name, last_name, phone, email
	// Email changes require reverification per FR-007
	// Returns the updated user profile or an error
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, updates *UpdateUserProfileRequest) (*authModels.User, error)

	// DeleteUserAccount performs a soft delete of the user account
	// Implements data retention compliance per FR-009 and FR-107
	// Returns an error if the operation fails
	DeleteUserAccount(ctx context.Context, userID uuid.UUID) error
}

// UpdateUserProfileRequest contains the fields that can be updated in a user profile
type UpdateUserProfileRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty"`
}

// userService is the concrete implementation of UserService
type userService struct {
	authRepo authRepos.AuthRepository
	db       *gorm.DB
	log      logger.Logger
}

// NewUserService creates a new instance of UserService
// Dependencies are injected for testability and loose coupling
func NewUserService(
	authRepo authRepos.AuthRepository,
	db *gorm.DB,
	log logger.Logger,
) UserService {
	return &userService{
		authRepo: authRepo,
		db:       db,
		log:      log,
	}
}

// GetCurrentUser retrieves the authenticated user's profile by user ID
// This is a simple delegation to the auth repository
func (s *userService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*authModels.User, error) {
	log := s.log.WithContext(ctx).WithField("user_id", userID.String())
	log.Info("fetching user profile")

	// Validate user ID
	if userID == uuid.Nil {
		log.Error("invalid user ID: nil UUID")
		return nil, appErrors.NewBadRequestError("User ID is required")
	}

	// Retrieve user from repository
	user, err := s.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, authRepos.ErrUserNotFound) {
			log.Warn("user not found")
			return nil, appErrors.NewNotFoundError("User")
		}
		log.ErrorWithErr("failed to retrieve user", err)
		return nil, appErrors.NewInternalServerError("Failed to retrieve user profile")
	}

	// Check if account is active
	if !user.IsActive() {
		log.WithField("status", string(user.AccountStatus)).Warn("inactive account access attempt")
		return nil, appErrors.NewForbiddenError("Account is not active")
	}

	log.Info("user profile retrieved successfully")
	return user, nil
}

// UpdateUserProfile updates the authenticated user's profile information
// Supports partial updates - only provided fields are updated
// Email changes require reverification per FR-007
func (s *userService) UpdateUserProfile(ctx context.Context, userID uuid.UUID, updates *UpdateUserProfileRequest) (*authModels.User, error) {
	log := s.log.WithContext(ctx).WithField("user_id", userID.String())
	log.Info("updating user profile")

	// Validate user ID
	if userID == uuid.Nil {
		log.Error("invalid user ID: nil UUID")
		return nil, appErrors.NewBadRequestError("User ID is required")
	}

	// Validate request
	if updates == nil {
		log.Error("nil update request")
		return nil, appErrors.NewBadRequestError("Update data is required")
	}

	// Retrieve current user
	user, err := s.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, authRepos.ErrUserNotFound) {
			log.Warn("user not found for update")
			return nil, appErrors.NewNotFoundError("User")
		}
		log.ErrorWithErr("failed to retrieve user for update", err)
		return nil, appErrors.NewInternalServerError("Failed to retrieve user profile")
	}

	// Check if account is active
	if !user.IsActive() {
		log.WithField("status", string(user.AccountStatus)).Warn("update attempt on inactive account")
		return nil, appErrors.NewForbiddenError("Account is not active")
	}

	// Track if email is being changed
	emailChanged := false
	oldEmail := user.Email

	// Apply updates (partial update pattern)
	if updates.FirstName != nil && *updates.FirstName != "" {
		user.FirstName = *updates.FirstName
		log.Debug("updating first name")
	}

	if updates.LastName != nil && *updates.LastName != "" {
		user.LastName = *updates.LastName
		log.Debug("updating last name")
	}

	if updates.Phone != nil {
		user.Phone = updates.Phone
		log.Debug("updating phone")
	}

	// Handle email change - requires reverification per FR-007
	if updates.Email != nil && *updates.Email != "" && *updates.Email != user.Email {
		// Validate email format (basic check)
		if !isValidEmail(*updates.Email) {
			log.WithField("email", *updates.Email).Warn("invalid email format")
			return nil, appErrors.NewBadRequestError("Invalid email format")
		}

		// Check if email is already taken by another user
		existingUser, err := s.authRepo.FindUserByEmail(ctx, *updates.Email)
		if err != nil && !errors.Is(err, authRepos.ErrUserNotFound) {
			log.ErrorWithErr("failed to check email uniqueness", err)
			return nil, appErrors.NewInternalServerError("Failed to verify email availability")
		}

		if existingUser != nil && existingUser.ID != userID {
			log.WithField("email", *updates.Email).Warn("email already in use")
			return nil, appErrors.NewConflictError("Email address is already in use")
		}

		user.Email = *updates.Email
		user.EmailVerified = false // Require reverification per FR-007
		emailChanged = true
		log.WithFields(map[string]interface{}{
			"old_email": oldEmail,
			"new_email": *updates.Email,
		}).Info("email changed, reverification required")
	}

	// Save updates to database
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		log.ErrorWithErr("failed to save user updates", err)
		return nil, appErrors.NewInternalServerError("Failed to update user profile")
	}

	// Log email change for audit trail
	if emailChanged {
		log.WithFields(map[string]interface{}{
			"old_email":      oldEmail,
			"new_email":      user.Email,
			"email_verified": user.EmailVerified,
		}).Info("user email changed successfully")
	}

	log.Info("user profile updated successfully")
	return user, nil
}

// DeleteUserAccount performs a soft delete of the user account
// Implements GORM soft delete pattern (sets deleted_at timestamp)
// This allows for data retention and potential account recovery per FR-009
func (s *userService) DeleteUserAccount(ctx context.Context, userID uuid.UUID) error {
	log := s.log.WithContext(ctx).WithField("user_id", userID.String())
	log.Info("deleting user account")

	// Validate user ID
	if userID == uuid.Nil {
		log.Error("invalid user ID: nil UUID")
		return appErrors.NewBadRequestError("User ID is required")
	}

	// Retrieve user to ensure it exists
	user, err := s.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, authRepos.ErrUserNotFound) {
			log.Warn("user not found for deletion")
			return appErrors.NewNotFoundError("User")
		}
		log.ErrorWithErr("failed to retrieve user for deletion", err)
		return appErrors.NewInternalServerError("Failed to retrieve user")
	}

	// Perform soft delete using GORM's soft delete feature
	// This sets the deleted_at timestamp instead of actually removing the record
	// Complies with FR-009 (data retention) and FR-107 (compliance)
	if err := s.db.WithContext(ctx).Delete(user).Error; err != nil {
		log.ErrorWithErr("failed to delete user account", err)
		return appErrors.NewInternalServerError("Failed to delete user account")
	}

	log.WithFields(map[string]interface{}{
		"email":          user.Email,
		"data_retention": "enabled",
	}).Info("user account deleted successfully (soft delete)")

	// Note: In a production system, you might want to:
	// 1. Invalidate all active sessions/tokens
	// 2. Schedule deletion of associated data per retention policy
	// 3. Send confirmation email
	// 4. Trigger cleanup of related resources (volunteer profile, registrations, etc.)
	// These operations could be handled by a background job or event system

	return nil
}

// isValidEmail performs basic email validation
// For production use, consider using a more robust email validation library
func isValidEmail(email string) bool {
	// Basic validation: must contain @ and have characters before and after
	// More sophisticated validation could use regex or dedicated libraries
	if len(email) < 3 {
		return false
	}

	atIndex := -1
	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				// Multiple @ symbols
				return false
			}
			atIndex = i
		}
	}

	// Must have @ with characters before and after
	return atIndex > 0 && atIndex < len(email)-1
}
