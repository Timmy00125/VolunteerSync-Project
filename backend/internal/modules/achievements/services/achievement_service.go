package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/repositories"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// CommunicationsService defines the interface for sending notifications
// This is injected from the communications module to avoid circular dependencies
type CommunicationsService interface {
	CreateNotification(ctx context.Context, input CreateNotificationInput) error
}

// CreateNotificationInput represents input for creating a notification
type CreateNotificationInput struct {
	RecipientID       uuid.UUID
	NotificationType  string
	Title             string
	Content           string
	ActionURL         *string
	Priority          string
	RelatedEntityType *string
	RelatedEntityID   *uuid.UUID
	DeliveryMethod    string
}

// AchievementService encapsulates achievement business logic, providing methods for
// checking eligibility, awarding badges, and retrieving volunteer achievements.
// Handlers and cron jobs should depend on this interface.
type AchievementService interface {
	// CheckAndAwardAchievements checks if a volunteer is eligible for any achievements
	// and automatically awards them. This is typically called by a cron job daily.
	// Returns the list of newly awarded achievements.
	CheckAndAwardAchievements(ctx context.Context, volunteerStats VolunteerStats) ([]models.VolunteerAchievement, error)

	// GetVolunteerAchievements retrieves all achievements earned by a specific volunteer
	GetVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) ([]models.VolunteerAchievement, error)

	// GetAllAchievements lists all available achievements, optionally filtered by organization
	GetAllAchievements(ctx context.Context, organizationID *uuid.UUID) ([]models.Achievement, error)

	// AwardCustomAchievement manually awards a custom achievement to a volunteer
	// This is used by coordinators to award organization-specific badges (FR-075)
	AwardCustomAchievement(ctx context.Context, input AwardCustomAchievementInput) (*models.VolunteerAchievement, error)

	// CreateCustomAchievement creates a new organization-specific achievement
	CreateCustomAchievement(ctx context.Context, input CreateCustomAchievementInput) (*models.Achievement, error)

	// GetAchievementByID retrieves a specific achievement by its ID
	GetAchievementByID(ctx context.Context, achievementID uuid.UUID) (*models.Achievement, error)

	// CountVolunteerAchievements returns the total number of achievements a volunteer has earned
	CountVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) (int64, error)
}

// VolunteerStats contains the statistics needed to check achievement eligibility
type VolunteerStats struct {
	VolunteerProfileID      uuid.UUID
	TotalHoursLogged        float64
	TotalEventsCompleted    int
	ConsecutiveMonthsActive int // Months with at least one event
	FirstEventDate          *time.Time
	LastEventDate           *time.Time
}

// AwardCustomAchievementInput contains data for manually awarding a custom achievement
type AwardCustomAchievementInput struct {
	VolunteerProfileID uuid.UUID
	AchievementID      uuid.UUID
	AwardedByUserID    uuid.UUID // The coordinator awarding the badge
}

// CreateCustomAchievementInput contains data for creating an organization-specific achievement
type CreateCustomAchievementInput struct {
	OrganizationID uuid.UUID
	Name           string
	Description    string
	IconURL        *string
}

// achievementService is the concrete implementation of AchievementService
type achievementService struct {
	achievementRepo repositories.AchievementRepository
	commService     CommunicationsService
	logger          *logger.Logger
}

// NewAchievementService creates a new instance of AchievementService
func NewAchievementService(
	achievementRepo repositories.AchievementRepository,
	commService CommunicationsService,
	logger *logger.Logger,
) AchievementService {
	return &achievementService{
		achievementRepo: achievementRepo,
		commService:     commService,
		logger:          logger,
	}
}

