package services

import (
	"context"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/models"
	oppServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/services"
	"github.com/google/uuid"
)

// opportunityServiceAdapter adapts the opportunities module's service to satisfy
// the OpportunityService interface expected by the registration service.
// This maintains module boundaries while enabling cross-module communication.
type opportunityServiceAdapter struct {
	oppService oppServices.OpportunityService
}

// NewOpportunityServiceAdapter creates a new adapter for the opportunity service
func NewOpportunityServiceAdapter(oppService oppServices.OpportunityService) OpportunityService {
	return &opportunityServiceAdapter{
		oppService: oppService,
	}
}

// GetOpportunity retrieves opportunity details needed for registration validation
func (a *opportunityServiceAdapter) GetOpportunity(ctx context.Context, id uuid.UUID) (*OpportunityDetails, error) {
	// Get the full opportunity from the opportunities module
	opp, err := a.oppService.GetOpportunity(ctx, id)
	if err != nil {
		return nil, err
	}

	// Extract and transform only the data needed by registration service
	details := &OpportunityDetails{
		ID:             opp.ID,
		OrganizationID: opp.OrganizationID,
		Title:          opp.Title,
		StartDate:      opp.StartDate,
		StartTime:      opp.StartTime,
		EndDate:        opp.EndDate,
		EndTime:        opp.EndTime,
		Capacity:       opp.Capacity,
		Status:         string(opp.Status),
		Location:       buildLocationString(opp),
		Timezone:       opp.Timezone,
	}

	// Note: CurrentRegistrations needs to be calculated by the registration service
	// using its own repository to avoid circular dependencies
	details.CurrentRegistrations = 0

	return details, nil
}

// buildLocationString constructs a readable location string from opportunity address
func buildLocationString(opp *models.Opportunity) string {
	location := opp.AddressLine1
	if opp.AddressLine2 != nil && *opp.AddressLine2 != "" {
		location += ", " + *opp.AddressLine2
	}
	location += ", " + opp.City + ", " + opp.State + " " + opp.PostalCode
	return location
}
