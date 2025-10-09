package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/volunteers/repositories"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/pdf"
	"github.com/google/uuid"
)

// Custom errors for service operations
var (
	// ErrInvalidProfileData is returned when profile data is invalid
	ErrInvalidProfileData = errors.New("invalid profile data")
	// ErrVolunteerProfileNotFound is returned when a volunteer profile cannot be found
	ErrVolunteerProfileNotFound = repositories.ErrVolunteerProfileNotFound
	// ErrVolunteerProfileAlreadyExists is returned when a profile already exists
	ErrVolunteerProfileAlreadyExists = repositories.ErrVolunteerProfileAlreadyExists
	// ErrUnauthorized is returned when a user doesn't have permission for an operation
	ErrUnauthorized = errors.New("unauthorized to perform this action")
)

// GeocodingService defines the interface for geocoding addresses to coordinates
// This will be implemented in the geocoding package (Task T133)
type GeocodingService interface {
	GeocodeAddress(ctx context.Context, address string) (lat, lng float64, err error)
}

// RegistrationServiceAdapter defines the minimal interface needed from registrations module
// This adapter pattern avoids circular dependencies between modules
type RegistrationServiceAdapter interface {
	// ListVolunteerRegistrations returns all registrations for a volunteer with filters
	ListVolunteerRegistrations(ctx context.Context, volunteerProfileID uuid.UUID, filters RegistrationFilters) ([]*RegistrationInfo, error)
}

// RegistrationFilters represents filters for listing registrations
type RegistrationFilters struct {
	Status    *string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// RegistrationInfo contains registration details needed for dashboard/analytics
type RegistrationInfo struct {
	ID               uuid.UUID
	OpportunityID    uuid.UUID
	OpportunityTitle string
	OrganizationID   uuid.UUID
	OrganizationName string
	Status           string
	Date             time.Time
	StartTime        time.Time
	EndTime          time.Time
	Location         string
	CauseCategory    string
	RegisteredAt     time.Time
	CheckedInAt      *time.Time
	HoursWorked      *float64
	HoursStatus      *string
	HoursLoggedAt    *time.Time
	HoursVerifiedAt  *time.Time
}

// HoursServiceAdapter defines the minimal interface needed from hours module
type HoursServiceAdapter interface {
	// GetHoursLogsByVolunteer returns all hours logs for a volunteer
	GetHoursLogsByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*HoursLogInfo, error)
	// GetPendingHoursForVolunteer returns pending hours logs awaiting verification
	GetPendingHoursForVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*HoursLogInfo, error)
}

// HoursLogInfo contains hours log details needed for dashboard/analytics
type HoursLogInfo struct {
	ID               uuid.UUID
	RegistrationID   uuid.UUID
	Hours            float64
	Status           string
	LoggedAt         time.Time
	VerifiedAt       *time.Time
	CoordinatorNotes *string
	VolunteerNotes   *string
}

// AchievementServiceAdapter defines the minimal interface needed from achievements module
type AchievementServiceAdapter interface {
	// GetVolunteerAchievements returns all achievements earned by a volunteer
	GetVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) ([]*AchievementInfo, error)
	// CountVolunteerAchievements returns total count of achievements earned
	CountVolunteerAchievements(ctx context.Context, volunteerProfileID uuid.UUID) (int64, error)
}

// AchievementInfo contains achievement details needed for dashboard
type AchievementInfo struct {
	ID          uuid.UUID
	Name        string
	Description string
	IconURL     *string
	EarnedAt    time.Time
}

