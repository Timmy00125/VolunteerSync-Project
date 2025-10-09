package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrHoursLogNotFound is returned when a hours log cannot be found
	ErrHoursLogNotFound = errors.New("hours log not found")
	// ErrInvalidHoursLogID is returned when the provided hours log ID is invalid
	ErrInvalidHoursLogID = errors.New("invalid hours log ID")
	// ErrInvalidRegistrationID is returned when the provided registration ID is invalid
	ErrInvalidRegistrationID = errors.New("invalid registration ID")
	// ErrInvalidUserID is returned when the provided user ID is invalid
	ErrInvalidUserID = errors.New("invalid user ID")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
	// ErrInvalidHoursAmount is returned when hours amount is zero or negative
	ErrInvalidHoursAmount = errors.New("hours amount must be positive")
	// ErrInvalidStatusTransition is returned when attempting an invalid status change
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// HoursRepository defines the interface for hours log data access
// Following Clean Architecture principles, this interface allows for dependency injection
// and makes the service layer independent of database implementation details
type HoursRepository interface {
	// CreateHoursLog creates a new immutable hours log entry
	// Hours logs are append-only for audit trail compliance (FR-054)
	CreateHoursLog(ctx context.Context, hoursLog *models.HoursLog) error

	// FindHoursLogByID retrieves a hours log by its unique identifier
	// Returns ErrHoursLogNotFound if no log exists with the given ID
	FindHoursLogByID(ctx context.Context, id uuid.UUID) (*models.HoursLog, error)

	// FindHoursLogsByRegistration retrieves all hours logs for a specific registration
	// Returns empty slice if no logs found
	FindHoursLogsByRegistration(ctx context.Context, registrationID uuid.UUID) ([]*models.HoursLog, error)

	// FindHoursLogsByVolunteer retrieves all hours logs for a specific volunteer
	// Joins through registrations to get all logs for a volunteer
	FindHoursLogsByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.HoursLog, error)

	// FindHoursLogsByCoordinator retrieves all hours logs created by a specific coordinator
	// Returns empty slice if no logs found
	FindHoursLogsByCoordinator(ctx context.Context, coordinatorUserID uuid.UUID) ([]*models.HoursLog, error)

	// UpdateHoursStatus updates only the status of a hours log
	// This is the only field that can be modified (status transitions only)
	// Status transitions: pending → verified, pending → disputed, disputed → verified
	UpdateHoursStatus(ctx context.Context, id uuid.UUID, status models.HoursStatus) error

	// VerifyHours marks a hours log as verified by the volunteer
	VerifyHours(ctx context.Context, id uuid.UUID, verifiedAt time.Time, volunteerNotes *string) error

	// DisputeHours marks a hours log as disputed by the volunteer
	DisputeHours(ctx context.Context, id uuid.UUID, disputeReason string, disputedAt time.Time) error

	// ResolveDispute marks a disputed hours log as resolved
	ResolveDispute(ctx context.Context, id uuid.UUID, resolutionNotes string, resolvedAt time.Time) error

	// AutoVerifyHours marks a hours log as auto-verified after 7 days
	// Sets both verified_at and auto_verified_at timestamps
	AutoVerifyHours(ctx context.Context, id uuid.UUID, autoVerifiedAt time.Time) error

	// FindPendingHoursOlderThan7Days finds all pending hours logs older than 7 days
	// Used by the auto-verification cron job (FR-049)
	FindPendingHoursOlderThan7Days(ctx context.Context, cutoffDate time.Time) ([]*models.HoursLog, error)

	// FindHoursLogsByStatus retrieves all hours logs with a specific status
	// Useful for admin dashboards and reporting
	FindHoursLogsByStatus(ctx context.Context, status models.HoursStatus) ([]*models.HoursLog, error)

	// GetTotalVerifiedHoursByVolunteer calculates the sum of all verified hours for a volunteer
	// Used for volunteer dashboard and impact metrics
	GetTotalVerifiedHoursByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) (float64, error)
}

// gormHoursRepository is the GORM implementation of HoursRepository
type gormHoursRepository struct {
	db *gorm.DB
}

// NewHoursRepository creates a new instance of HoursRepository using GORM
func NewHoursRepository(db *gorm.DB) HoursRepository {
	return &gormHoursRepository{
		db: db,
	}
}

