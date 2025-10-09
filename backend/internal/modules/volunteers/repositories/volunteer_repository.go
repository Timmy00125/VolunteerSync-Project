package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrVolunteerProfileNotFound is returned when a volunteer profile cannot be found
	ErrVolunteerProfileNotFound = errors.New("volunteer profile not found")
	// ErrVolunteerProfileAlreadyExists is returned when attempting to create a profile for a user who already has one
	ErrVolunteerProfileAlreadyExists = errors.New("volunteer profile already exists for this user")
	// ErrInvalidVolunteerProfileID is returned when the provided profile ID is invalid
	ErrInvalidVolunteerProfileID = errors.New("invalid volunteer profile ID")
	// ErrInvalidUserID is returned when the provided user ID is invalid
	ErrInvalidUserID = errors.New("invalid user ID")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// VolunteerRepository defines the interface for volunteer profile data access
// Following Clean Architecture principles, this interface allows for dependency injection
// and makes the service layer independent of database implementation details
type VolunteerRepository interface {
	// CreateVolunteerProfile creates a new volunteer profile for a user
	// Should be called during user registration if user type is volunteer
	// Returns ErrVolunteerProfileAlreadyExists if profile already exists
	CreateVolunteerProfile(ctx context.Context, profile *models.VolunteerProfile) error

	// FindVolunteerProfileByID retrieves a volunteer profile by its unique identifier
	// Returns ErrVolunteerProfileNotFound if no profile exists with the given ID
	FindVolunteerProfileByID(ctx context.Context, id uuid.UUID) (*models.VolunteerProfile, error)

	// FindVolunteerProfileByUserID retrieves a volunteer profile by user ID
	// Returns ErrVolunteerProfileNotFound if no profile exists for the user
	FindVolunteerProfileByUserID(ctx context.Context, userID uuid.UUID) (*models.VolunteerProfile, error)

	// UpdateVolunteerProfile updates an existing volunteer profile
	// Only updates the fields that are provided (partial updates supported)
	UpdateVolunteerProfile(ctx context.Context, profile *models.VolunteerProfile) error

	// DeleteVolunteerProfile soft deletes a volunteer profile by its ID
	DeleteVolunteerProfile(ctx context.Context, id uuid.UUID) error

	// AddSkills associates skills with a volunteer profile (many-to-many)
	// Uses the volunteer_skills junction table
	AddSkills(ctx context.Context, profileID uuid.UUID, skillIDs []uuid.UUID) error

	// RemoveSkills removes skill associations from a volunteer profile
	RemoveSkills(ctx context.Context, profileID uuid.UUID, skillIDs []uuid.UUID) error

	// GetSkills retrieves all skills associated with a volunteer profile
	GetSkills(ctx context.Context, profileID uuid.UUID) ([]uuid.UUID, error)

	// AddInterests associates cause interests with a volunteer profile (many-to-many)
	// Uses the volunteer_interests junction table
	AddInterests(ctx context.Context, profileID uuid.UUID, causeIDs []uuid.UUID) error

	// RemoveInterests removes interest associations from a volunteer profile
	RemoveInterests(ctx context.Context, profileID uuid.UUID, causeIDs []uuid.UUID) error

	// GetInterests retrieves all cause interests associated with a volunteer profile
	GetInterests(ctx context.Context, profileID uuid.UUID) ([]uuid.UUID, error)

	// IncrementTotalHours adds hours to the volunteer's total hours count
	// Called when hours are verified
	IncrementTotalHours(ctx context.Context, profileID uuid.UUID, hours float64) error

	// IncrementTotalEvents increments the total number of events attended
	// Called when an event is completed
	IncrementTotalEvents(ctx context.Context, profileID uuid.UUID) error
}

// gormVolunteerRepository is the GORM implementation of VolunteerRepository
type gormVolunteerRepository struct {
	db *gorm.DB
}

// NewVolunteerRepository creates a new instance of VolunteerRepository using GORM
func NewVolunteerRepository(db *gorm.DB) VolunteerRepository {
	return &gormVolunteerRepository{
		db: db,
	}
}