// VolunteerService encapsulates volunteer profile business logic
// Provides methods for profile management, dashboard metrics, analytics, and impact reporting
type VolunteerService interface {
	// GetVolunteerProfile retrieves a volunteer profile by user ID
	GetVolunteerProfile(ctx context.Context, userID uuid.UUID) (*models.VolunteerProfile, error)

	// UpdateVolunteerProfile updates a volunteer profile with geocoding for location changes
	// Manages skills and interests associations
	UpdateVolunteerProfile(ctx context.Context, userID uuid.UUID, input UpdateVolunteerProfileInput) (*models.VolunteerProfile, error)

	// GetDashboard retrieves dashboard metrics for a volunteer
	// Includes total hours, events, organizations, and recent activity
	GetDashboard(ctx context.Context, userID uuid.UUID) (*DashboardResponse, error)

	// GetAnalytics retrieves analytics data for a volunteer
	// Includes hours over time, events by cause category, etc.
	GetAnalytics(ctx context.Context, userID uuid.UUID, dateRange DateRange) (*AnalyticsResponse, error)

	// GenerateImpactReport generates a PDF impact report for a volunteer
	// Returns the PDF file content as bytes
	GenerateImpactReport(ctx context.Context, userID uuid.UUID) ([]byte, error)

	// CreateVolunteerProfile creates a new volunteer profile (called during registration)
	CreateVolunteerProfile(ctx context.Context, userID uuid.UUID) (*models.VolunteerProfile, error)
}

// UpdateVolunteerProfileInput represents the input for updating a volunteer profile
type UpdateVolunteerProfileInput struct {
	ProfilePhotoURL          *string
	Biography                *string
	Location                 *string
	AvailabilityMonday       *bool
	AvailabilityTuesday      *bool
	AvailabilityWednesday    *bool
	AvailabilityThursday     *bool
	AvailabilityFriday       *bool
	AvailabilitySaturday     *bool
	AvailabilitySunday       *bool
	PreferredTime            *models.PreferredTime
	EmergencyContactName     *string
	EmergencyContactPhone    *string
	PrivacyShowHours         *bool
	PrivacyShowEvents        *bool
	PrivacyShowOrganizations *bool
	NotificationInApp        *bool
	NotificationBrowserPush  *bool
	SkillIDs                 []uuid.UUID // Skills to set (replaces existing)
	InterestIDs              []uuid.UUID // Interests to set (replaces existing)
}

// DashboardResponse represents the dashboard data for a volunteer
type DashboardResponse struct {
	Profile            *models.VolunteerProfile `json:"profile"`
	TotalHours         float64                  `json:"total_hours"`
	TotalEvents        int                      `json:"total_events"`
	TotalOrganizations int                      `json:"total_organizations"`
	RecentEvents       []RecentEvent            `json:"recent_events"`
	UpcomingEvents     []UpcomingEvent          `json:"upcoming_events"`
	Achievements       []Achievement            `json:"achievements"`
	HoursThisMonth     float64                  `json:"hours_this_month"`
	EventsThisMonth    int                      `json:"events_this_month"`
}

// RecentEvent represents a recently completed event
type RecentEvent struct {
	ID               uuid.UUID `json:"id"`
	OpportunityTitle string    `json:"opportunity_title"`
	OrganizationName string    `json:"organization_name"`
	Date             time.Time `json:"date"`
	HoursLogged      float64   `json:"hours_logged"`
	Status           string    `json:"status"`
}

// UpcomingEvent represents an upcoming registered event
type UpcomingEvent struct {
	ID               uuid.UUID `json:"id"`
	OpportunityTitle string    `json:"opportunity_title"`
	OrganizationName string    `json:"organization_name"`
	Date             time.Time `json:"date"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Location         string    `json:"location"`
	Status           string    `json:"status"`
}

// Achievement represents a volunteer achievement/badge
type Achievement struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	EarnedAt    time.Time `json:"earned_at"`
}

// DateRange represents a date range for analytics queries
type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

// AnalyticsResponse represents analytics data for a volunteer
type AnalyticsResponse struct {
	HoursOverTime        []DataPoint        `json:"hours_over_time"`
	EventsByCause        []CategoryCount    `json:"events_by_cause"`
	HoursByCause         []CategoryCount    `json:"hours_by_cause"`
	OrganizationStats    []OrganizationStat `json:"organization_stats"`
	TotalHours           float64            `json:"total_hours"`
	TotalEvents          int                `json:"total_events"`
	AverageHoursPerEvent float64            `json:"average_hours_per_event"`
}

// DataPoint represents a data point for time-series charts
type DataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// CategoryCount represents a count by category (e.g., events by cause)
type CategoryCount struct {
	Category string  `json:"category"`
	Count    int     `json:"count"`
	Hours    float64 `json:"hours,omitempty"`
}

// OrganizationStat represents statistics for a specific organization
type OrganizationStat struct {
	OrganizationID   uuid.UUID `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	TotalHours       float64   `json:"total_hours"`
	TotalEvents      int       `json:"total_events"`
}

