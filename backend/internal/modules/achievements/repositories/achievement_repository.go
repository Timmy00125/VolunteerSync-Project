package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrAchievementNotFound is returned when an achievement cannot be found
	ErrAchievementNotFound = errors.New("achievement not found")
	// ErrAchievementAlreadyExists is returned when attempting to create a duplicate achievement
	ErrAchievementAlreadyExists = errors.New("achievement already exists")
	// ErrVolunteerAchievementAlreadyAwarded is returned when a volunteer already has a specific badge
	ErrVolunteerAchievementAlreadyAwarded = errors.New("volunteer has already earned this achievement")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// AchievementRepository defines the interface for achievement data access
// Following Clean Architecture principles, this interface allows for dependency injection
type AchievementRepository interface {
	// CreateAchievement creates a new achievement (badge) in the database
	// Used for creating system badges and organization-specific custom badges
	CreateAchievement(ctx context.Context, achievement *models.Achievement) error

	// FindAchievementByID retrieves an achievement by its unique identifier
	FindAchievementByID(ctx context.Context, achievementID uuid.UUID) (*models.Achievement, error)

	// ListAchievements retrieves all achievements, optionally filtered by badge type or organization
	// If organizationID is nil, returns all system badges
	// If organizationID is provided, returns system badges + that org's custom badges
	ListAchievements(ctx context.Context, organizationID *uuid.UUID) ([]models.Achievement, error)

	// ListSystemAchievements retrieves all system-wide achievements
	ListSystemAchievements(ctx context.Context) ([]models.Achievement, error)

	// AwardAchievement creates a new VolunteerAchievement record to award a badge to a volunteer
	// Returns ErrVolunteerAchievementAlreadyAwarded if the volunteer already has this badge
	AwardAchievement(ctx context.Context, volunteerAchievement *models.VolunteerAchievement) error

	// FindVolunteerAchievements retrieves all achievements earned by a specific volunteer
	// Includes the achievement details via preloading
	FindVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) ([]models.VolunteerAchievement, error)

	// FindVolunteerAchievementsByIDs retrieves achievements for multiple volunteers
	// Useful for batch operations or displaying multiple volunteer profiles
	FindVolunteerAchievementsByIDs(ctx context.Context, volunteerProfileIDs []uuid.UUID) (map[uuid.UUID][]models.VolunteerAchievement, error)

	// HasVolunteerEarnedAchievement checks if a volunteer has already earned a specific achievement
	HasVolunteerEarnedAchievement(ctx context.Context, volunteerProfileID, achievementID uuid.UUID) (bool, error)

	// CountVolunteerAchievements counts how many achievements a volunteer has earned
	CountVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) (int64, error)

	// FindAchievementsByCriteria finds achievements that match specific criteria type and value
	// Used by the auto-award system to find eligible badges based on volunteer stats
	FindAchievementsByCriteria(ctx context.Context, criteriaType models.CriteriaType, criteriaValue int) ([]models.Achievement, error)
}

// gormAchievementRepository is the GORM implementation of AchievementRepository
type gormAchievementRepository struct {
	db *gorm.DB
}

// NewAchievementRepository creates a new instance of AchievementRepository using GORM
func NewAchievementRepository(db *gorm.DB) AchievementRepository {
	return &gormAchievementRepository{
		db: db,
	}
}

// CreateAchievement creates a new achievement in the database
func (r *gormAchievementRepository) CreateAchievement(ctx context.Context, achievement *models.Achievement) error {
	if achievement == nil {
		return fmt.Errorf("achievement cannot be nil")
	}

	if achievement.Name == "" {
		return fmt.Errorf("achievement name is required")
	}

	if err := r.db.WithContext(ctx).Create(achievement).Error; err != nil {
		return fmt.Errorf("%w: failed to create achievement: %v", ErrDatabaseOperation, err)
	}

	return nil
}

// FindAchievementByID retrieves an achievement by its unique identifier
func (r *gormAchievementRepository) FindAchievementByID(ctx context.Context, achievementID uuid.UUID) (*models.Achievement, error) {
	var achievement models.Achievement

	if err := r.db.WithContext(ctx).
		Where("id = ?", achievementID).
		First(&achievement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAchievementNotFound
		}
		return nil, fmt.Errorf("%w: failed to find achievement: %v", ErrDatabaseOperation, err)
	}

	return &achievement, nil
}

// ListAchievements retrieves achievements filtered by organization
// If organizationID is nil, returns only system badges
// If organizationID is provided, returns system badges + that org's custom badges
func (r *gormAchievementRepository) ListAchievements(ctx context.Context, organizationID *uuid.UUID) ([]models.Achievement, error) {
	var achievements []models.Achievement

	query := r.db.WithContext(ctx)

	if organizationID == nil {
		// Return only system badges
		query = query.Where("badge_type = ?", models.BadgeTypeSystem)
	} else {
		// Return system badges AND this organization's custom badges
		query = query.Where(
			"badge_type = ? OR (badge_type = ? AND organization_id = ?)",
			models.BadgeTypeSystem,
			models.BadgeTypeOrganizationCustom,
			*organizationID,
		)
	}

	if err := query.Order("created_at ASC").Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("%w: failed to list achievements: %v", ErrDatabaseOperation, err)
	}

	return achievements, nil
}

