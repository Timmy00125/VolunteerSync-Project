package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/repositories"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/google/uuid"
)

// Custom errors for service operations
var (
	// ErrOpportunityNotFound is returned when an opportunity cannot be found
	ErrOpportunityNotFound = repositories.ErrOpportunityNotFound
	// ErrInvalidOpportunityData is returned when opportunity data is invalid
	ErrInvalidOpportunityData = errors.New("invalid opportunity data")
	// ErrCannotEditPastEvent is returned when attempting to edit a past event
	ErrCannotEditPastEvent = errors.New("cannot edit past events except for hour logging")
	// ErrOpportunityFull is returned when attempting to register for a full opportunity
	ErrOpportunityFull = errors.New("opportunity is at capacity")
	// ErrUnauthorized is returned when a user doesn't have permission for an operation
	ErrUnauthorized = errors.New("unauthorized to perform this action")
)

// GeocodingService defines the interface for geocoding addresses to coordinates
type GeocodingService interface {
	GeocodeAddress(ctx context.Context, address string) (lat, lng float64, err error)
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	NotifyVolunteersOfCancellation(ctx context.Context, opportunityID uuid.UUID, reason string) error
}

// OpportunityService encapsulates opportunity business logic
type OpportunityService interface {
	// CreateOpportunity creates a new opportunity with geocoding
	CreateOpportunity(ctx context.Context, input CreateOpportunityInput, creatorUserID uuid.UUID) (*models.Opportunity, error)

	// GetOpportunity retrieves an opportunity by ID
	GetOpportunity(ctx context.Context, oppID uuid.UUID) (*models.Opportunity, error)

	// UpdateOpportunity updates an opportunity
	UpdateOpportunity(ctx context.Context, oppID uuid.UUID, input UpdateOpportunityInput, userID uuid.UUID) (*models.Opportunity, error)

	// ListOpportunities searches opportunities with complex filters
	ListOpportunities(ctx context.Context, filters OpportunityListFilters) (*OpportunityListResponse, error)

	// CancelOpportunity cancels an opportunity and notifies registered volunteers
	CancelOpportunity(ctx context.Context, oppID uuid.UUID, reason string, userID uuid.UUID) error

	// CompleteOpportunity marks an opportunity as completed
	CompleteOpportunity(ctx context.Context, oppID uuid.UUID, userID uuid.UUID) error

	// CreateRecurringOpportunities generates child instances for a recurring opportunity
	CreateRecurringOpportunities(ctx context.Context, oppID uuid.UUID) ([]models.Opportunity, error)

	// AutoCompleteOpportunities finds and auto-completes opportunities 7 days after end date
	AutoCompleteOpportunities(ctx context.Context) error
}

// CreateOpportunityInput represents the input for creating a new opportunity
type CreateOpportunityInput struct {
	OrganizationID     uuid.UUID
	Title              string
	Description        string
	PublishImmediately bool
	StartDate          time.Time
	StartTime          time.Time
	EndDate            time.Time
	EndTime            time.Time
	Timezone           string
	AddressLine1       string
	AddressLine2       *string
	City               string
	State              string
	PostalCode         string
	Country            string
	Capacity           int
	MinAge             *int
	IsRecurring        bool
	RecurrencePattern  *models.RecurrencePattern
	RecurrenceEndDate  *time.Time
	SkillIDs           []uuid.UUID
	CauseIDs           []uuid.UUID
	DocumentIDs        []uuid.UUID
}

// UpdateOpportunityInput represents the input for updating an opportunity
type UpdateOpportunityInput struct {
	Title        *string
	Description  *string
	Status       *models.OpportunityStatus
	StartDate    *time.Time
	StartTime    *time.Time
	EndDate      *time.Time
	EndTime      *time.Time
	Timezone     *string
	AddressLine1 *string
	AddressLine2 *string
	City         *string
	State        *string
	PostalCode   *string
	Country      *string
	Capacity     *int
	MinAge       *int
	SkillIDs     []uuid.UUID
	CauseIDs     []uuid.UUID
	DocumentIDs  []uuid.UUID
}

// OpportunityListFilters represents filters for listing opportunities
type OpportunityListFilters struct {
	Search         string
	OrganizationID *uuid.UUID
	Status         *models.OpportunityStatus
	City           string
	State          string
	Latitude       *float64
	Longitude      *float64
	RadiusKm       *float64
	StartDateFrom  *time.Time
	StartDateTo    *time.Time
	CauseIDs       []uuid.UUID
	SkillIDs       []uuid.UUID
	MinAge         *int
	OnlyRecurring  bool
	Page           int
	Limit          int
	SortBy         string
	SortOrder      string
}