// volunteerService is the concrete implementation of VolunteerService
type volunteerService struct {
	repo                repositories.VolunteerRepository
	geocodingService    GeocodingService // Optional, can be nil
	registrationService RegistrationServiceAdapter
	hoursService        HoursServiceAdapter
	achievementService  AchievementServiceAdapter
	logger              *logger.Logger
}

// NewVolunteerService creates a new instance of VolunteerService
func NewVolunteerService(
	repo repositories.VolunteerRepository,
	geocodingService GeocodingService,
	registrationService RegistrationServiceAdapter,
	hoursService HoursServiceAdapter,
	achievementService AchievementServiceAdapter,
	logger *logger.Logger,
) VolunteerService {
	return &volunteerService{
		repo:                repo,
		geocodingService:    geocodingService,
		registrationService: registrationService,
		hoursService:        hoursService,
		achievementService:  achievementService,
		logger:              logger,
	}
}

// GetVolunteerProfile retrieves a volunteer profile by user ID
func (s *volunteerService) GetVolunteerProfile(ctx context.Context, userID uuid.UUID) (*models.VolunteerProfile, error) {
	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("user ID is required", nil)
	}

	profile, err := s.repo.FindVolunteerProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrVolunteerProfileNotFound) {
			return nil, ErrVolunteerProfileNotFound
		}
		s.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		}).Error("Failed to get volunteer profile")
		return nil, fmt.Errorf("failed to get volunteer profile: %w", err)
	}

	return profile, nil
}

