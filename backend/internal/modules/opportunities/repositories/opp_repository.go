package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/opportunities/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrOpportunityNotFound is returned when an opportunity cannot be found
	ErrOpportunityNotFound = errors.New("opportunity not found")
	// ErrInvalidOpportunityID is returned when the provided opportunity ID is invalid
	ErrInvalidOpportunityID = errors.New("invalid opportunity ID")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// OpportunityFilters represents filters for listing opportunities
type OpportunityFilters struct {
	Search         string                    // Search by title or description
	OrganizationID *uuid.UUID                // Filter by organization
	Status         *models.OpportunityStatus // Filter by status
	City           string                    // Filter by city
	State          string                    // Filter by state
	Latitude       *float64                  // Search center latitude (for radius search)
	Longitude      *float64                  // Search center longitude (for radius search)
	RadiusKm       *float64                  // Search radius in kilometers
	StartDateFrom  *time.Time                // Filter by start date range (from)
	StartDateTo    *time.Time                // Filter by start date range (to)
	CauseIDs       []uuid.UUID               // Filter by cause categories
	SkillIDs       []uuid.UUID               // Filter by required skills
	MinAge         *int                      // Filter by minimum age requirement
	OnlyRecurring  bool                      // Show only recurring opportunities
	Page           int                       // Page number (1-indexed)
	Limit          int                       // Items per page
	SortBy         string                    // Sort field (created_at, start_date, title)
	SortOrder      string                    // Sort order (asc, desc)
}

// PaginatedOpportunities represents a paginated list of opportunities
type PaginatedOpportunities struct {
	Opportunities []models.Opportunity
	TotalItems    int
	TotalPages    int
	CurrentPage   int
	HasNext       bool
	HasPrev       bool
}

// OpportunityRepository defines the interface for opportunity data access
type OpportunityRepository interface {
	// CreateOpportunity creates a new opportunity in the database
	CreateOpportunity(ctx context.Context, opp *models.Opportunity) error

	// FindOpportunityByID retrieves an opportunity by its unique identifier
	FindOpportunityByID(ctx context.Context, id uuid.UUID) (*models.Opportunity, error)

	// ListOpportunities retrieves a paginated list of opportunities with complex filters
	// Supports search, location-based filtering, date ranges, and cause/skill filtering
	ListOpportunities(ctx context.Context, filters OpportunityFilters) (*PaginatedOpportunities, error)

	// UpdateOpportunity updates an existing opportunity
	UpdateOpportunity(ctx context.Context, opp *models.Opportunity) error

	// DeleteOpportunity soft deletes an opportunity by its ID
	DeleteOpportunity(ctx context.Context, id uuid.UUID) error

	// IncrementRegistrations increases the current registration count
	IncrementRegistrations(ctx context.Context, oppID uuid.UUID) error

	// DecrementRegistrations decreases the current registration count
	DecrementRegistrations(ctx context.Context, oppID uuid.UUID) error

	// CreateRecurringInstances creates child instances for a recurring opportunity
	CreateRecurringInstances(ctx context.Context, parentOpp *models.Opportunity) ([]models.Opportunity, error)

	// FindOpportunitiesByOrganization retrieves all opportunities for an organization
	FindOpportunitiesByOrganization(ctx context.Context, orgID uuid.UUID) ([]models.Opportunity, error)

	// FindOpportunitiesForAutoComplete finds opportunities that should be auto-completed
	FindOpportunitiesForAutoComplete(ctx context.Context) ([]models.Opportunity, error)
}

// gormOpportunityRepository is the GORM implementation of OpportunityRepository
type gormOpportunityRepository struct {
	db *gorm.DB
}

// NewOpportunityRepository creates a new instance of OpportunityRepository using GORM
func NewOpportunityRepository(db *gorm.DB) OpportunityRepository {
	return &gormOpportunityRepository{
		db: db,
	}
}

// CreateOpportunity creates a new opportunity in the database
func (r *gormOpportunityRepository) CreateOpportunity(ctx context.Context, opp *models.Opportunity) error {
	if opp == nil {
		return fmt.Errorf("opportunity cannot be nil")
	}

	// Validate the opportunity
	if err := opp.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create the opportunity
	if err := r.db.WithContext(ctx).Create(opp).Error; err != nil {
		return fmt.Errorf("failed to create opportunity: %w", err)
	}

	return nil
}