// OpportunityListResponse represents a paginated list of opportunities
type OpportunityListResponse struct {
	Opportunities []models.Opportunity `json:"data"`
	Pagination    PaginationInfo       `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalPages int  `json:"total_pages"`
	TotalItems int  `json:"total_items"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// opportunityService is the concrete implementation
type opportunityService struct {
	repo                repositories.OpportunityRepository
	geocodingService    GeocodingService
	notificationService NotificationService
	logger              *logger.Logger
}

// NewOpportunityService creates a new instance of OpportunityService
func NewOpportunityService(
	repo repositories.OpportunityRepository,
	geocodingService GeocodingService,
	notificationService NotificationService,
	logger *logger.Logger,
) OpportunityService {
	return &opportunityService{
		repo:                repo,
		geocodingService:    geocodingService,
		notificationService: notificationService,
		logger:              logger,
	}
}

// CreateOpportunity creates a new opportunity with geocoding
func (s *opportunityService) CreateOpportunity(ctx context.Context, input CreateOpportunityInput, creatorUserID uuid.UUID) (*models.Opportunity, error) {
	// Validate input
	if input.Title == "" {
		return nil, apperrors.NewBadRequestError("title is required")
	}
	if input.Description == "" {
		return nil, apperrors.NewBadRequestError("description is required")
	}
	if input.Capacity <= 0 {
		return nil, apperrors.NewBadRequestError("capacity must be greater than 0")
	}
	if input.EndDate.Before(input.StartDate) {
		return nil, apperrors.NewBadRequestError("end date must be on or after start date")
	}

	// Set status
	status := models.OpportunityStatusDraft
	var publishedAt *time.Time
	if input.PublishImmediately {
		status = models.OpportunityStatusPublished
		now := time.Now()
		publishedAt = &now
	}

	// Create opportunity model
	opp := &models.Opportunity{
		OrganizationID:    input.OrganizationID,
		CreatedByUserID:   creatorUserID,
		Title:             input.Title,
		Description:       input.Description,
		Status:            status,
		StartDate:         input.StartDate,
		StartTime:         input.StartTime,
		EndDate:           input.EndDate,
		EndTime:           input.EndTime,
		Timezone:          input.Timezone,
		AddressLine1:      input.AddressLine1,
		AddressLine2:      input.AddressLine2,
		City:              input.City,
		State:             input.State,
		PostalCode:        input.PostalCode,
		Country:           input.Country,
		Capacity:          input.Capacity,
		MinAge:            input.MinAge,
		IsRecurring:       input.IsRecurring,
		RecurrencePattern: input.RecurrencePattern,
		RecurrenceEndDate: input.RecurrenceEndDate,
		PublishedAt:       publishedAt,
	}

	// Geocode address
	if s.geocodingService != nil {
		address := fmt.Sprintf("%s, %s, %s %s, %s",
			input.AddressLine1, input.City, input.State, input.PostalCode, input.Country)
		lat, lng, err := s.geocodingService.GeocodeAddress(ctx, address)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"address": address,
				"error":   err.Error(),
			}).Warn("Failed to geocode address")
		} else {
			opp.Latitude = &lat
			opp.Longitude = &lng
		}
	}

	// Validate
	if err := opp.Validate(); err != nil {
		return nil, apperrors.NewValidationError("invalid opportunity data", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}

	// Create in database
	if err := s.repo.CreateOpportunity(ctx, opp); err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to create opportunity")
		return nil, apperrors.NewInternalServerError("Failed to create opportunity")
	}

	// Create recurring instances
	if input.IsRecurring && input.PublishImmediately {
		_, err := s.repo.CreateRecurringInstances(ctx, opp)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"opportunity_id": opp.ID,
				"error":          err.Error(),
			}).Error("Failed to create recurring instances")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"opportunity_id": opp.ID,
		"status":         status,
	}).Info("Opportunity created successfully")

	return opp, nil
}