// UpdateVolunteerProfile updates a volunteer profile with geocoding for location changes
func (s *volunteerService) UpdateVolunteerProfile(ctx context.Context, userID uuid.UUID, input UpdateVolunteerProfileInput) (*models.VolunteerProfile, error) {
	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("user ID is required", nil)
	}

	// Get existing profile
	profile, err := s.repo.FindVolunteerProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrVolunteerProfileNotFound) {
			return nil, ErrVolunteerProfileNotFound
		}
		s.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		}).Error("Failed to find volunteer profile")
		return nil, fmt.Errorf("failed to find volunteer profile: %w", err)
	}

	// Update fields if provided
	if input.ProfilePhotoURL != nil {
		profile.ProfilePhotoURL = input.ProfilePhotoURL
	}
	if input.Biography != nil {
		profile.Biography = input.Biography
	}
	if input.Location != nil {
		profile.Location = input.Location

		// Geocode the new location if geocoding service is available
		if s.geocodingService != nil && *input.Location != "" {
			lat, lng, err := s.geocodingService.GeocodeAddress(ctx, *input.Location)
			if err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error":    err.Error(),
					"location": *input.Location,
				}).Warn("Failed to geocode location")
				// Don't fail the update if geocoding fails
			} else {
				profile.Latitude = &lat
				profile.Longitude = &lng
			}
		}
	}

	// Update availability fields
	if input.AvailabilityMonday != nil {
		profile.AvailabilityMonday = *input.AvailabilityMonday
	}
	if input.AvailabilityTuesday != nil {
		profile.AvailabilityTuesday = *input.AvailabilityTuesday
	}
	if input.AvailabilityWednesday != nil {
		profile.AvailabilityWednesday = *input.AvailabilityWednesday
	}
	if input.AvailabilityThursday != nil {
		profile.AvailabilityThursday = *input.AvailabilityThursday
	}
	if input.AvailabilityFriday != nil {
		profile.AvailabilityFriday = *input.AvailabilityFriday
	}
	if input.AvailabilitySaturday != nil {
		profile.AvailabilitySaturday = *input.AvailabilitySaturday
	}
	if input.AvailabilitySunday != nil {
		profile.AvailabilitySunday = *input.AvailabilitySunday
	}

	if input.PreferredTime != nil {
		profile.PreferredTime = input.PreferredTime
	}
	if input.EmergencyContactName != nil {
		profile.EmergencyContactName = input.EmergencyContactName
	}
	if input.EmergencyContactPhone != nil {
		profile.EmergencyContactPhone = input.EmergencyContactPhone
	}

	// Update privacy settings
	if input.PrivacyShowHours != nil {
		profile.PrivacyShowHours = *input.PrivacyShowHours
	}
	if input.PrivacyShowEvents != nil {
		profile.PrivacyShowEvents = *input.PrivacyShowEvents
	}
	if input.PrivacyShowOrganizations != nil {
		profile.PrivacyShowOrganizations = *input.PrivacyShowOrganizations
	}

	// Update notification settings
	if input.NotificationInApp != nil {
		profile.NotificationInApp = *input.NotificationInApp
	}
	if input.NotificationBrowserPush != nil {
		profile.NotificationBrowserPush = *input.NotificationBrowserPush
	}

	// Update the profile in the database
	if err := s.repo.UpdateVolunteerProfile(ctx, profile); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"profile_id": profile.ID.String(),
		}).Error("Failed to update volunteer profile")
		return nil, fmt.Errorf("failed to update volunteer profile: %w", err)
	}

	// Update skills if provided (replace existing)
	if input.SkillIDs != nil {
		// Get existing skills
		existingSkills, err := s.repo.GetSkills(ctx, profile.ID)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      err.Error(),
				"profile_id": profile.ID.String(),
			}).Error("Failed to get existing skills")
			return nil, fmt.Errorf("failed to get existing skills: %w", err)
		}

		// Remove existing skills
		if len(existingSkills) > 0 {
			if err := s.repo.RemoveSkills(ctx, profile.ID, existingSkills); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error":      err.Error(),
					"profile_id": profile.ID.String(),
				}).Error("Failed to remove existing skills")
				return nil, fmt.Errorf("failed to remove existing skills: %w", err)
			}
		}

		// Add new skills
		if len(input.SkillIDs) > 0 {
			if err := s.repo.AddSkills(ctx, profile.ID, input.SkillIDs); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error":      err.Error(),
					"profile_id": profile.ID.String(),
				}).Error("Failed to add skills")
				return nil, fmt.Errorf("failed to add skills: %w", err)
			}
		}
	}

	// Update interests if provided (replace existing)
	if input.InterestIDs != nil {
		// Get existing interests
		existingInterests, err := s.repo.GetInterests(ctx, profile.ID)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      err.Error(),
				"profile_id": profile.ID.String(),
			}).Error("Failed to get existing interests")
			return nil, fmt.Errorf("failed to get existing interests: %w", err)
		}

		// Remove existing interests
		if len(existingInterests) > 0 {
			if err := s.repo.RemoveInterests(ctx, profile.ID, existingInterests); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error":      err.Error(),
					"profile_id": profile.ID.String(),
				}).Error("Failed to remove existing interests")
				return nil, fmt.Errorf("failed to remove existing interests: %w", err)
			}
		}

		// Add new interests
		if len(input.InterestIDs) > 0 {
			if err := s.repo.AddInterests(ctx, profile.ID, input.InterestIDs); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error":      err.Error(),
					"profile_id": profile.ID.String(),
				}).Error("Failed to add interests")
				return nil, fmt.Errorf("failed to add interests: %w", err)
			}
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"profile_id": profile.ID.String(),
		"user_id":    userID.String(),
	}).Info("Volunteer profile updated successfully")

	return profile, nil
}