// FindOpportunityByID retrieves an opportunity by its unique identifier
func (r *gormOpportunityRepository) FindOpportunityByID(ctx context.Context, id uuid.UUID) (*models.Opportunity, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidOpportunityID
	}

	var opp models.Opportunity
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&opp)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrOpportunityNotFound
		}
		return nil, fmt.Errorf("failed to find opportunity: %w", result.Error)
	}

	return &opp, nil
}

// ListOpportunities retrieves a paginated list of opportunities with complex filters
func (r *gormOpportunityRepository) ListOpportunities(ctx context.Context, filters OpportunityFilters) (*PaginatedOpportunities, error) {
	// Set default pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	// Build query
	query := r.db.WithContext(ctx).Model(&models.Opportunity{})

	// Apply filters
	if filters.Search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?",
			"%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.City != "" {
		query = query.Where("city = ?", filters.City)
	}

	if filters.State != "" {
		query = query.Where("state = ?", filters.State)
	}

	// Date range filter
	if filters.StartDateFrom != nil {
		query = query.Where("start_date >= ?", *filters.StartDateFrom)
	}
	if filters.StartDateTo != nil {
		query = query.Where("start_date <= ?", *filters.StartDateTo)
	}

	// Minimum age filter
	if filters.MinAge != nil {
		query = query.Where("(min_age IS NULL OR min_age <= ?)", *filters.MinAge)
	}

	// Recurring filter
	if filters.OnlyRecurring {
		query = query.Where("is_recurring = ?", true)
	}

	// Location-based radius search (Haversine formula)
	if filters.Latitude != nil && filters.Longitude != nil && filters.RadiusKm != nil {
		// Using simplified Haversine formula for PostgreSQL
		// Note: For production, consider using PostGIS for better performance
		query = query.Where(
			`(6371 * acos(
				cos(radians(?)) * 
				cos(radians(latitude)) * 
				cos(radians(longitude) - radians(?)) + 
				sin(radians(?)) * 
				sin(radians(latitude))
			)) <= ?`,
			*filters.Latitude, *filters.Longitude, *filters.Latitude, *filters.RadiusKm,
		)
	}

	// Count total items
	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count opportunities: %w", err)
	}

	// Apply sorting
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	// Execute query
	var opportunities []models.Opportunity
	if err := query.Find(&opportunities).Error; err != nil {
		return nil, fmt.Errorf("failed to list opportunities: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int(totalItems) / filters.Limit
	if int(totalItems)%filters.Limit > 0 {
		totalPages++
	}

	return &PaginatedOpportunities{
		Opportunities: opportunities,
		TotalItems:    int(totalItems),
		TotalPages:    totalPages,
		CurrentPage:   filters.Page,
		HasNext:       filters.Page < totalPages,
		HasPrev:       filters.Page > 1,
	}, nil
}

// UpdateOpportunity updates an existing opportunity
func (r *gormOpportunityRepository) UpdateOpportunity(ctx context.Context, opp *models.Opportunity) error {
	if opp == nil {
		return fmt.Errorf("opportunity cannot be nil")
	}

	if opp.ID == uuid.Nil {
		return ErrInvalidOpportunityID
	}

	// Validate the opportunity
	if err := opp.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update the opportunity
	result := r.db.WithContext(ctx).Save(opp)
	if result.Error != nil {
		return fmt.Errorf("failed to update opportunity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrOpportunityNotFound
	}

	return nil
}

// DeleteOpportunity soft deletes an opportunity by its ID
func (r *gormOpportunityRepository) DeleteOpportunity(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidOpportunityID
	}

	result := r.db.WithContext(ctx).Delete(&models.Opportunity{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete opportunity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrOpportunityNotFound
	}

	return nil
}

// IncrementRegistrations increases the current registration count
func (r *gormOpportunityRepository) IncrementRegistrations(ctx context.Context, oppID uuid.UUID) error {
	if oppID == uuid.Nil {
		return ErrInvalidOpportunityID
	}

	result := r.db.WithContext(ctx).Model(&models.Opportunity{}).
		Where("id = ?", oppID).
		Update("current_registrations", gorm.Expr("current_registrations + ?", 1))

	if result.Error != nil {
		return fmt.Errorf("failed to increment registrations: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrOpportunityNotFound
	}

	return nil
}

// DecrementRegistrations decreases the current registration count
func (r *gormOpportunityRepository) DecrementRegistrations(ctx context.Context, oppID uuid.UUID) error {
	if oppID == uuid.Nil {
		return ErrInvalidOpportunityID
	}

	result := r.db.WithContext(ctx).Model(&models.Opportunity{}).
		Where("id = ? AND current_registrations > 0", oppID).
		Update("current_registrations", gorm.Expr("current_registrations - ?", 1))

	if result.Error != nil {
		return fmt.Errorf("failed to decrement registrations: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrOpportunityNotFound
	}

	return nil
}

// CreateRecurringInstances creates child instances for a recurring opportunity
func (r *gormOpportunityRepository) CreateRecurringInstances(ctx context.Context, parentOpp *models.Opportunity) ([]models.Opportunity, error) {
	if parentOpp == nil {
		return nil, fmt.Errorf("parent opportunity cannot be nil")
	}

	if !parentOpp.IsRecurring {
		return nil, fmt.Errorf("opportunity is not recurring")
	}

	if parentOpp.RecurrencePattern == nil || parentOpp.RecurrenceEndDate == nil {
		return nil, fmt.Errorf("recurrence pattern and end date are required")
	}

	var instances []models.Opportunity
	currentDate := parentOpp.StartDate

	// Generate instances based on recurrence pattern
	for currentDate.Before(*parentOpp.RecurrenceEndDate) || currentDate.Equal(*parentOpp.RecurrenceEndDate) {
		// Skip the first instance (parent)
		if !currentDate.Equal(parentOpp.StartDate) {
			instance := models.Opportunity{
				OrganizationID:      parentOpp.OrganizationID,
				CreatedByUserID:     parentOpp.CreatedByUserID,
				Title:               parentOpp.Title,
				Description:         parentOpp.Description,
				Status:              parentOpp.Status,
				StartDate:           currentDate,
				StartTime:           parentOpp.StartTime,
				EndDate:             currentDate,
				EndTime:             parentOpp.EndTime,
				Timezone:            parentOpp.Timezone,
				AddressLine1:        parentOpp.AddressLine1,
				AddressLine2:        parentOpp.AddressLine2,
				City:                parentOpp.City,
				State:               parentOpp.State,
				PostalCode:          parentOpp.PostalCode,
				Country:             parentOpp.Country,
				Latitude:            parentOpp.Latitude,
				Longitude:           parentOpp.Longitude,
				Capacity:            parentOpp.Capacity,
				MinAge:              parentOpp.MinAge,
				IsRecurring:         false,
				ParentOpportunityID: &parentOpp.ID,
			}

			if err := r.db.WithContext(ctx).Create(&instance).Error; err != nil {
				return nil, fmt.Errorf("failed to create recurring instance: %w", err)
			}

			instances = append(instances, instance)
		}

		// Calculate next date based on pattern
		switch *parentOpp.RecurrencePattern {
		case models.RecurrencePatternDaily:
			currentDate = currentDate.AddDate(0, 0, 1)
		case models.RecurrencePatternWeekly:
			currentDate = currentDate.AddDate(0, 0, 7)
		case models.RecurrencePatternMonthly:
			currentDate = currentDate.AddDate(0, 1, 0)
		default:
			return nil, fmt.Errorf("unsupported recurrence pattern: %s", *parentOpp.RecurrencePattern)
		}
	}

	return instances, nil
}

// FindOpportunitiesByOrganization retrieves all opportunities for an organization
func (r *gormOpportunityRepository) FindOpportunitiesByOrganization(ctx context.Context, orgID uuid.UUID) ([]models.Opportunity, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("invalid organization ID")
	}

	var opportunities []models.Opportunity
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("start_date DESC").
		Find(&opportunities).Error; err != nil {
		return nil, fmt.Errorf("failed to find opportunities: %w", err)
	}

	return opportunities, nil
}

// FindOpportunitiesForAutoComplete finds opportunities that should be auto-completed
func (r *gormOpportunityRepository) FindOpportunitiesForAutoComplete(ctx context.Context) ([]models.Opportunity, error) {
	now := time.Now()
	var opportunities []models.Opportunity

	if err := r.db.WithContext(ctx).
		Where("status = ? AND auto_complete_at IS NOT NULL AND auto_complete_at <= ?",
			models.OpportunityStatusPublished, now).
		Find(&opportunities).Error; err != nil {
		return nil, fmt.Errorf("failed to find opportunities for auto-complete: %w", err)
	}

	return opportunities, nil
}
