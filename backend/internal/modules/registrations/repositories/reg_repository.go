package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrRegistrationNotFound is returned when a registration cannot be found
	ErrRegistrationNotFound = errors.New("registration not found")
	// ErrRegistrationAlreadyExists is returned when a volunteer tries to register for the same opportunity twice
	ErrRegistrationAlreadyExists = errors.New("registration already exists for this volunteer and opportunity")
	// ErrInvalidRegistrationID is returned when the provided registration ID is invalid
	ErrInvalidRegistrationID = errors.New("invalid registration ID")
	// ErrInvalidOpportunityID is returned when the provided opportunity ID is invalid
	ErrInvalidOpportunityID = errors.New("invalid opportunity ID")
	// ErrInvalidVolunteerProfileID is returned when the provided volunteer profile ID is invalid
	ErrInvalidVolunteerProfileID = errors.New("invalid volunteer profile ID")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// RegistrationRepository defines the interface for registration data access
// Following Clean Architecture principles, this interface allows for dependency injection
// and makes the service layer independent of database implementation details
type RegistrationRepository interface {
	// CreateRegistration creates a new registration for a volunteer to an opportunity
	// Returns ErrRegistrationAlreadyExists if the volunteer is already registered
	CreateRegistration(ctx context.Context, registration *models.Registration) error

	// FindRegistrationByID retrieves a registration by its unique identifier
	// Returns ErrRegistrationNotFound if no registration exists with the given ID
	FindRegistrationByID(ctx context.Context, id uuid.UUID) (*models.Registration, error)

	// FindRegistrationsByVolunteer retrieves all registrations for a specific volunteer
	// Returns empty slice if no registrations found
	FindRegistrationsByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.Registration, error)

	// FindRegistrationsByOpportunity retrieves all registrations for a specific opportunity
	// Returns empty slice if no registrations found
	FindRegistrationsByOpportunity(ctx context.Context, opportunityID uuid.UUID) ([]*models.Registration, error)

	// FindRegistrationByVolunteerAndOpportunity finds a specific registration by volunteer and opportunity
	// Returns ErrRegistrationNotFound if no registration exists
	FindRegistrationByVolunteerAndOpportunity(ctx context.Context, volunteerProfileID, opportunityID uuid.UUID) (*models.Registration, error)

	// UpdateRegistrationStatus updates the status of a registration
	UpdateRegistrationStatus(ctx context.Context, id uuid.UUID, status models.RegistrationStatus) error

	// CheckIn marks a volunteer as checked in for an event
	CheckIn(ctx context.Context, id uuid.UUID, checkedInAt time.Time) error

	// CancelRegistration cancels a registration with optional reason
	CancelRegistration(ctx context.Context, id uuid.UUID, reason *string, cancelledAt time.Time) error

	// UpdateRegistration updates an existing registration with modified fields
	UpdateRegistration(ctx context.Context, registration *models.Registration) error

	// LogHours logs volunteer hours for a registration
	LogHours(ctx context.Context, id uuid.UUID, hours float64, status models.HoursStatus, loggedAt time.Time) error

	// VerifyHours marks logged hours as verified by the volunteer
	VerifyHours(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error

	// DisputeHours marks logged hours as disputed by the volunteer
	DisputeHours(ctx context.Context, id uuid.UUID) error

	// FindOverlappingRegistrations finds confirmed registrations for a volunteer that overlap with a given time range
	// Used to warn volunteers about scheduling conflicts
	FindOverlappingRegistrations(ctx context.Context, volunteerProfileID uuid.UUID, startTime, endTime time.Time) ([]*models.Registration, error)

	// CountRegistrationsByOpportunity counts the number of confirmed registrations for an opportunity
	// Used to check capacity
	CountConfirmedRegistrationsByOpportunity(ctx context.Context, opportunityID uuid.UUID) (int64, error)

	// FindWaitlistedRegistrations finds all waitlisted registrations for an opportunity
	FindWaitlistedRegistrations(ctx context.Context, opportunityID uuid.UUID) ([]*models.Registration, error)
}