// GetDashboard retrieves dashboard metrics for a volunteer
func (s *volunteerService) GetDashboard(ctx context.Context, userID uuid.UUID) (*DashboardResponse, error) {
	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("user ID is required", nil)
	}

	// Get volunteer profile
	profile, err := s.repo.FindVolunteerProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrVolunteerProfileNotFound) {
			return nil, ErrVolunteerProfileNotFound
		}
		s.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		}).Error("Failed to find volunteer profile")
		return nil, fmt.Errorf("failed to find volunteer profile: %w", err)
	}

	// Initialize dashboard response
	dashboard := &DashboardResponse{
		Profile:            profile,
		TotalHours:         profile.TotalHours,
		TotalEvents:        profile.TotalEvents,
		TotalOrganizations: 0,
		RecentEvents:       []RecentEvent{},
		UpcomingEvents:     []UpcomingEvent{},
		Achievements:       []Achievement{},
		HoursThisMonth:     0,
		EventsThisMonth:    0,
	}

	// Fetch all registrations for the volunteer
	if s.registrationService != nil {
		registrations, err := s.registrationService.ListVolunteerRegistrations(ctx, profile.ID, RegistrationFilters{
			Limit: 1000, // Get all registrations
		})
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      err.Error(),
				"profile_id": profile.ID.String(),
			}).Warn("Failed to fetch registrations for dashboard")
			// Continue with partial data
		} else {
			// Calculate total organizations (unique organizations)
			orgMap := make(map[uuid.UUID]bool)
			now := time.Now()
			currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

			for _, reg := range registrations {
				orgMap[reg.OrganizationID] = true

				// Count events this month
				if reg.Date.After(currentMonthStart) || reg.Date.Equal(currentMonthStart) {
					dashboard.EventsThisMonth++
				}

				// Build recent events (completed, last 5)
				if reg.Status == "completed" && reg.Date.Before(now) {
					dashboard.RecentEvents = append(dashboard.RecentEvents, RecentEvent{
						ID:               reg.ID,
						OpportunityTitle: reg.OpportunityTitle,
						OrganizationName: reg.OrganizationName,
						Date:             reg.Date,
						HoursLogged:      getFloatValue(reg.HoursWorked),
						Status:           getStringValue(reg.HoursStatus),
					})
				}

				// Build upcoming events (confirmed, future, next 5)
				if reg.Status == "confirmed" && reg.Date.After(now) {
					dashboard.UpcomingEvents = append(dashboard.UpcomingEvents, UpcomingEvent{
						ID:               reg.ID,
						OpportunityTitle: reg.OpportunityTitle,
						OrganizationName: reg.OrganizationName,
						Date:             reg.Date,
						StartTime:        reg.StartTime,
						EndTime:          reg.EndTime,
						Location:         reg.Location,
						Status:           reg.Status,
					})
				}
			}

			dashboard.TotalOrganizations = len(orgMap)

			// Sort and limit recent events (most recent first)
			if len(dashboard.RecentEvents) > 1 {
				// Sort by date descending
				for i := 0; i < len(dashboard.RecentEvents)-1; i++ {
					for j := i + 1; j < len(dashboard.RecentEvents); j++ {
						if dashboard.RecentEvents[i].Date.Before(dashboard.RecentEvents[j].Date) {
							dashboard.RecentEvents[i], dashboard.RecentEvents[j] = dashboard.RecentEvents[j], dashboard.RecentEvents[i]
						}
					}
				}
			}
			if len(dashboard.RecentEvents) > 5 {
				dashboard.RecentEvents = dashboard.RecentEvents[:5]
			}

			// Sort and limit upcoming events (soonest first)
			if len(dashboard.UpcomingEvents) > 1 {
				// Sort by date ascending
				for i := 0; i < len(dashboard.UpcomingEvents)-1; i++ {
					for j := i + 1; j < len(dashboard.UpcomingEvents); j++ {
						if dashboard.UpcomingEvents[i].Date.After(dashboard.UpcomingEvents[j].Date) {
							dashboard.UpcomingEvents[i], dashboard.UpcomingEvents[j] = dashboard.UpcomingEvents[j], dashboard.UpcomingEvents[i]
						}
					}
				}
			}
			if len(dashboard.UpcomingEvents) > 5 {
				dashboard.UpcomingEvents = dashboard.UpcomingEvents[:5]
			}
		}
	}

	// Fetch hours logs to calculate hours this month
	if s.hoursService != nil {
		hoursLogs, err := s.hoursService.GetHoursLogsByVolunteer(ctx, profile.ID)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      err.Error(),
				"profile_id": profile.ID.String(),
			}).Warn("Failed to fetch hours logs for dashboard")
			// Continue with partial data
		} else {
			now := time.Now()
			currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

			for _, log := range hoursLogs {
				// Only count verified hours
				if log.Status == "verified" && log.LoggedAt.After(currentMonthStart) || log.LoggedAt.Equal(currentMonthStart) {
					dashboard.HoursThisMonth += log.Hours
				}
			}
		}
	}

	// Fetch achievements
	if s.achievementService != nil {
		achievements, err := s.achievementService.GetVolunteerAchievements(ctx, profile.ID)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      err.Error(),
				"profile_id": profile.ID.String(),
			}).Warn("Failed to fetch achievements for dashboard")
			// Continue with partial data
		} else {
			for _, ach := range achievements {
				dashboard.Achievements = append(dashboard.Achievements, Achievement{
					ID:          ach.ID,
					Name:        ach.Name,
					Description: ach.Description,
					IconURL:     getStringValue(ach.IconURL),
					EarnedAt:    ach.EarnedAt,
				})
			}
		}
	}

	s.logger.WithField("user_id", userID.String()).Info("Dashboard retrieved successfully")

	return dashboard, nil
}