// CheckAndAwardAchievements checks if a volunteer is eligible for any achievements
// and automatically awards them based on their statistics (FR-073)
func (s *achievementService) CheckAndAwardAchievements(ctx context.Context, volunteerStats VolunteerStats) ([]models.VolunteerAchievement, error) {
	s.logger.WithFields(map[string]interface{}{
		"volunteer_profile_id": volunteerStats.VolunteerProfileID,
		"total_hours":          volunteerStats.TotalHoursLogged,
		"total_events":         volunteerStats.TotalEventsCompleted,
		"consecutive_months":   volunteerStats.ConsecutiveMonthsActive,
	}).Info("Checking achievements for volunteer")

	newAchievements := []models.VolunteerAchievement{}

	// Check hours milestones
	hoursAchievements, err := s.checkHoursMilestones(ctx, volunteerStats)
	if err != nil {
		return nil, fmt.Errorf("failed to check hours milestones: %w", err)
	}
	newAchievements = append(newAchievements, hoursAchievements...)

	// Check events milestones
	eventsAchievements, err := s.checkEventsMilestones(ctx, volunteerStats)
	if err != nil {
		return nil, fmt.Errorf("failed to check events milestones: %w", err)
	}
	newAchievements = append(newAchievements, eventsAchievements...)

	// Check consistency milestones
	consistencyAchievements, err := s.checkConsistencyMilestones(ctx, volunteerStats)
	if err != nil {
		return nil, fmt.Errorf("failed to check consistency milestones: %w", err)
	}
	newAchievements = append(newAchievements, consistencyAchievements...)

	s.logger.WithFields(map[string]interface{}{
		"volunteer_profile_id":   volunteerStats.VolunteerProfileID,
		"new_achievements_count": len(newAchievements),
	}).Info("Awarded achievements")

	return newAchievements, nil
}

// checkHoursMilestones checks and awards hours-based achievements
func (s *achievementService) checkHoursMilestones(ctx context.Context, volunteerStats VolunteerStats) ([]models.VolunteerAchievement, error) {
	newAchievements := []models.VolunteerAchievement{}

	// Define hours milestones to check
	milestones := []int{1, 10, 25, 50, 100, 250, 500, 1000}

	for _, milestone := range milestones {
		// Check if volunteer has reached this milestone
		if volunteerStats.TotalHoursLogged >= float64(milestone) {
			// Find achievements matching this milestone
			achievements, err := s.achievementRepo.FindAchievementsByCriteria(
				ctx,
				models.CriteriaTypeHoursMilestone,
				milestone,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to find achievements for hours milestone %d: %w", milestone, err)
			}

			// Try to award each matching achievement
			for _, achievement := range achievements {
				awarded, err := s.awardAchievementIfNotExists(ctx, volunteerStats.VolunteerProfileID, achievement.ID, nil)
				if err != nil {
					s.logger.WithFields(map[string]interface{}{
						"error":                err,
						"achievement_id":       achievement.ID,
						"volunteer_profile_id": volunteerStats.VolunteerProfileID,
					}).Error("Failed to award hours milestone achievement")
					continue
				}
				if awarded != nil {
					newAchievements = append(newAchievements, *awarded)
				}
			}
		}
	}

	return newAchievements, nil
}

// checkEventsMilestones checks and awards events-based achievements
func (s *achievementService) checkEventsMilestones(ctx context.Context, volunteerStats VolunteerStats) ([]models.VolunteerAchievement, error) {
	newAchievements := []models.VolunteerAchievement{}

	// Define events milestones to check
	milestones := []int{1, 5, 10, 25, 50, 100}

	for _, milestone := range milestones {
		// Check if volunteer has reached this milestone
		if volunteerStats.TotalEventsCompleted >= milestone {
			// Find achievements matching this milestone
			achievements, err := s.achievementRepo.FindAchievementsByCriteria(
				ctx,
				models.CriteriaTypeEventsMilestone,
				milestone,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to find achievements for events milestone %d: %w", milestone, err)
			}

			// Try to award each matching achievement
			for _, achievement := range achievements {
				awarded, err := s.awardAchievementIfNotExists(ctx, volunteerStats.VolunteerProfileID, achievement.ID, nil)
				if err != nil {
					s.logger.WithFields(map[string]interface{}{
						"error":                err,
						"achievement_id":       achievement.ID,
						"volunteer_profile_id": volunteerStats.VolunteerProfileID,
					}).Error("Failed to award events milestone achievement")
					continue
				}
				if awarded != nil {
					newAchievements = append(newAchievements, *awarded)
				}
			}
		}
	}

	return newAchievements, nil
}

