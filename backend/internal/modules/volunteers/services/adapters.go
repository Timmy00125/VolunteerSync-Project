package services

import (
	"context"

	achievementServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/achievements/services"
	hoursServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/services"
	opportunitiesModels "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/models"
	opportunitiesServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/services"
	organizationsServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/services"
	registrationModels "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/models"
	registrationServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/services"
	"github.com/google/uuid"
)

// registrationServiceAdapter adapts the registrations module service to the RegistrationServiceAdapter interface
type registrationServiceAdapter struct {
	registrationService registrationServices.RegistrationService
	opportunityService  opportunitiesServices.OpportunityService
	organizationService organizationsServices.OrganizationService
}

// NewRegistrationServiceAdapter creates a new registration service adapter
func NewRegistrationServiceAdapter(
	registrationService registrationServices.RegistrationService,
	opportunityService opportunitiesServices.OpportunityService,
	organizationService organizationsServices.OrganizationService,
) RegistrationServiceAdapter {
	return &registrationServiceAdapter{
		registrationService: registrationService,
		opportunityService:  opportunityService,
		organizationService: organizationService,
	}
}

// ListVolunteerRegistrations implements RegistrationServiceAdapter
func (a *registrationServiceAdapter) ListVolunteerRegistrations(
	ctx context.Context,
	volunteerProfileID uuid.UUID,
	filters RegistrationFilters,
) ([]*RegistrationInfo, error) {
	// Convert filters to registration service filters
	regFilters := registrationServices.RegistrationFilters{
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}

	if filters.Status != nil {
		status := registrationModels.RegistrationStatus(*filters.Status)
		regFilters.Status = &status
	}
	if filters.StartDate != nil {
		regFilters.StartDate = filters.StartDate
	}
	if filters.EndDate != nil {
		regFilters.EndDate = filters.EndDate
	}

	// Fetch registrations from registration service
	registrations, err := a.registrationService.ListVolunteerRegistrations(ctx, volunteerProfileID, regFilters)
	if err != nil {
		return nil, err
	}

	// Convert to RegistrationInfo
	result := make([]*RegistrationInfo, 0, len(registrations))
	for _, reg := range registrations {
		// Fetch opportunity details to get title, date, etc.
		opportunity, err := a.opportunityService.GetOpportunity(ctx, reg.OpportunityID)
		if err != nil {
			// Skip this registration if we can't get opportunity details
			continue
		}

		// Fetch organization details to get name
		organization, err := a.organizationService.GetOrganization(ctx, opportunity.OrganizationID)
		if err != nil {
			// Skip this registration if we can't get organization details
			continue
		}

		info := &RegistrationInfo{
			ID:               reg.ID,
			OpportunityID:    reg.OpportunityID,
			OpportunityTitle: opportunity.Title,
			OrganizationID:   opportunity.OrganizationID,
			OrganizationName: organization.Name,
			Status:           string(reg.Status),
			Date:             opportunity.StartDate, // Use start date as the event date
			StartTime:        opportunity.StartTime,
			EndTime:          opportunity.EndTime,
			Location:         getOpportunityLocation(opportunity),
			CauseCategory:    "", // TODO: Fetch from opportunity_causes junction table if needed
			RegisteredAt:     reg.RegisteredAt,
			CheckedInAt:      reg.CheckedInAt,
			HoursWorked:      reg.HoursWorked,
			HoursLoggedAt:    reg.HoursLoggedAt,
			HoursVerifiedAt:  reg.HoursVerifiedAt,
		}

		if reg.HoursStatus != nil {
			status := string(*reg.HoursStatus)
			info.HoursStatus = &status
		}

		result = append(result, info)
	}

	return result, nil
}

// getOpportunityLocation formats the opportunity location from address fields
func getOpportunityLocation(opp *opportunitiesModels.Opportunity) string {
	// Build a simple location string from address fields
	if opp.City != "" && opp.State != "" {
		return opp.City + ", " + opp.State
	}
	if opp.City != "" {
		return opp.City
	}
	return opp.AddressLine1
}

