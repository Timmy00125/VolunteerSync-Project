package services

import (
	"context"
	"fmt"

	regRepos "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/repositories"
	"github.com/google/uuid"
)

// registrationRepositoryAdapter adapts the registrations module's repository to satisfy
// the RegistrationRepository interface expected by the communications service.
// This maintains module boundaries while enabling broadcast message functionality.
type registrationRepositoryAdapter struct {
	regRepo regRepos.RegistrationRepository
}

// NewRegistrationRepositoryAdapter creates a new adapter for the registration repository
func NewRegistrationRepositoryAdapter(regRepo regRepos.RegistrationRepository) RegistrationRepository {
	return &registrationRepositoryAdapter{
		regRepo: regRepo,
	}
}

// FindVolunteersByOpportunity retrieves all volunteer profile IDs registered for an opportunity
// This is used for sending broadcast messages to all event volunteers
func (a *registrationRepositoryAdapter) FindVolunteersByOpportunity(ctx context.Context, opportunityID uuid.UUID) ([]uuid.UUID, error) {
	if opportunityID == uuid.Nil {
		return nil, fmt.Errorf("opportunity ID is required")
	}

	// Get all registrations for the opportunity
	registrations, err := a.regRepo.FindRegistrationsByOpportunity(ctx, opportunityID)
	if err != nil {
		return nil, fmt.Errorf("failed to find registrations: %w", err)
	}

	// Extract unique volunteer profile IDs
	// Use a map to deduplicate in case of multiple registrations
	volunteerMap := make(map[uuid.UUID]bool)
	for _, reg := range registrations {
		// Only include confirmed or checked-in volunteers, not cancelled or waitlisted
		if reg.Status == "confirmed" || reg.Status == "attended" {
			volunteerMap[reg.VolunteerProfileID] = true
		}
	}

	// Convert map to slice
	volunteerIDs := make([]uuid.UUID, 0, len(volunteerMap))
	for id := range volunteerMap {
		volunteerIDs = append(volunteerIDs, id)
	}

	return volunteerIDs, nil
}
