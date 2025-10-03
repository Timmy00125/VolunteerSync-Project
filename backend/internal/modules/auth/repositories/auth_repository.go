package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when attempting to create a user with an existing email
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	// ErrInvalidUserID is returned when the provided user ID is invalid
	ErrInvalidUserID = errors.New("invalid user ID")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// AuthRepository defines the interface for user authentication data access
// Following Clean Architecture principles, this interface allows for dependency injection
// and makes the service layer independent of database implementation details
type AuthRepository interface {
	// CreateUser creates a new user in the database with transaction support
	// Returns the created user or an error if the operation fails
	CreateUser(ctx context.Context, user *models.User) error

	// FindUserByEmail retrieves a user by their email address
	// Returns ErrUserNotFound if no user exists with the given email
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)

	// FindUserByID retrieves a user by their unique identifier
	// Returns ErrUserNotFound if no user exists with the given ID
	FindUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// UpdatePassword updates a user's password hash
	// The password should already be hashed before calling this method
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error

	// GetSecurityQuestions retrieves the three security questions for a user
	// Returns the questions (without the answers) for password reset flow
	GetSecurityQuestions(ctx context.Context, email string) ([]string, error)

	// UpdateLastLogin updates the user's last login timestamp to the current time
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error

	// UpdateAccountStatus updates the user's account status (active, inactive, suspended)
	UpdateAccountStatus(ctx context.Context, userID uuid.UUID, status models.AccountStatus) error
}

// gormAuthRepository is the GORM implementation of AuthRepository
type gormAuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new instance of AuthRepository using GORM
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &gormAuthRepository{
		db: db,
	}
}

// CreateUser creates a new user in the database with transaction support
// This ensures atomic operations and prevents partial user creation
func (r *gormAuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Validate input
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Use a transaction to ensure atomic operation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if user with email already exists
		var existingUser models.User
		result := tx.Where("email = ?", user.Email).First(&existingUser)

		if result.Error == nil {
			// User found, email already exists
			return ErrUserAlreadyExists
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Unexpected error during lookup
			return fmt.Errorf("failed to check existing user: %w", result.Error)
		}

		// Create the user
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return err
		}
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return nil
}

// FindUserByEmail retrieves a user by their email address
func (r *gormAuthRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: failed to find user by email: %v", ErrDatabaseOperation, result.Error)
	}

	return &user, nil
}

// FindUserByID retrieves a user by their unique identifier
func (r *gormAuthRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	var user models.User
	result := r.db.WithContext(ctx).First(&user, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: failed to find user by ID: %v", ErrDatabaseOperation, result.Error)
	}

	return &user, nil
}

// UpdatePassword updates a user's password hash
func (r *gormAuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}

	if passwordHash == "" {
		return fmt.Errorf("password hash is required")
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash)

	if result.Error != nil {
		return fmt.Errorf("%w: failed to update password: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetSecurityQuestions retrieves the three security questions for a user
func (r *gormAuthRepository) GetSecurityQuestions(ctx context.Context, email string) ([]string, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	var user models.User
	result := r.db.WithContext(ctx).
		Select("security_question_1", "security_question_2", "security_question_3").
		Where("email = ?", email).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: failed to get security questions: %v", ErrDatabaseOperation, result.Error)
	}

	return user.GetSecurityQuestions(), nil
}

// UpdateLastLogin updates the user's last login timestamp to the current time
func (r *gormAuthRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now)

	if result.Error != nil {
		return fmt.Errorf("%w: failed to update last login: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateAccountStatus updates the user's account status
func (r *gormAuthRepository) UpdateAccountStatus(ctx context.Context, userID uuid.UUID, status models.AccountStatus) error {
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}

	// Validate status enum
	validStatuses := map[models.AccountStatus]bool{
		models.AccountStatusActive:    true,
		models.AccountStatusInactive:  true,
		models.AccountStatusSuspended: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid account status: %s", status)
	}

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("account_status", status)

	if result.Error != nil {
		return fmt.Errorf("%w: failed to update account status: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