// gormRegistrationRepository is the GORM implementation of RegistrationRepository
type gormRegistrationRepository struct {
	db *gorm.DB
}

// NewRegistrationRepository creates a new instance of RegistrationRepository using GORM
func NewRegistrationRepository(db *gorm.DB) RegistrationRepository {
	return &gormRegistrationRepository{
		db: db,
	}
}

// CreateRegistration creates a new registration for a volunteer to an opportunity
func (r *gormRegistrationRepository) CreateRegistration(ctx context.Context, registration *models.Registration) error {
	// Validate input
	if registration == nil {
		return fmt.Errorf("registration cannot be nil")
	}

	if registration.OpportunityID == uuid.Nil {
		return ErrInvalidOpportunityID
	}

	if registration.VolunteerProfileID == uuid.Nil {
		return ErrInvalidVolunteerProfileID
	}

	// Use a transaction to ensure atomic operation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if registration already exists
		var existingReg models.Registration
		result := tx.Where("opportunity_id = ? AND volunteer_profile_id = ? AND deleted_at IS NULL",
			registration.OpportunityID, registration.VolunteerProfileID).First(&existingReg)

		if result.Error == nil {
			// Registration found, already exists
			return ErrRegistrationAlreadyExists
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Unexpected error during lookup
			return fmt.Errorf("failed to check existing registration: %w", result.Error)
		}

		// Create the registration
		if err := tx.Create(registration).Error; err != nil {
			return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		}

		return nil
	})

	return err
}

// FindRegistrationByID retrieves a registration by its unique identifier
func (r *gormRegistrationRepository) FindRegistrationByID(ctx context.Context, id uuid.UUID) (*models.Registration, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidRegistrationID
	}

	var registration models.Registration
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&registration)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrRegistrationNotFound
	}

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return &registration, nil
}

// FindRegistrationsByVolunteer retrieves all registrations for a specific volunteer
func (r *gormRegistrationRepository) FindRegistrationsByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.Registration, error) {
	if volunteerProfileID == uuid.Nil {
		return nil, ErrInvalidVolunteerProfileID
	}

	var registrations []*models.Registration
	result := r.db.WithContext(ctx).
		Where("volunteer_profile_id = ?", volunteerProfileID).
		Order("registered_at DESC").
		Find(&registrations)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return registrations, nil
}

// FindRegistrationsByOpportunity retrieves all registrations for a specific opportunity
func (r *gormRegistrationRepository) FindRegistrationsByOpportunity(ctx context.Context, opportunityID uuid.UUID) ([]*models.Registration, error) {
	if opportunityID == uuid.Nil {
		return nil, ErrInvalidOpportunityID
	}

	var registrations []*models.Registration
	result := r.db.WithContext(ctx).
		Where("opportunity_id = ?", opportunityID).
		Order("registered_at ASC").
		Find(&registrations)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return registrations, nil
}

// FindRegistrationByVolunteerAndOpportunity finds a specific registration by volunteer and opportunity
func (r *gormRegistrationRepository) FindRegistrationByVolunteerAndOpportunity(ctx context.Context, volunteerProfileID, opportunityID uuid.UUID) (*models.Registration, error) {
	if volunteerProfileID == uuid.Nil {
		return nil, ErrInvalidVolunteerProfileID
	}

	if opportunityID == uuid.Nil {
		return nil, ErrInvalidOpportunityID
	}

	var registration models.Registration
	result := r.db.WithContext(ctx).
		Where("volunteer_profile_id = ? AND opportunity_id = ?", volunteerProfileID, opportunityID).
		First(&registration)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrRegistrationNotFound
	}

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return &registration, nil
}

