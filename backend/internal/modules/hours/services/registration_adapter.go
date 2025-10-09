package services

import (
	"context"

	regServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/services"
	"github.com/google/uuid"
)

// registrationServiceAdapter adapts the registrations module's service to satisfy
// the RegistrationService interface expected by the hours service.
// This maintains module boundaries while enabling cross-module communication.
type registrationServiceAdapter struct {
	regService regServices.RegistrationService
}

// NewRegistrationServiceAdapter creates a new adapter for the registration service
func NewRegistrationServiceAdapter(regService regServices.RegistrationService) RegistrationService {
	return &registrationServiceAdapter{
		regService: regService,
	}
}

// GetRegistration retrieves registration details needed for hours validation
func (a *registrationServiceAdapter) GetRegistration(ctx context.Context, registrationID uuid.UUID) (*RegistrationDetails, error) {
	// Get the full registration from the registrations module
	reg, err := a.regService.GetRegistration(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	// Extract and transform only the data needed by hours service
	details := &RegistrationDetails{
		ID:                 reg.ID,
		OpportunityID:      reg.OpportunityID,
		VolunteerProfileID: reg.VolunteerProfileID,
		Status:             string(reg.Status),
		CheckedInAt:        reg.CheckedInAt,
	}

	return details, nil
}

// UpdateRegistrationHours updates the hours information on a registration record
// This is called after hours are logged to keep registration data in sync
func (a *registrationServiceAdapter) UpdateRegistrationHours(ctx context.Context, registrationID uuid.UUID, hours float64, status string) error {
	// Note: This would require adding a method to the registration service
	// For now, we'll acknowledge this limitation in the adapter
	// The registration service would need to implement:
	// UpdateHoursInformation(ctx context.Context, registrationID uuid.UUID, hours float64, hoursStatus string) error

	// TODO: Once registration service implements UpdateHoursInformation, call it here
	// For now, this is a no-op to avoid breaking the build
	// The hours are still tracked correctly in the hours_logs table

	return nil // Placeholder until registration service adds the method
}