// GetAnalytics retrieves analytics data for a volunteer
func (s *volunteerService) GetAnalytics(ctx context.Context, userID uuid.UUID, dateRange DateRange) (*AnalyticsResponse, error) {
	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("user ID is required", nil)
	}

	// Get volunteer profile
	profile, err := s.repo.FindVolunteerProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrVolunteerProfileNotFound) {
			return nil, ErrVolunteerProfileNotFound
		}
		s.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		}).Error("Failed to find volunteer profile")
		return nil, fmt.Errorf("failed to find volunteer profile: %w", err)
	}

	// Initialize analytics response
	analytics := &AnalyticsResponse{
		HoursOverTime:        []DataPoint{},
		EventsByCause:        []CategoryCount{},
		HoursByCause:         []CategoryCount{},
		OrganizationStats:    []OrganizationStat{},
		TotalHours:           profile.TotalHours,
		TotalEvents:          profile.TotalEvents,
		AverageHoursPerEvent: 0,
	}

	if profile.TotalEvents > 0 {
		analytics.AverageHoursPerEvent = profile.TotalHours / float64(profile.TotalEvents)
	}

	// Fetch registrations for analytics
	if s.registrationService != nil {
		// Apply date range filter
		filters := RegistrationFilters{
			Limit: 1000,
		}
		if !dateRange.StartDate.IsZero() {
			filters.StartDate = &dateRange.StartDate
		}
		if !dateRange.EndDate.IsZero() {
			filters.EndDate = &dateRange.EndDate
		}

		registrations, err := s.registrationService.ListVolunteerRegistrations(ctx, profile.ID, filters)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      err.Error(),
				"profile_id": profile.ID.String(),
			}).Warn("Failed to fetch registrations for analytics")
			// Continue with partial data
		} else {
			// Build cause category counts and organization stats
			causeCountMap := make(map[string]int)
			causeHoursMap := make(map[string]float64)
			orgStatsMap := make(map[uuid.UUID]*OrganizationStat)

			for _, reg := range registrations {
				// Count events by cause
				if reg.CauseCategory != "" {
					causeCountMap[reg.CauseCategory]++
					if reg.HoursWorked != nil {
						causeHoursMap[reg.CauseCategory] += *reg.HoursWorked
					}
				}

				// Aggregate organization stats
				if _, exists := orgStatsMap[reg.OrganizationID]; !exists {
					orgStatsMap[reg.OrganizationID] = &OrganizationStat{
						OrganizationID:   reg.OrganizationID,
						OrganizationName: reg.OrganizationName,
						TotalHours:       0,
						TotalEvents:      0,
					}
				}
				orgStatsMap[reg.OrganizationID].TotalEvents++
				if reg.HoursWorked != nil {
					orgStatsMap[reg.OrganizationID].TotalHours += *reg.HoursWorked
				}
			}

			// Convert maps to slices
			for cause, count := range causeCountMap {
				analytics.EventsByCause = append(analytics.EventsByCause, CategoryCount{
					Category: cause,
					Count:    count,
				})
			}
			for cause, hours := range causeHoursMap {
				analytics.HoursByCause = append(analytics.HoursByCause, CategoryCount{
					Category: cause,
					Count:    0, // Not used for hours
					Hours:    hours,
				})
			}
			for _, stat := range orgStatsMap {
				analytics.OrganizationStats = append(analytics.OrganizationStats, *stat)
			}
		}
	}

	// Fetch hours logs for time-series data
	if s.hoursService != nil {
		hoursLogs, err := s.hoursService.GetHoursLogsByVolunteer(ctx, profile.ID)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      err.Error(),
				"profile_id": profile.ID.String(),
			}).Warn("Failed to fetch hours logs for analytics")
			// Continue with partial data
		} else {
			// Build hours over time (grouped by month)
			hoursPerMonth := make(map[string]float64)

			for _, log := range hoursLogs {
				// Only include verified hours
				if log.Status != "verified" {
					continue
				}

				// Apply date range filter
				if !dateRange.StartDate.IsZero() && log.LoggedAt.Before(dateRange.StartDate) {
					continue
				}
				if !dateRange.EndDate.IsZero() && log.LoggedAt.After(dateRange.EndDate) {
					continue
				}

				// Group by month (YYYY-MM format)
				monthKey := log.LoggedAt.Format("2006-01")
				hoursPerMonth[monthKey] += log.Hours
			}

			// Convert to data points and sort
			for monthKey, hours := range hoursPerMonth {
				// Parse month key back to time
				monthTime, err := time.Parse("2006-01", monthKey)
				if err != nil {
					continue
				}
				analytics.HoursOverTime = append(analytics.HoursOverTime, DataPoint{
					Date:  monthTime,
					Value: hours,
				})
			}

			// Sort data points by date
			if len(analytics.HoursOverTime) > 1 {
				for i := 0; i < len(analytics.HoursOverTime)-1; i++ {
					for j := i + 1; j < len(analytics.HoursOverTime); j++ {
						if analytics.HoursOverTime[i].Date.After(analytics.HoursOverTime[j].Date) {
							analytics.HoursOverTime[i], analytics.HoursOverTime[j] = analytics.HoursOverTime[j], analytics.HoursOverTime[i]
						}
					}
				}
			}
		}
	}

	s.logger.WithField("user_id", userID.String()).Info("Analytics retrieved successfully")

	return analytics, nil
}