// UpdateRegistrationStatus updates the status of a registration
func (r *gormRegistrationRepository) UpdateRegistrationStatus(ctx context.Context, id uuid.UUID, status models.RegistrationStatus) error {
	if id == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	result := r.db.WithContext(ctx).
		Model(&models.Registration{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

// CheckIn marks a volunteer as checked in for an event
func (r *gormRegistrationRepository) CheckIn(ctx context.Context, id uuid.UUID, checkedInAt time.Time) error {
	if id == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	result := r.db.WithContext(ctx).
		Model(&models.Registration{}).
		Where("id = ?", id).
		Update("checked_in_at", checkedInAt)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

// CancelRegistration cancels a registration with optional reason
func (r *gormRegistrationRepository) CancelRegistration(ctx context.Context, id uuid.UUID, reason *string, cancelledAt time.Time) error {
	if id == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	updates := map[string]interface{}{
		"status":       models.RegistrationStatusCancelled,
		"cancelled_at": cancelledAt,
	}

	if reason != nil {
		updates["cancellation_reason"] = *reason
	}

	result := r.db.WithContext(ctx).
		Model(&models.Registration{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

// UpdateRegistration updates an existing registration with modified fields
func (r *gormRegistrationRepository) UpdateRegistration(ctx context.Context, registration *models.Registration) error {
	if registration == nil {
		return fmt.Errorf("registration cannot be nil")
	}

	if registration.ID == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	result := r.db.WithContext(ctx).Save(registration)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

// LogHours logs volunteer hours for a registration
func (r *gormRegistrationRepository) LogHours(ctx context.Context, id uuid.UUID, hours float64, status models.HoursStatus, loggedAt time.Time) error {
	if id == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	updates := map[string]interface{}{
		"hours_worked":    hours,
		"hours_status":    status,
		"hours_logged_at": loggedAt,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Registration{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

// VerifyHours marks logged hours as verified by the volunteer
func (r *gormRegistrationRepository) VerifyHours(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error {
	if id == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	updates := map[string]interface{}{
		"hours_status":      models.HoursStatusVerified,
		"hours_verified_at": verifiedAt,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Registration{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

// DisputeHours marks logged hours as disputed by the volunteer
func (r *gormRegistrationRepository) DisputeHours(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	result := r.db.WithContext(ctx).
		Model(&models.Registration{}).
		Where("id = ?", id).
		Update("hours_status", models.HoursStatusDisputed)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrRegistrationNotFound
	}

	return nil
}

// FindOverlappingRegistrations finds confirmed registrations for a volunteer that overlap with a given time range
// Note: This requires joining with opportunities table to check event times
// For now, we'll need to implement this at the service layer with opportunity data
func (r *gormRegistrationRepository) FindOverlappingRegistrations(ctx context.Context, volunteerProfileID uuid.UUID, startTime, endTime time.Time) ([]*models.Registration, error) {
	if volunteerProfileID == uuid.Nil {
		return nil, ErrInvalidVolunteerProfileID
	}

	// This query needs to join with opportunities table
	// Implementation will be in service layer where we have access to opportunity repository
	var registrations []*models.Registration
	result := r.db.WithContext(ctx).
		Where("volunteer_profile_id = ? AND status = ?", volunteerProfileID, models.RegistrationStatusConfirmed).
		Find(&registrations)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return registrations, nil
}

// CountConfirmedRegistrationsByOpportunity counts the number of confirmed registrations for an opportunity
func (r *gormRegistrationRepository) CountConfirmedRegistrationsByOpportunity(ctx context.Context, opportunityID uuid.UUID) (int64, error) {
	if opportunityID == uuid.Nil {
		return 0, ErrInvalidOpportunityID
	}

	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.Registration{}).
		Where("opportunity_id = ? AND status = ?", opportunityID, models.RegistrationStatusConfirmed).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return count, nil
}

// FindWaitlistedRegistrations finds all waitlisted registrations for an opportunity
func (r *gormRegistrationRepository) FindWaitlistedRegistrations(ctx context.Context, opportunityID uuid.UUID) ([]*models.Registration, error) {
	if opportunityID == uuid.Nil {
		return nil, ErrInvalidOpportunityID
	}

	var registrations []*models.Registration
	result := r.db.WithContext(ctx).
		Where("opportunity_id = ? AND status = ?", opportunityID, models.RegistrationStatusWaitlisted).
		Order("registered_at ASC").
		Find(&registrations)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return registrations, nil
}