// checkConsistencyMilestones checks and awards consistency-based achievements
func (s *achievementService) checkConsistencyMilestones(ctx context.Context, volunteerStats VolunteerStats) ([]models.VolunteerAchievement, error) {
	newAchievements := []models.VolunteerAchievement{}

	// Define consistency milestones to check (in months)
	milestones := []int{3, 6, 12, 24}

	for _, milestone := range milestones {
		// Check if volunteer has reached this milestone
		if volunteerStats.ConsecutiveMonthsActive >= milestone {
			// Find achievements matching this milestone
			achievements, err := s.achievementRepo.FindAchievementsByCriteria(
				ctx,
				models.CriteriaTypeConsistency,
				milestone,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to find achievements for consistency milestone %d: %w", milestone, err)
			}

			// Try to award each matching achievement
			for _, achievement := range achievements {
				awarded, err := s.awardAchievementIfNotExists(ctx, volunteerStats.VolunteerProfileID, achievement.ID, nil)
				if err != nil {
					s.logger.WithFields(map[string]interface{}{
						"error":                err,
						"achievement_id":       achievement.ID,
						"volunteer_profile_id": volunteerStats.VolunteerProfileID,
					}).Error("Failed to award consistency milestone achievement")
					continue
				}
				if awarded != nil {
					newAchievements = append(newAchievements, *awarded)
				}
			}
		}
	}

	return newAchievements, nil
}

// awardAchievementIfNotExists awards an achievement to a volunteer if they don't already have it
// Returns the awarded achievement if new, or nil if already earned
func (s *achievementService) awardAchievementIfNotExists(
	ctx context.Context,
	volunteerProfileID, achievementID uuid.UUID,
	awardedByUserID *uuid.UUID,
) (*models.VolunteerAchievement, error) {
	// Check if volunteer already has this achievement
	hasAchievement, err := s.achievementRepo.HasVolunteerEarnedAchievement(ctx, volunteerProfileID, achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if volunteer has achievement: %w", err)
	}

	if hasAchievement {
		// Already earned, nothing to do
		return nil, nil
	}

	// Award the achievement
	volunteerAchievement := &models.VolunteerAchievement{
		VolunteerProfileID: volunteerProfileID,
		AchievementID:      achievementID,
		EarnedAt:           time.Now(),
		AwardedByUserID:    awardedByUserID,
	}

	if err := s.achievementRepo.AwardAchievement(ctx, volunteerAchievement); err != nil {
		return nil, fmt.Errorf("failed to award achievement: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"achievement_id":       achievementID,
		"volunteer_profile_id": volunteerProfileID,
	}).Info("Achievement awarded")

	// Load the achievement details for the response
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		s.logger.WithField("error", err).Warn("Failed to load achievement details after awarding")
	} else {
		volunteerAchievement.Achievement = *achievement
	}

	return volunteerAchievement, nil
}

// GetVolunteerAchievements retrieves all achievements earned by a specific volunteer
func (s *achievementService) GetVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) ([]models.VolunteerAchievement, error) {
	achievements, err := s.achievementRepo.FindVolunteerAchievements(ctx, volunteerProfileID)
	if err != nil {
		return nil, apperrors.NewInternalServerError("failed to retrieve volunteer achievements").WithError(err)
	}

	return achievements, nil
}

// GetAllAchievements lists all available achievements, optionally filtered by organization
func (s *achievementService) GetAllAchievements(ctx context.Context, organizationID *uuid.UUID) ([]models.Achievement, error) {
	achievements, err := s.achievementRepo.ListAchievements(ctx, organizationID)
	if err != nil {
		return nil, apperrors.NewInternalServerError("failed to retrieve achievements").WithError(err)
	}

	return achievements, nil
}