// GenerateImpactReport generates a PDF impact report for a volunteer
func (s *volunteerService) GenerateImpactReport(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("user ID is required", nil)
	}

	// Get volunteer profile
	profile, err := s.repo.FindVolunteerProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrVolunteerProfileNotFound) {
			return nil, ErrVolunteerProfileNotFound
		}
		s.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		}).Error("Failed to find volunteer profile")
		return nil, fmt.Errorf("failed to find volunteer profile: %w", err)
	}

	// Create PDF generator
	generator := pdf.NewGenerator()
	generator.SetTitle("Volunteer Impact Report")
	generator.SetAuthor("VolunteerSync Platform")

	// Add header
	generator.AddHeader(
		"Volunteer Impact Report",
		fmt.Sprintf("Profile ID: %s", profile.ID.String()),
	)

	// Add generation date footer
	generator.AddFooter(time.Now())

	// Profile Information Section
	generator.AddSectionTitle("Volunteer Profile")

	if profile.Biography != nil && *profile.Biography != "" {
		generator.AddText(*profile.Biography)
		generator.GetPDF().Ln(3)
	}

	if profile.Location != nil && *profile.Location != "" {
		generator.AddKeyValue("Location", *profile.Location)
	}

	if profile.EmergencyContactName != nil && *profile.EmergencyContactName != "" {
		generator.AddKeyValue("Emergency Contact", *profile.EmergencyContactName)
	}

	generator.AddDivider()

	// Impact Metrics Section
	generator.AddSectionTitle("Impact Summary")

	// Add metric boxes for key stats
	metrics := []pdf.Metric{
		{Label: "Total Hours", Value: fmt.Sprintf("%.1f", profile.TotalHours)},
		{Label: "Total Events", Value: fmt.Sprintf("%d", profile.TotalEvents)},
	}

	// Calculate average hours per event
	avgHours := 0.0
	if profile.TotalEvents > 0 {
		avgHours = profile.TotalHours / float64(profile.TotalEvents)
	}
	metrics = append(metrics, pdf.Metric{
		Label: "Avg Hours/Event",
		Value: fmt.Sprintf("%.1f", avgHours),
	})

	generator.AddMetricsRow(metrics)

	generator.AddDivider()

	// Availability Section
	generator.AddSectionTitle("Availability")

	availabilityDays := []struct {
		day   string
		avail bool
	}{
		{"Monday", profile.AvailabilityMonday},
		{"Tuesday", profile.AvailabilityTuesday},
		{"Wednesday", profile.AvailabilityWednesday},
		{"Thursday", profile.AvailabilityThursday},
		{"Friday", profile.AvailabilityFriday},
		{"Saturday", profile.AvailabilitySaturday},
		{"Sunday", profile.AvailabilitySunday},
	}

	availText := "Available on: "
	availCount := 0
	for _, day := range availabilityDays {
		if day.avail {
			if availCount > 0 {
				availText += ", "
			}
			availText += day.day
			availCount++
		}
	}

	if availCount == 0 {
		availText = "Availability not specified"
	}

	generator.AddText(availText)

	if profile.PreferredTime != nil && *profile.PreferredTime != "" {
		generator.GetPDF().Ln(2)
		generator.AddKeyValue("Preferred Time", string(*profile.PreferredTime))
	}

	generator.AddDivider()

	// Note about future enhancements
	generator.AddSectionTitle("About This Report")
	generator.AddText("This impact report showcases your volunteer contributions. Future enhancements will include detailed event history, organization breakdowns, skills and interests, achievement displays, and visual charts of your volunteer journey over time.")

	// Generate the PDF
	pdfBytes, err := generator.Output()
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		}).Error("Failed to generate PDF")
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":    userID.String(),
		"profile_id": profile.ID.String(),
		"pdf_size":   len(pdfBytes),
	}).Info("Impact report generated successfully")

	return pdfBytes, nil
}