// hoursServiceAdapter adapts the hours module service to the HoursServiceAdapter interface
type hoursServiceAdapter struct {
	hoursService hoursServices.HoursService
}

// NewHoursServiceAdapter creates a new hours service adapter
func NewHoursServiceAdapter(hoursService hoursServices.HoursService) HoursServiceAdapter {
	return &hoursServiceAdapter{
		hoursService: hoursService,
	}
}

// GetHoursLogsByVolunteer implements HoursServiceAdapter
func (a *hoursServiceAdapter) GetHoursLogsByVolunteer(
	ctx context.Context,
	volunteerProfileID uuid.UUID,
) ([]*HoursLogInfo, error) {
	// Fetch hours logs from hours service
	hoursLogs, err := a.hoursService.GetHoursLogsByVolunteer(ctx, volunteerProfileID)
	if err != nil {
		return nil, err
	}

	// Convert to HoursLogInfo
	result := make([]*HoursLogInfo, 0, len(hoursLogs))
	for _, log := range hoursLogs {
		info := &HoursLogInfo{
			ID:               log.ID,
			RegistrationID:   log.RegistrationID,
			Hours:            log.Hours,
			Status:           string(log.Status),
			LoggedAt:         log.LoggedAt,
			VerifiedAt:       log.VerifiedAt,
			CoordinatorNotes: log.CoordinatorNotes,
			VolunteerNotes:   log.VolunteerNotes,
		}
		result = append(result, info)
	}

	return result, nil
}

// GetPendingHoursForVolunteer implements HoursServiceAdapter
func (a *hoursServiceAdapter) GetPendingHoursForVolunteer(
	ctx context.Context,
	volunteerProfileID uuid.UUID,
) ([]*HoursLogInfo, error) {
	// Fetch pending hours from hours service
	hoursLogs, err := a.hoursService.GetPendingHoursForVolunteer(ctx, volunteerProfileID)
	if err != nil {
		return nil, err
	}

	// Convert to HoursLogInfo
	result := make([]*HoursLogInfo, 0, len(hoursLogs))
	for _, log := range hoursLogs {
		info := &HoursLogInfo{
			ID:               log.ID,
			RegistrationID:   log.RegistrationID,
			Hours:            log.Hours,
			Status:           string(log.Status),
			LoggedAt:         log.LoggedAt,
			VerifiedAt:       log.VerifiedAt,
			CoordinatorNotes: log.CoordinatorNotes,
			VolunteerNotes:   log.VolunteerNotes,
		}
		result = append(result, info)
	}

	return result, nil
}

// achievementServiceAdapter adapts the achievements module service to the AchievementServiceAdapter interface
type achievementServiceAdapter struct {
	achievementService achievementServices.AchievementService
}

// NewAchievementServiceAdapter creates a new achievement service adapter
func NewAchievementServiceAdapter(achievementService achievementServices.AchievementService) AchievementServiceAdapter {
	return &achievementServiceAdapter{
		achievementService: achievementService,
	}
}

// GetVolunteerAchievements implements AchievementServiceAdapter
func (a *achievementServiceAdapter) GetVolunteerAchievements(
	ctx context.Context,
	volunteerProfileID uuid.UUID,
) ([]*AchievementInfo, error) {
	// Fetch achievements from achievement service
	achievements, err := a.achievementService.GetVolunteerAchievements(ctx, volunteerProfileID)
	if err != nil {
		return nil, err
	}

	// Convert to AchievementInfo
	result := make([]*AchievementInfo, 0, len(achievements))
	for _, ach := range achievements {
		info := &AchievementInfo{
			ID:          ach.AchievementID,
			Name:        ach.Achievement.Name,
			Description: ach.Achievement.Description,
			IconURL:     ach.Achievement.IconURL,
			EarnedAt:    ach.EarnedAt,
		}
		result = append(result, info)
	}

	return result, nil
}

// CountVolunteerAchievements implements AchievementServiceAdapter
func (a *achievementServiceAdapter) CountVolunteerAchievements(
	ctx context.Context,
	volunteerProfileID uuid.UUID,
) (int64, error) {
	return a.achievementService.CountVolunteerAchievements(ctx, volunteerProfileID)
}