// CreateHoursLog creates a new immutable hours log entry
func (r *gormHoursRepository) CreateHoursLog(ctx context.Context, hoursLog *models.HoursLog) error {
	// Validate input
	if hoursLog == nil {
		return fmt.Errorf("hours log cannot be nil")
	}

	if hoursLog.RegistrationID == uuid.Nil {
		return ErrInvalidRegistrationID
	}

	if hoursLog.LoggedByUserID == uuid.Nil {
		return ErrInvalidUserID
	}

	if hoursLog.Hours <= 0 {
		return ErrInvalidHoursAmount
	}

	// Create the hours log
	if err := r.db.WithContext(ctx).Create(hoursLog).Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return nil
}

// FindHoursLogByID retrieves a hours log by its unique identifier
func (r *gormHoursRepository) FindHoursLogByID(ctx context.Context, id uuid.UUID) (*models.HoursLog, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidHoursLogID
	}

	var hoursLog models.HoursLog
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&hoursLog)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrHoursLogNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return &hoursLog, nil
}

// FindHoursLogsByRegistration retrieves all hours logs for a specific registration
func (r *gormHoursRepository) FindHoursLogsByRegistration(ctx context.Context, registrationID uuid.UUID) ([]*models.HoursLog, error) {
	if registrationID == uuid.Nil {
		return nil, ErrInvalidRegistrationID
	}

	var hoursLogs []*models.HoursLog
	result := r.db.WithContext(ctx).
		Where("registration_id = ?", registrationID).
		Order("logged_at DESC").
		Find(&hoursLogs)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return hoursLogs, nil
}

// FindHoursLogsByVolunteer retrieves all hours logs for a specific volunteer
func (r *gormHoursRepository) FindHoursLogsByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.HoursLog, error) {
	if volunteerProfileID == uuid.Nil {
		return nil, fmt.Errorf("invalid volunteer profile ID")
	}

	var hoursLogs []*models.HoursLog
	result := r.db.WithContext(ctx).
		Joins("JOIN registrations ON registrations.id = hours_logs.registration_id").
		Where("registrations.volunteer_profile_id = ? AND registrations.deleted_at IS NULL", volunteerProfileID).
		Order("hours_logs.logged_at DESC").
		Find(&hoursLogs)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return hoursLogs, nil
}

// FindHoursLogsByCoordinator retrieves all hours logs created by a specific coordinator
func (r *gormHoursRepository) FindHoursLogsByCoordinator(ctx context.Context, coordinatorUserID uuid.UUID) ([]*models.HoursLog, error) {
	if coordinatorUserID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	var hoursLogs []*models.HoursLog
	result := r.db.WithContext(ctx).
		Where("logged_by_user_id = ?", coordinatorUserID).
		Order("logged_at DESC").
		Find(&hoursLogs)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return hoursLogs, nil
}