// CreateVolunteerProfile creates a new volunteer profile for a user
func (r *gormVolunteerRepository) CreateVolunteerProfile(ctx context.Context, profile *models.VolunteerProfile) error {
	// Validate input
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}

	if profile.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	// Use a transaction to ensure atomic operation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if profile already exists for this user
		var existingProfile models.VolunteerProfile
		result := tx.Where("user_id = ?", profile.UserID).First(&existingProfile)

		if result.Error == nil {
			// Profile found, already exists
			return ErrVolunteerProfileAlreadyExists
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Unexpected error during lookup
			return fmt.Errorf("failed to check existing profile: %w", result.Error)
		}

		// Create the profile
		if err := tx.Create(profile).Error; err != nil {
			return fmt.Errorf("failed to create volunteer profile: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// FindVolunteerProfileByID retrieves a volunteer profile by its unique identifier
func (r *gormVolunteerRepository) FindVolunteerProfileByID(ctx context.Context, id uuid.UUID) (*models.VolunteerProfile, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidVolunteerProfileID
	}

	var profile models.VolunteerProfile
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&profile)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrVolunteerProfileNotFound
		}
		return nil, fmt.Errorf("failed to find volunteer profile: %w", result.Error)
	}

	return &profile, nil
}

// FindVolunteerProfileByUserID retrieves a volunteer profile by user ID
func (r *gormVolunteerRepository) FindVolunteerProfileByUserID(ctx context.Context, userID uuid.UUID) (*models.VolunteerProfile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	var profile models.VolunteerProfile
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrVolunteerProfileNotFound
		}
		return nil, fmt.Errorf("failed to find volunteer profile: %w", result.Error)
	}

	return &profile, nil
}

// UpdateVolunteerProfile updates an existing volunteer profile
func (r *gormVolunteerRepository) UpdateVolunteerProfile(ctx context.Context, profile *models.VolunteerProfile) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}

	if profile.ID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	// Check if profile exists
	var existingProfile models.VolunteerProfile
	result := r.db.WithContext(ctx).Where("id = ?", profile.ID).First(&existingProfile)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrVolunteerProfileNotFound
		}
		return fmt.Errorf("failed to check existing profile: %w", result.Error)
	}

	// Update the profile (GORM will only update non-zero fields)
	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		return fmt.Errorf("failed to update volunteer profile: %w", err)
	}

	return nil
}

// DeleteVolunteerProfile soft deletes a volunteer profile by its ID
func (r *gormVolunteerRepository) DeleteVolunteerProfile(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	result := r.db.WithContext(ctx).Delete(&models.VolunteerProfile{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete volunteer profile: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrVolunteerProfileNotFound
	}

	return nil
}

// AddSkills associates skills with a volunteer profile
func (r *gormVolunteerRepository) AddSkills(ctx context.Context, profileID uuid.UUID, skillIDs []uuid.UUID) error {
	if profileID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	if len(skillIDs) == 0 {
		return nil // Nothing to add
	}

	// Use a transaction to ensure atomic operation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify profile exists
		var profile models.VolunteerProfile
		if err := tx.Where("id = ?", profileID).First(&profile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVolunteerProfileNotFound
			}
			return fmt.Errorf("failed to verify profile: %w", err)
		}

		// Insert skills (using raw SQL to handle ON CONFLICT)
		for _, skillID := range skillIDs {
			// Check if association already exists
			var count int64
			tx.Table("volunteer_skills").
				Where("volunteer_profile_id = ? AND skill_id = ?", profileID, skillID).
				Count(&count)

			if count > 0 {
				continue // Skip if already exists
			}

			// Insert new association
			if err := tx.Exec(
				"INSERT INTO volunteer_skills (volunteer_profile_id, skill_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
				profileID, skillID,
			).Error; err != nil {
				return fmt.Errorf("failed to add skill %s: %w", skillID, err)
			}
		}

		return nil
	})

	return err
}

// RemoveSkills removes skill associations from a volunteer profile
func (r *gormVolunteerRepository) RemoveSkills(ctx context.Context, profileID uuid.UUID, skillIDs []uuid.UUID) error {
	if profileID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	if len(skillIDs) == 0 {
		return nil // Nothing to remove
	}

	result := r.db.WithContext(ctx).
		Exec("DELETE FROM volunteer_skills WHERE volunteer_profile_id = ? AND skill_id = ANY(?)",
			profileID, skillIDs)

	if result.Error != nil {
		return fmt.Errorf("failed to remove skills: %w", result.Error)
	}

	return nil
}