// ListSystemAchievements retrieves all system-wide achievements
func (r *gormAchievementRepository) ListSystemAchievements(ctx context.Context) ([]models.Achievement, error) {
	var achievements []models.Achievement

	if err := r.db.WithContext(ctx).
		Where("badge_type = ?", models.BadgeTypeSystem).
		Order("created_at ASC").
		Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("%w: failed to list system achievements: %v", ErrDatabaseOperation, err)
	}

	return achievements, nil
}

// AwardAchievement creates a new VolunteerAchievement record to award a badge to a volunteer
func (r *gormAchievementRepository) AwardAchievement(ctx context.Context, volunteerAchievement *models.VolunteerAchievement) error {
	if volunteerAchievement == nil {
		return fmt.Errorf("volunteer achievement cannot be nil")
	}

	// Use a transaction to ensure atomicity and check for duplicates
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if volunteer already has this achievement
		var existing models.VolunteerAchievement
		result := tx.Where(
			"volunteer_profile_id = ? AND achievement_id = ?",
			volunteerAchievement.VolunteerProfileID,
			volunteerAchievement.AchievementID,
		).First(&existing)

		if result.Error == nil {
			// Achievement already awarded
			return ErrVolunteerAchievementAlreadyAwarded
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Unexpected error during lookup
			return fmt.Errorf("failed to check existing achievement: %w", result.Error)
		}

		// Set EarnedAt to current time if not provided
		if volunteerAchievement.EarnedAt.IsZero() {
			volunteerAchievement.EarnedAt = time.Now()
		}

		// Award the achievement
		if err := tx.Create(volunteerAchievement).Error; err != nil {
			return fmt.Errorf("failed to award achievement: %w", err)
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrVolunteerAchievementAlreadyAwarded) {
			return err
		}
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return nil
}

// FindVolunteerAchievements retrieves all achievements earned by a specific volunteer
func (r *gormAchievementRepository) FindVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) ([]models.VolunteerAchievement, error) {
	var volunteerAchievements []models.VolunteerAchievement

	if err := r.db.WithContext(ctx).
		Preload("Achievement"). // Load achievement details
		Where("volunteer_profile_id = ?", volunteerProfileID).
		Order("earned_at DESC"). // Most recent first
		Find(&volunteerAchievements).Error; err != nil {
		return nil, fmt.Errorf("%w: failed to find volunteer achievements: %v", ErrDatabaseOperation, err)
	}

	return volunteerAchievements, nil
}

// FindVolunteerAchievementsByIDs retrieves achievements for multiple volunteers
func (r *gormAchievementRepository) FindVolunteerAchievementsByIDs(ctx context.Context, volunteerProfileIDs []uuid.UUID) (map[uuid.UUID][]models.VolunteerAchievement, error) {
	if len(volunteerProfileIDs) == 0 {
		return make(map[uuid.UUID][]models.VolunteerAchievement), nil
	}

	var volunteerAchievements []models.VolunteerAchievement

	if err := r.db.WithContext(ctx).
		Preload("Achievement").
		Where("volunteer_profile_id IN ?", volunteerProfileIDs).
		Order("earned_at DESC").
		Find(&volunteerAchievements).Error; err != nil {
		return nil, fmt.Errorf("%w: failed to find volunteer achievements: %v", ErrDatabaseOperation, err)
	}

	// Group by volunteer profile ID
	result := make(map[uuid.UUID][]models.VolunteerAchievement)
	for _, va := range volunteerAchievements {
		result[va.VolunteerProfileID] = append(result[va.VolunteerProfileID], va)
	}

	return result, nil
}

// HasVolunteerEarnedAchievement checks if a volunteer has already earned a specific achievement
func (r *gormAchievementRepository) HasVolunteerEarnedAchievement(ctx context.Context, volunteerProfileID, achievementID uuid.UUID) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&models.VolunteerAchievement{}).
		Where("volunteer_profile_id = ? AND achievement_id = ?", volunteerProfileID, achievementID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%w: failed to check volunteer achievement: %v", ErrDatabaseOperation, err)
	}

	return count > 0, nil
}

// CountVolunteerAchievements counts how many achievements a volunteer has earned
func (r *gormAchievementRepository) CountVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&models.VolunteerAchievement{}).
		Where("volunteer_profile_id = ?", volunteerProfileID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%w: failed to count volunteer achievements: %v", ErrDatabaseOperation, err)
	}

	return count, nil
}

// FindAchievementsByCriteria finds achievements that match specific criteria type and value
// Used by the auto-award system to find eligible badges based on volunteer stats
func (r *gormAchievementRepository) FindAchievementsByCriteria(ctx context.Context, criteriaType models.CriteriaType, criteriaValue int) ([]models.Achievement, error) {
	var achievements []models.Achievement

	if err := r.db.WithContext(ctx).
		Where("criteria_type = ? AND criteria_value = ?", criteriaType, criteriaValue).
		Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("%w: failed to find achievements by criteria: %v", ErrDatabaseOperation, err)
	}

	return achievements, nil
}