// CreateVolunteerProfile creates a new volunteer profile during user registration
func (s *volunteerService) CreateVolunteerProfile(ctx context.Context, userID uuid.UUID) (*models.VolunteerProfile, error) {
	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("user ID is required", nil)
	}

	// Create new profile with default values
	profile := &models.VolunteerProfile{
		UserID:                   userID,
		TotalHours:               0,
		TotalEvents:              0,
		PrivacyShowHours:         true,
		PrivacyShowEvents:        true,
		PrivacyShowOrganizations: true,
		NotificationInApp:        true,
		NotificationBrowserPush:  false,
	}

	// Save to database
	if err := s.repo.CreateVolunteerProfile(ctx, profile); err != nil {
		if errors.Is(err, repositories.ErrVolunteerProfileAlreadyExists) {
			return nil, ErrVolunteerProfileAlreadyExists
		}
		s.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		}).Error("Failed to create volunteer profile")
		return nil, fmt.Errorf("failed to create volunteer profile: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"profile_id": profile.ID.String(),
		"user_id":    userID.String(),
	}).Info("Volunteer profile created successfully")

	return profile, nil
}

// Helper functions for handling pointer values

// getFloatValue safely extracts float value from pointer, returns 0 if nil
func getFloatValue(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// getStringValue safely extracts string value from pointer, returns empty string if nil
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