// GetSkills retrieves all skills associated with a volunteer profile
func (r *gormVolunteerRepository) GetSkills(ctx context.Context, profileID uuid.UUID) ([]uuid.UUID, error) {
	if profileID == uuid.Nil {
		return nil, ErrInvalidVolunteerProfileID
	}

	var skillIDs []uuid.UUID
	result := r.db.WithContext(ctx).
		Table("volunteer_skills").
		Where("volunteer_profile_id = ?", profileID).
		Pluck("skill_id", &skillIDs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get skills: %w", result.Error)
	}

	return skillIDs, nil
}

// AddInterests associates cause interests with a volunteer profile
func (r *gormVolunteerRepository) AddInterests(ctx context.Context, profileID uuid.UUID, causeIDs []uuid.UUID) error {
	if profileID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	if len(causeIDs) == 0 {
		return nil // Nothing to add
	}

	// Use a transaction to ensure atomic operation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify profile exists
		var profile models.VolunteerProfile
		if err := tx.Where("id = ?", profileID).First(&profile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVolunteerProfileNotFound
			}
			return fmt.Errorf("failed to verify profile: %w", err)
		}

		// Insert interests (using raw SQL to handle duplicates)
		for _, causeID := range causeIDs {
			// Check if association already exists
			var count int64
			tx.Table("volunteer_interests").
				Where("volunteer_profile_id = ? AND cause_category_id = ?", profileID, causeID).
				Count(&count)

			if count > 0 {
				continue // Skip if already exists
			}

			// Insert new association
			if err := tx.Exec(
				"INSERT INTO volunteer_interests (volunteer_profile_id, cause_category_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
				profileID, causeID,
			).Error; err != nil {
				return fmt.Errorf("failed to add interest %s: %w", causeID, err)
			}
		}

		return nil
	})

	return err
}

// RemoveInterests removes interest associations from a volunteer profile
func (r *gormVolunteerRepository) RemoveInterests(ctx context.Context, profileID uuid.UUID, causeIDs []uuid.UUID) error {
	if profileID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	if len(causeIDs) == 0 {
		return nil // Nothing to remove
	}

	result := r.db.WithContext(ctx).
		Exec("DELETE FROM volunteer_interests WHERE volunteer_profile_id = ? AND cause_category_id = ANY(?)",
			profileID, causeIDs)

	if result.Error != nil {
		return fmt.Errorf("failed to remove interests: %w", result.Error)
	}

	return nil
}

// GetInterests retrieves all cause interests associated with a volunteer profile
func (r *gormVolunteerRepository) GetInterests(ctx context.Context, profileID uuid.UUID) ([]uuid.UUID, error) {
	if profileID == uuid.Nil {
		return nil, ErrInvalidVolunteerProfileID
	}

	var causeIDs []uuid.UUID
	result := r.db.WithContext(ctx).
		Table("volunteer_interests").
		Where("volunteer_profile_id = ?", profileID).
		Pluck("cause_category_id", &causeIDs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get interests: %w", result.Error)
	}

	return causeIDs, nil
}

// IncrementTotalHours adds hours to the volunteer's total hours count
func (r *gormVolunteerRepository) IncrementTotalHours(ctx context.Context, profileID uuid.UUID, hours float64) error {
	if profileID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	if hours <= 0 {
		return fmt.Errorf("hours must be positive")
	}

	result := r.db.WithContext(ctx).
		Model(&models.VolunteerProfile{}).
		Where("id = ?", profileID).
		UpdateColumn("total_hours", gorm.Expr("total_hours + ?", hours))

	if result.Error != nil {
		return fmt.Errorf("failed to increment total hours: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrVolunteerProfileNotFound
	}

	return nil
}

// IncrementTotalEvents increments the total number of events attended
func (r *gormVolunteerRepository) IncrementTotalEvents(ctx context.Context, profileID uuid.UUID) error {
	if profileID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	result := r.db.WithContext(ctx).
		Model(&models.VolunteerProfile{}).
		Where("id = ?", profileID).
		UpdateColumn("total_events", gorm.Expr("total_events + ?", 1))

	if result.Error != nil {
		return fmt.Errorf("failed to increment total events: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrVolunteerProfileNotFound
	}

	return nil
}