// AwardCustomAchievement manually awards a custom achievement to a volunteer (FR-075)
func (s *achievementService) AwardCustomAchievement(ctx context.Context, input AwardCustomAchievementInput) (*models.VolunteerAchievement, error) {
	// Verify the achievement exists and is a custom badge
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, input.AchievementID)
	if err != nil {
		if err == repositories.ErrAchievementNotFound {
			return nil, apperrors.NewNotFoundError("achievement")
		}
		return nil, apperrors.NewInternalServerError("failed to retrieve achievement").WithError(err)
	}

	// Ensure it's a custom achievement (not a system badge that should be auto-awarded)
	if achievement.BadgeType != models.BadgeTypeOrganizationCustom {
		return nil, apperrors.NewValidationError("only custom achievements can be manually awarded", nil)
	}

	// Award the achievement
	volunteerAchievement := &models.VolunteerAchievement{
		VolunteerProfileID: input.VolunteerProfileID,
		AchievementID:      input.AchievementID,
		EarnedAt:           time.Now(),
		AwardedByUserID:    &input.AwardedByUserID,
	}

	if err := s.achievementRepo.AwardAchievement(ctx, volunteerAchievement); err != nil {
		if err == repositories.ErrVolunteerAchievementAlreadyAwarded {
			return nil, apperrors.NewValidationError("volunteer has already earned this achievement", nil)
		}
		return nil, apperrors.NewInternalServerError("failed to award achievement").WithError(err)
	}

	volunteerAchievement.Achievement = *achievement

	s.logger.WithFields(map[string]interface{}{
		"achievement_id":       input.AchievementID,
		"volunteer_profile_id": input.VolunteerProfileID,
		"awarded_by_user_id":   input.AwardedByUserID,
	}).Info("Custom achievement awarded")

	// Send notification to volunteer (FR-076)
	if s.commService != nil {
		actionURL := "/volunteer/achievements"
		entityType := "achievement"
		notificationInput := CreateNotificationInput{
			RecipientID:       input.VolunteerProfileID,
			NotificationType:  "achievement_earned",
			Title:             "Achievement Unlocked! 🏆",
			Content:           fmt.Sprintf("Congratulations! You've earned the '%s' achievement: %s", achievement.Name, achievement.Description),
			ActionURL:         &actionURL,
			Priority:          "medium",
			RelatedEntityType: &entityType,
			RelatedEntityID:   &input.AchievementID,
			DeliveryMethod:    "in_app",
		}

		if err := s.commService.CreateNotification(ctx, notificationInput); err != nil {
			// Log error but don't fail the achievement award operation
			s.logger.WithFields(map[string]interface{}{
				"achievement_id":       input.AchievementID,
				"volunteer_profile_id": input.VolunteerProfileID,
				"error":                err.Error(),
			}).Error("Failed to send achievement notification")
		} else {
			s.logger.WithFields(map[string]interface{}{
				"achievement_id":       input.AchievementID,
				"volunteer_profile_id": input.VolunteerProfileID,
			}).Info("Achievement notification sent successfully")
		}
	}

	return volunteerAchievement, nil
}

// CreateCustomAchievement creates a new organization-specific achievement
func (s *achievementService) CreateCustomAchievement(ctx context.Context, input CreateCustomAchievementInput) (*models.Achievement, error) {
	// Validate input
	if input.Name == "" {
		return nil, apperrors.NewValidationError("achievement name is required", nil)
	}
	if input.Description == "" {
		return nil, apperrors.NewValidationError("achievement description is required", nil)
	}

	achievement := &models.Achievement{
		OrganizationID: &input.OrganizationID,
		Name:           input.Name,
		Description:    input.Description,
		IconURL:        input.IconURL,
		BadgeType:      models.BadgeTypeOrganizationCustom,
		CriteriaType:   ptrCriteriaType(models.CriteriaTypeCustom),
	}

	if err := s.achievementRepo.CreateAchievement(ctx, achievement); err != nil {
		return nil, apperrors.NewInternalServerError("failed to create achievement").WithError(err)
	}

	s.logger.WithFields(map[string]interface{}{
		"achievement_id":  achievement.ID,
		"organization_id": input.OrganizationID,
		"name":            input.Name,
	}).Info("Custom achievement created")

	return achievement, nil
}

// GetAchievementByID retrieves a specific achievement by its ID
func (s *achievementService) GetAchievementByID(ctx context.Context, achievementID uuid.UUID) (*models.Achievement, error) {
	achievement, err := s.achievementRepo.FindAchievementByID(ctx, achievementID)
	if err != nil {
		if err == repositories.ErrAchievementNotFound {
			return nil, apperrors.NewNotFoundError("achievement")
		}
		return nil, apperrors.NewInternalServerError("failed to retrieve achievement").WithError(err)
	}

	return achievement, nil
}

// CountVolunteerAchievements returns the total number of achievements a volunteer has earned
func (s *achievementService) CountVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) (int64, error) {
	count, err := s.achievementRepo.CountVolunteerAchievements(ctx, volunteerProfileID)
	if err != nil {
		return 0, apperrors.NewInternalServerError("failed to count volunteer achievements").WithError(err)
	}

	return count, nil
}

// Helper function to create pointer to CriteriaType
func ptrCriteriaType(ct models.CriteriaType) *models.CriteriaType {
	return &ct
}
