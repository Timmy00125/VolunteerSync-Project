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
	repo             repositories.VolunteerRepository
	geocodingService GeocodingService // Optional, can be nil
	logger           *logger.Logger
	// TODO: Add dependencies for fetching registrations, hours, achievements when those modules are ready
}

// NewVolunteerService creates a new instance of VolunteerService
func NewVolunteerService(
	repo repositories.VolunteerRepository,
	geocodingService GeocodingService,
	logger *logger.Logger,
) VolunteerService {
	return &volunteerService{
		repo:             repo,
		geocodingService: geocodingService,
		logger:           logger,
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

	// TODO: Fetch registrations, hours, achievements from respective modules when available
	// For now, return basic profile data with placeholder values
	dashboard := &DashboardResponse{
		Profile:            profile,
		TotalHours:         profile.TotalHours,
		TotalEvents:        profile.TotalEvents,
		TotalOrganizations: 0, // Will be calculated from registrations
		RecentEvents:       []RecentEvent{},
		UpcomingEvents:     []UpcomingEvent{},
		Achievements:       []Achievement{},
		HoursThisMonth:     0, // Will be calculated from hours logs
		EventsThisMonth:    0, // Will be calculated from registrations
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

	// TODO: Fetch analytics data from registrations and hours modules when available
	// For now, return basic metrics
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

	// TODO: Implement PDF generation using a library like gopdf or wkhtmltopdf
	// For now, return a placeholder
	s.logger.WithFields(map[string]interface{}{
		"user_id":    userID.String(),
		"profile_id": profile.ID.String(),
	}).Info("Impact report generation requested")

	// Return placeholder PDF content
	placeholder := []byte("PDF Impact Report Placeholder - To be implemented with PDF generation library")

	return placeholder, nil
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