// GetOpportunity retrieves an opportunity by ID
func (s *opportunityService) GetOpportunity(ctx context.Context, oppID uuid.UUID) (*models.Opportunity, error) {
	opp, err := s.repo.FindOpportunityByID(ctx, oppID)
	if err != nil {
		if errors.Is(err, repositories.ErrOpportunityNotFound) {
			return nil, apperrors.NewNotFoundError("opportunity")
		}
		s.logger.WithFields(map[string]interface{}{
			"opportunity_id": oppID,
			"error":          err.Error(),
		}).Error("Failed to get opportunity")
		return nil, apperrors.NewInternalServerError("Failed to get opportunity")
	}

	return opp, nil
}

// UpdateOpportunity updates an opportunity
func (s *opportunityService) UpdateOpportunity(ctx context.Context, oppID uuid.UUID, input UpdateOpportunityInput, userID uuid.UUID) (*models.Opportunity, error) {
	opp, err := s.repo.FindOpportunityByID(ctx, oppID)
	if err != nil {
		if errors.Is(err, repositories.ErrOpportunityNotFound) {
			return nil, apperrors.NewNotFoundError("opportunity")
		}
		s.logger.WithField("error", err.Error()).Error("Failed to get opportunity for update")
		return nil, apperrors.NewInternalServerError("Failed to get opportunity")
	}

	// Check authorization
	if opp.CreatedByUserID != userID {
		s.logger.WithFields(map[string]interface{}{
			"opportunity_id": oppID,
			"user_id":        userID,
		}).Warn("Unauthorized update attempt")
		return nil, apperrors.NewForbiddenError("You don't have permission to update this opportunity")
	}

	// Prevent editing past events
	if !opp.CanEdit() {
		return nil, apperrors.NewBadRequestError("Cannot edit past events")
	}

	// Apply updates
	if input.Title != nil {
		opp.Title = *input.Title
	}
	if input.Description != nil {
		opp.Description = *input.Description
	}
	if input.Status != nil {
		opp.Status = *input.Status
	}
	if input.Capacity != nil {
		opp.Capacity = *input.Capacity
	}

	// Validate
	if err := opp.Validate(); err != nil {
		return nil, apperrors.NewValidationError("invalid opportunity data", map[string]interface{}{
			"validation_error": err.Error(),
		})
	}

	// Update in database
	if err := s.repo.UpdateOpportunity(ctx, opp); err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to update opportunity")
		return nil, apperrors.NewInternalServerError("Failed to update opportunity")
	}

	s.logger.WithField("opportunity_id", oppID).Info("Opportunity updated successfully")
	return opp, nil
}

// ListOpportunities searches opportunities with complex filters
func (s *opportunityService) ListOpportunities(ctx context.Context, filters OpportunityListFilters) (*OpportunityListResponse, error) {
	repoFilters := repositories.OpportunityFilters{
		Search:         filters.Search,
		OrganizationID: filters.OrganizationID,
		Status:         filters.Status,
		City:           filters.City,
		State:          filters.State,
		Latitude:       filters.Latitude,
		Longitude:      filters.Longitude,
		RadiusKm:       filters.RadiusKm,
		StartDateFrom:  filters.StartDateFrom,
		StartDateTo:    filters.StartDateTo,
		CauseIDs:       filters.CauseIDs,
		SkillIDs:       filters.SkillIDs,
		MinAge:         filters.MinAge,
		OnlyRecurring:  filters.OnlyRecurring,
		Page:           filters.Page,
		Limit:          filters.Limit,
		SortBy:         filters.SortBy,
		SortOrder:      filters.SortOrder,
	}

	startTime := time.Now()
	result, err := s.repo.ListOpportunities(ctx, repoFilters)
	duration := time.Since(startTime)

	// Log slow queries
	if duration > 2*time.Second {
		s.logger.WithField("duration_ms", duration.Milliseconds()).Warn("Slow opportunity search query")
	}

	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to list opportunities")
		return nil, apperrors.NewInternalServerError("Failed to list opportunities")
	}

	return &OpportunityListResponse{
		Opportunities: result.Opportunities,
		Pagination: PaginationInfo{
			Page:       result.CurrentPage,
			Limit:      filters.Limit,
			TotalPages: result.TotalPages,
			TotalItems: result.TotalItems,
			HasNext:    result.HasNext,
			HasPrev:    result.HasPrev,
		},
	}, nil
}