// UpdateHoursStatus updates only the status of a hours log
func (r *gormHoursRepository) UpdateHoursStatus(ctx context.Context, id uuid.UUID, status models.HoursStatus) error {
	if id == uuid.Nil {
		return ErrInvalidHoursLogID
	}

	// Validate status
	if status != models.HoursStatusPending && status != models.HoursStatusVerified && status != models.HoursStatusDisputed {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Get existing log to validate status transition
	existingLog, err := r.FindHoursLogByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate status transition
	// Valid transitions: pending → verified, pending → disputed, disputed → verified
	if existingLog.Status == models.HoursStatusVerified {
		return ErrInvalidStatusTransition
	}

	result := r.db.WithContext(ctx).
		Model(&models.HoursLog{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrHoursLogNotFound
	}

	return nil
}

// VerifyHours marks a hours log as verified by the volunteer
func (r *gormHoursRepository) VerifyHours(ctx context.Context, id uuid.UUID, verifiedAt time.Time, volunteerNotes *string) error {
	if id == uuid.Nil {
		return ErrInvalidHoursLogID
	}

	// Get existing log to validate it's in pending status
	existingLog, err := r.FindHoursLogByID(ctx, id)
	if err != nil {
		return err
	}

	if !existingLog.CanBeVerified() {
		return ErrInvalidStatusTransition
	}

	updates := map[string]interface{}{
		"status":      models.HoursStatusVerified,
		"verified_at": verifiedAt,
		"updated_at":  time.Now(),
	}

	if volunteerNotes != nil {
		updates["volunteer_notes"] = *volunteerNotes
	}

	result := r.db.WithContext(ctx).
		Model(&models.HoursLog{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrHoursLogNotFound
	}

	return nil
}

// DisputeHours marks a hours log as disputed by the volunteer
func (r *gormHoursRepository) DisputeHours(ctx context.Context, id uuid.UUID, disputeReason string, disputedAt time.Time) error {
	if id == uuid.Nil {
		return ErrInvalidHoursLogID
	}

	if disputeReason == "" {
		return fmt.Errorf("dispute reason is required")
	}

	// Get existing log to validate it's in pending status
	existingLog, err := r.FindHoursLogByID(ctx, id)
	if err != nil {
		return err
	}

	if !existingLog.CanBeDisputed() {
		return ErrInvalidStatusTransition
	}

	updates := map[string]interface{}{
		"status":         models.HoursStatusDisputed,
		"dispute_reason": disputeReason,
		"disputed_at":    disputedAt,
		"updated_at":     time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.HoursLog{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrHoursLogNotFound
	}

	return nil
}

// ResolveDispute marks a disputed hours log as resolved
func (r *gormHoursRepository) ResolveDispute(ctx context.Context, id uuid.UUID, resolutionNotes string, resolvedAt time.Time) error {
	if id == uuid.Nil {
		return ErrInvalidHoursLogID
	}

	if resolutionNotes == "" {
		return fmt.Errorf("resolution notes are required")
	}

	// Get existing log to validate it's in disputed status
	existingLog, err := r.FindHoursLogByID(ctx, id)
	if err != nil {
		return err
	}

	if !existingLog.IsDisputed() {
		return fmt.Errorf("can only resolve disputed hours logs")
	}

	updates := map[string]interface{}{
		"status":           models.HoursStatusVerified,
		"resolution_notes": resolutionNotes,
		"resolved_at":      resolvedAt,
		"verified_at":      resolvedAt,
		"updated_at":       time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.HoursLog{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrHoursLogNotFound
	}

	return nil
}

// AutoVerifyHours marks a hours log as auto-verified after 7 days
func (r *gormHoursRepository) AutoVerifyHours(ctx context.Context, id uuid.UUID, autoVerifiedAt time.Time) error {
	if id == uuid.Nil {
		return ErrInvalidHoursLogID
	}

	// Get existing log to validate it's in pending status
	existingLog, err := r.FindHoursLogByID(ctx, id)
	if err != nil {
		return err
	}

	if !existingLog.IsPending() {
		return fmt.Errorf("can only auto-verify pending hours logs")
	}

	updates := map[string]interface{}{
		"status":           models.HoursStatusVerified,
		"verified_at":      autoVerifiedAt,
		"auto_verified_at": autoVerifiedAt,
		"updated_at":       time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.HoursLog{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrHoursLogNotFound
	}

	return nil
}

// FindPendingHoursOlderThan7Days finds all pending hours logs older than 7 days
func (r *gormHoursRepository) FindPendingHoursOlderThan7Days(ctx context.Context, cutoffDate time.Time) ([]*models.HoursLog, error) {
	var hoursLogs []*models.HoursLog
	result := r.db.WithContext(ctx).
		Where("status = ? AND logged_at < ?", models.HoursStatusPending, cutoffDate).
		Order("logged_at ASC").
		Find(&hoursLogs)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return hoursLogs, nil
}

// FindHoursLogsByStatus retrieves all hours logs with a specific status
func (r *gormHoursRepository) FindHoursLogsByStatus(ctx context.Context, status models.HoursStatus) ([]*models.HoursLog, error) {
	var hoursLogs []*models.HoursLog
	result := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("logged_at DESC").
		Find(&hoursLogs)

	if result.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return hoursLogs, nil
}

// GetTotalVerifiedHoursByVolunteer calculates the sum of all verified hours for a volunteer
func (r *gormHoursRepository) GetTotalVerifiedHoursByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) (float64, error) {
	if volunteerProfileID == uuid.Nil {
		return 0, fmt.Errorf("invalid volunteer profile ID")
	}

	var totalHours float64
	result := r.db.WithContext(ctx).
		Table("hours_logs").
		Select("COALESCE(SUM(hours_logs.hours), 0) as total").
		Joins("JOIN registrations ON registrations.id = hours_logs.registration_id").
		Where("registrations.volunteer_profile_id = ? AND hours_logs.status = ? AND registrations.deleted_at IS NULL AND hours_logs.deleted_at IS NULL",
			volunteerProfileID, models.HoursStatusVerified).
		Scan(&totalHours)

	if result.Error != nil {
		return 0, fmt.Errorf("%w: %v", ErrDatabaseOperation, result.Error)
	}

	return totalHours, nil
}
