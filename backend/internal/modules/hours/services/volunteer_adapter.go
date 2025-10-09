package services

import (
	"context"
	"fmt"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/repositories"
	"github.com/google/uuid"
)

// volunteerServiceAdapter adapts volunteer repository access to satisfy
// the VolunteerService interface expected by the hours service.
// This maintains module boundaries while enabling volunteer hours updates.
//
// Note: We use repository directly instead of service to avoid circular dependencies
// and because this is a simple data update operation that doesn't require business logic.
type volunteerServiceAdapter struct {
	volunteerRepo repositories.VolunteerRepository
}

// NewVolunteerServiceAdapter creates a new adapter for volunteer operations
func NewVolunteerServiceAdapter(volunteerRepo repositories.VolunteerRepository) VolunteerService {
	return &volunteerServiceAdapter{
		volunteerRepo: volunteerRepo,
	}
}

// IncrementTotalHours increments the total hours for a volunteer profile
// This is called after hours are verified to update the volunteer's cumulative hours
func (a *volunteerServiceAdapter) IncrementTotalHours(ctx context.Context, volunteerProfileID uuid.UUID, hours float64) error {
	if volunteerProfileID == uuid.Nil {
		return fmt.Errorf("volunteer profile ID is required")
	}

	if hours <= 0 {
		return fmt.Errorf("hours must be positive")
	}

	// Get the current volunteer profile
	profile, err := a.volunteerRepo.FindVolunteerProfileByID(ctx, volunteerProfileID)
	if err != nil {
		return fmt.Errorf("failed to get volunteer profile: %w", err)
	}

	// Increment total hours
	profile.TotalHours += hours

	// Update the profile
	if err := a.volunteerRepo.UpdateVolunteerProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to update volunteer total hours: %w", err)
	}

	return nil
}