// CancelOpportunity cancels an opportunity and notifies registered volunteers
func (s *opportunityService) CancelOpportunity(ctx context.Context, oppID uuid.UUID, reason string, userID uuid.UUID) error {
	opp, err := s.repo.FindOpportunityByID(ctx, oppID)
	if err != nil {
		if errors.Is(err, repositories.ErrOpportunityNotFound) {
			return apperrors.NewNotFoundError("opportunity")
		}
		return apperrors.NewInternalServerError("Failed to get opportunity")
	}

	if opp.CreatedByUserID != userID {
		return apperrors.NewForbiddenError("You don't have permission to cancel this opportunity")
	}

	if opp.Status == models.OpportunityStatusCancelled {
		return apperrors.NewBadRequestError("Opportunity is already cancelled")
	}

	now := time.Now()
	opp.Status = models.OpportunityStatusCancelled
	opp.CancelledAt = &now
	opp.CancellationReason = &reason

	if err := s.repo.UpdateOpportunity(ctx, opp); err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to cancel opportunity")
		return apperrors.NewInternalServerError("Failed to cancel opportunity")
	}

	// Notify volunteers
	if s.notificationService != nil {
		if err := s.notificationService.NotifyVolunteersOfCancellation(ctx, oppID, reason); err != nil {
			s.logger.WithField("error", err.Error()).Warn("Failed to notify volunteers of cancellation")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"opportunity_id": oppID,
		"reason":         reason,
	}).Info("Opportunity cancelled successfully")

	return nil
}

// CompleteOpportunity marks an opportunity as completed
func (s *opportunityService) CompleteOpportunity(ctx context.Context, oppID uuid.UUID, userID uuid.UUID) error {
	opp, err := s.repo.FindOpportunityByID(ctx, oppID)
	if err != nil {
		if errors.Is(err, repositories.ErrOpportunityNotFound) {
			return apperrors.NewNotFoundError("opportunity")
		}
		return apperrors.NewInternalServerError("Failed to get opportunity")
	}

	if opp.CreatedByUserID != userID {
		return apperrors.NewForbiddenError("You don't have permission to complete this opportunity")
	}

	if opp.Status == models.OpportunityStatusCompleted {
		return apperrors.NewBadRequestError("Opportunity is already completed")
	}

	now := time.Now()
	opp.Status = models.OpportunityStatusCompleted
	opp.CompletedAt = &now

	if err := s.repo.UpdateOpportunity(ctx, opp); err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to complete opportunity")
		return apperrors.NewInternalServerError("Failed to complete opportunity")
	}

	s.logger.WithField("opportunity_id", oppID).Info("Opportunity completed successfully")
	return nil
}

// CreateRecurringOpportunities generates child instances for a recurring opportunity
func (s *opportunityService) CreateRecurringOpportunities(ctx context.Context, oppID uuid.UUID) ([]models.Opportunity, error) {
	opp, err := s.repo.FindOpportunityByID(ctx, oppID)
	if err != nil {
		if errors.Is(err, repositories.ErrOpportunityNotFound) {
			return nil, apperrors.NewNotFoundError("opportunity")
		}
		return nil, apperrors.NewInternalServerError("Failed to get opportunity")
	}

	if !opp.IsRecurring {
		return nil, apperrors.NewBadRequestError("Opportunity is not recurring")
	}

	instances, err := s.repo.CreateRecurringInstances(ctx, opp)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to create recurring instances")
		return nil, apperrors.NewInternalServerError("Failed to create recurring instances")
	}

	s.logger.WithFields(map[string]interface{}{
		"opportunity_id": oppID,
		"count":          len(instances),
	}).Info("Recurring instances created")

	return instances, nil
}

// AutoCompleteOpportunities finds and auto-completes opportunities 7 days after end date
func (s *opportunityService) AutoCompleteOpportunities(ctx context.Context) error {
	opportunities, err := s.repo.FindOpportunitiesForAutoComplete(ctx)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to find opportunities for auto-complete")
		return fmt.Errorf("failed to find opportunities for auto-complete: %w", err)
	}

	completedCount := 0
	for _, opp := range opportunities {
		now := time.Now()
		opp.Status = models.OpportunityStatusCompleted
		opp.CompletedAt = &now

		if err := s.repo.UpdateOpportunity(ctx, &opp); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"opportunity_id": opp.ID,
				"error":          err.Error(),
			}).Error("Failed to auto-complete opportunity")
			continue
		}

		completedCount++
	}

	s.logger.WithFields(map[string]interface{}{
		"completed": completedCount,
		"total":     len(opportunities),
	}).Info("Auto-completed opportunities")

	return nil
}
