package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/hours/repositories"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/google/uuid"
)

// Custom errors for service operations
var (
	// ErrInvalidHoursData is returned when hours data is invalid
	ErrInvalidHoursData = errors.New("invalid hours data")
	// ErrHoursLogNotFound is returned when a hours log cannot be found
	ErrHoursLogNotFound = repositories.ErrHoursLogNotFound
	// ErrUnauthorized is returned when a user doesn't have permission for an operation
	ErrUnauthorized = errors.New("unauthorized to perform this action")
	// ErrInvalidStatusTransition is returned when a status transition is invalid
	ErrInvalidStatusTransition = repositories.ErrInvalidStatusTransition
	// ErrAlreadyVerified is returned when attempting to modify already verified hours
	ErrAlreadyVerified = errors.New("hours are already verified and cannot be modified")
	// ErrAlreadyDisputed is returned when attempting to verify already disputed hours
	ErrAlreadyDisputed = errors.New("hours are already disputed")
)

// RegistrationService defines the interface for registration operations
// This will be injected from the registrations module
type RegistrationService interface {
	GetRegistration(ctx context.Context, registrationID uuid.UUID) (*RegistrationDetails, error)
	UpdateRegistrationHours(ctx context.Context, registrationID uuid.UUID, hours float64, status string) error
}

// RegistrationDetails represents registration data needed by hours service
type RegistrationDetails struct {
	ID                 uuid.UUID
	OpportunityID      uuid.UUID
	VolunteerProfileID uuid.UUID
	Status             string
	CheckedInAt        *time.Time
}

// VolunteerService defines the interface for volunteer operations
// This will be injected from the volunteers module
type VolunteerService interface {
	IncrementTotalHours(ctx context.Context, volunteerProfileID uuid.UUID, hours float64) error
}

// NotificationService defines the interface for sending notifications
// This will be implemented in the communications module
type NotificationService interface {
	SendHoursLoggedNotification(ctx context.Context, volunteerID, registrationID uuid.UUID, hours float64) error
	SendHoursVerifiedNotification(ctx context.Context, coordinatorUserID, registrationID uuid.UUID) error
	SendHoursDisputedNotification(ctx context.Context, coordinatorUserID, registrationID uuid.UUID, reason string) error
	SendDisputeResolvedNotification(ctx context.Context, volunteerID, registrationID uuid.UUID) error
}

// HoursService encapsulates hours tracking business logic
// Provides methods for logging, verifying, disputing, and auto-verifying volunteer hours
type HoursService interface {
	// LogHours logs volunteer hours for a registration (by coordinator)
	// Creates log, updates registration, notifies volunteer (FR-047)
	LogHours(ctx context.Context, input LogHoursInput) (*models.HoursLog, error)

	// VerifyHours verifies logged hours (by volunteer)
	// Updates status to verified, increments volunteer total_hours (FR-048)
	VerifyHours(ctx context.Context, hoursLogID, volunteerProfileID uuid.UUID, notes *string) error

	// DisputeHours disputes logged hours (by volunteer)
	// Updates status to disputed, notifies coordinator (FR-050)
	DisputeHours(ctx context.Context, hoursLogID, volunteerProfileID uuid.UUID, reason string) error

	// ResolveDispute resolves a disputed hours log (by coordinator)
	// Updates status to verified, increments volunteer total_hours
	ResolveDispute(ctx context.Context, hoursLogID, coordinatorUserID uuid.UUID, resolutionNotes string) error

	// AutoVerifyOldHours auto-verifies hours older than 7 days
	// Cron job that runs daily (FR-049)
	AutoVerifyOldHours(ctx context.Context) (int, error)

	// GetHoursLog retrieves a specific hours log by ID
	GetHoursLog(ctx context.Context, hoursLogID uuid.UUID) (*models.HoursLog, error)

	// GetHoursLogsByRegistration retrieves all hours logs for a registration
	GetHoursLogsByRegistration(ctx context.Context, registrationID uuid.UUID) ([]*models.HoursLog, error)

	// GetHoursLogsByVolunteer retrieves all hours logs for a volunteer
	GetHoursLogsByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.HoursLog, error)

	// GetPendingHoursForVolunteer retrieves pending hours logs for a volunteer
	GetPendingHoursForVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.HoursLog, error)
}

// LogHoursInput represents the input for logging volunteer hours
type LogHoursInput struct {
	RegistrationID   uuid.UUID
	Hours            float64
	LoggedByUserID   uuid.UUID // Coordinator who is logging the hours
	CoordinatorNotes *string
}

// hoursService is the implementation of HoursService
type hoursService struct {
	hoursRepo           repositories.HoursRepository
	registrationService RegistrationService
	volunteerService    VolunteerService
	notificationService NotificationService
	logger              logger.Logger
}

// NewHoursService creates a new instance of HoursService
func NewHoursService(
	hoursRepo repositories.HoursRepository,
	registrationService RegistrationService,
	volunteerService VolunteerService,
	notificationService NotificationService,
	logger logger.Logger,
) HoursService {
	return &hoursService{
		hoursRepo:           hoursRepo,
		registrationService: registrationService,
		volunteerService:    volunteerService,
		notificationService: notificationService,
		logger:              logger,
	}
}

// LogHours logs volunteer hours for a registration (by coordinator)
func (s *hoursService) LogHours(ctx context.Context, input LogHoursInput) (*models.HoursLog, error) {
	// Validate input
	if input.RegistrationID == uuid.Nil {
		return nil, fmt.Errorf("%w: registration ID is required", ErrInvalidHoursData)
	}
	if input.LoggedByUserID == uuid.Nil {
		return nil, fmt.Errorf("%w: logged by user ID is required", ErrInvalidHoursData)
	}
	if input.Hours <= 0 {
		return nil, fmt.Errorf("%w: hours must be positive", ErrInvalidHoursData)
	}

	// Get registration details
	registration, err := s.registrationService.GetRegistration(ctx, input.RegistrationID)
	if err != nil {
		s.logger.WithContext(ctx).WithField("registration_id", input.RegistrationID.String()).Error("Failed to get registration")
		return nil, fmt.Errorf("failed to get registration: %w", err)
	}

	// Verify registration is completed or checked in
	if registration.CheckedInAt == nil {
		s.logger.WithContext(ctx).Warn("Cannot log hours for registration without check-in")
		return nil, fmt.Errorf("%w: volunteer must be checked in before logging hours", ErrInvalidHoursData)
	}

	// Create hours log with pending status
	hoursLog := &models.HoursLog{
		RegistrationID:   input.RegistrationID,
		Hours:            input.Hours,
		LoggedByUserID:   input.LoggedByUserID,
		Status:           models.HoursStatusPending,
		CoordinatorNotes: input.CoordinatorNotes,
		LoggedAt:         time.Now(),
	}

	// Save to database
	if err := s.hoursRepo.CreateHoursLog(ctx, hoursLog); err != nil {
		s.logger.WithContext(ctx).Error("Failed to create hours log")
		return nil, fmt.Errorf("failed to create hours log: %w", err)
	}

	// Update registration with hours information
	if err := s.registrationService.UpdateRegistrationHours(ctx, input.RegistrationID, input.Hours, string(models.HoursStatusPending)); err != nil {
		s.logger.WithContext(ctx).Error("Failed to update registration hours")
		// Don't fail the operation, just log the error
	}

	// Send notification to volunteer
	if err := s.notificationService.SendHoursLoggedNotification(ctx, registration.VolunteerProfileID, input.RegistrationID, input.Hours); err != nil {
		s.logger.WithContext(ctx).Error("Failed to send hours logged notification")
		// Don't fail the operation, just log the error
	}

	s.logger.WithContext(ctx).
		WithField("hours_log_id", hoursLog.ID.String()).
		WithField("registration_id", input.RegistrationID.String()).
		WithField("hours", input.Hours).
		Info("Hours logged successfully")

	return hoursLog, nil
}

// VerifyHours verifies logged hours (by volunteer)
func (s *hoursService) VerifyHours(ctx context.Context, hoursLogID, volunteerProfileID uuid.UUID, notes *string) error {
	// Validate input
	if hoursLogID == uuid.Nil {
		return fmt.Errorf("%w: hours log ID is required", ErrInvalidHoursData)
	}
	if volunteerProfileID == uuid.Nil {
		return fmt.Errorf("%w: volunteer profile ID is required", ErrInvalidHoursData)
	}

	// Get hours log
	hoursLog, err := s.hoursRepo.FindHoursLogByID(ctx, hoursLogID)
	if err != nil {
		return err
	}

	// Get registration to verify volunteer ownership
	registration, err := s.registrationService.GetRegistration(ctx, hoursLog.RegistrationID)
	if err != nil {
		return fmt.Errorf("failed to get registration: %w", err)
	}

	// Verify the volunteer owns this registration
	if registration.VolunteerProfileID != volunteerProfileID {
		s.logger.WithContext(ctx).Warn("Unauthorized attempt to verify hours")
		return ErrUnauthorized
	}

	// Verify hours log can be verified
	if !hoursLog.CanBeVerified() {
		return ErrInvalidStatusTransition
	}

	// Verify hours in repository
	verifiedAt := time.Now()
	if err := s.hoursRepo.VerifyHours(ctx, hoursLogID, verifiedAt, notes); err != nil {
		s.logger.WithContext(ctx).Error("Failed to verify hours")
		return fmt.Errorf("failed to verify hours: %w", err)
	}

	// Increment volunteer's total hours
	if err := s.volunteerService.IncrementTotalHours(ctx, volunteerProfileID, hoursLog.Hours); err != nil {
		s.logger.WithContext(ctx).Error("Failed to increment volunteer total hours")
		// Don't fail the operation, just log the error
	}

	// Update registration hours status
	if err := s.registrationService.UpdateRegistrationHours(ctx, hoursLog.RegistrationID, hoursLog.Hours, string(models.HoursStatusVerified)); err != nil {
		s.logger.WithContext(ctx).Error("Failed to update registration hours status")
		// Don't fail the operation, just log the error
	}

	// Send notification to coordinator
	if err := s.notificationService.SendHoursVerifiedNotification(ctx, hoursLog.LoggedByUserID, hoursLog.RegistrationID); err != nil {
		s.logger.WithContext(ctx).Error("Failed to send hours verified notification")
		// Don't fail the operation, just log the error
	}

	s.logger.WithContext(ctx).
		WithField("hours_log_id", hoursLogID.String()).
		WithField("volunteer_profile_id", volunteerProfileID.String()).
		Info("Hours verified successfully")

	return nil
}

// DisputeHours disputes logged hours (by volunteer)
func (s *hoursService) DisputeHours(ctx context.Context, hoursLogID, volunteerProfileID uuid.UUID, reason string) error {
	// Validate input
	if hoursLogID == uuid.Nil {
		return fmt.Errorf("%w: hours log ID is required", ErrInvalidHoursData)
	}
	if volunteerProfileID == uuid.Nil {
		return fmt.Errorf("%w: volunteer profile ID is required", ErrInvalidHoursData)
	}
	if reason == "" {
		return fmt.Errorf("%w: dispute reason is required", ErrInvalidHoursData)
	}

	// Get hours log
	hoursLog, err := s.hoursRepo.FindHoursLogByID(ctx, hoursLogID)
	if err != nil {
		return err
	}

	// Get registration to verify volunteer ownership
	registration, err := s.registrationService.GetRegistration(ctx, hoursLog.RegistrationID)
	if err != nil {
		return fmt.Errorf("failed to get registration: %w", err)
	}

	// Verify the volunteer owns this registration
	if registration.VolunteerProfileID != volunteerProfileID {
		s.logger.WithContext(ctx).Warn("Unauthorized attempt to dispute hours")
		return ErrUnauthorized
	}

	// Verify hours log can be disputed
	if !hoursLog.CanBeDisputed() {
		return ErrInvalidStatusTransition
	}

	// Dispute hours in repository
	disputedAt := time.Now()
	if err := s.hoursRepo.DisputeHours(ctx, hoursLogID, reason, disputedAt); err != nil {
		s.logger.WithContext(ctx).Error("Failed to dispute hours")
		return fmt.Errorf("failed to dispute hours: %w", err)
	}

	// Update registration hours status
	if err := s.registrationService.UpdateRegistrationHours(ctx, hoursLog.RegistrationID, hoursLog.Hours, string(models.HoursStatusDisputed)); err != nil {
		s.logger.WithContext(ctx).Error("Failed to update registration hours status")
		// Don't fail the operation, just log the error
	}

	// Send notification to coordinator
	if err := s.notificationService.SendHoursDisputedNotification(ctx, hoursLog.LoggedByUserID, hoursLog.RegistrationID, reason); err != nil {
		s.logger.WithContext(ctx).Error("Failed to send hours disputed notification")
		// Don't fail the operation, just log the error
	}

	s.logger.WithContext(ctx).
		WithField("hours_log_id", hoursLogID.String()).
		WithField("volunteer_profile_id", volunteerProfileID.String()).
		WithField("reason", reason).
		Info("Hours disputed successfully")

	return nil
}

// ResolveDispute resolves a disputed hours log (by coordinator)
func (s *hoursService) ResolveDispute(ctx context.Context, hoursLogID, coordinatorUserID uuid.UUID, resolutionNotes string) error {
	// Validate input
	if hoursLogID == uuid.Nil {
		return fmt.Errorf("%w: hours log ID is required", ErrInvalidHoursData)
	}
	if coordinatorUserID == uuid.Nil {
		return fmt.Errorf("%w: coordinator user ID is required", ErrInvalidHoursData)
	}
	if resolutionNotes == "" {
		return fmt.Errorf("%w: resolution notes are required", ErrInvalidHoursData)
	}

	// Get hours log
	hoursLog, err := s.hoursRepo.FindHoursLogByID(ctx, hoursLogID)
	if err != nil {
		return err
	}

	// Verify hours log is disputed
	if !hoursLog.IsDisputed() {
		return fmt.Errorf("can only resolve disputed hours logs")
	}

	// Verify coordinator is the one who logged the hours (authorization check)
	if hoursLog.LoggedByUserID != coordinatorUserID {
		s.logger.WithContext(ctx).Warn("Unauthorized attempt to resolve dispute")
		return ErrUnauthorized
	}

	// Get registration details
	registration, err := s.registrationService.GetRegistration(ctx, hoursLog.RegistrationID)
	if err != nil {
		return fmt.Errorf("failed to get registration: %w", err)
	}

	// Resolve dispute in repository
	resolvedAt := time.Now()
	if err := s.hoursRepo.ResolveDispute(ctx, hoursLogID, resolutionNotes, resolvedAt); err != nil {
		s.logger.WithContext(ctx).Error("Failed to resolve dispute")
		return fmt.Errorf("failed to resolve dispute: %w", err)
	}

	// Increment volunteer's total hours (now that it's verified)
	if err := s.volunteerService.IncrementTotalHours(ctx, registration.VolunteerProfileID, hoursLog.Hours); err != nil {
		s.logger.WithContext(ctx).Error("Failed to increment volunteer total hours")
		// Don't fail the operation, just log the error
	}

	// Update registration hours status
	if err := s.registrationService.UpdateRegistrationHours(ctx, hoursLog.RegistrationID, hoursLog.Hours, string(models.HoursStatusVerified)); err != nil {
		s.logger.WithContext(ctx).Error("Failed to update registration hours status")
		// Don't fail the operation, just log the error
	}

	// Send notification to volunteer
	if err := s.notificationService.SendDisputeResolvedNotification(ctx, registration.VolunteerProfileID, hoursLog.RegistrationID); err != nil {
		s.logger.WithContext(ctx).Error("Failed to send dispute resolved notification")
		// Don't fail the operation, just log the error
	}

	s.logger.WithContext(ctx).
		WithField("hours_log_id", hoursLogID.String()).
		WithField("coordinator_user_id", coordinatorUserID.String()).
		Info("Dispute resolved successfully")

	return nil
}

// AutoVerifyOldHours auto-verifies hours older than 7 days (FR-049)
// This should be called by a cron job daily
func (s *hoursService) AutoVerifyOldHours(ctx context.Context) (int, error) {
	// Calculate cutoff date (7 days ago)
	cutoffDate := time.Now().Add(-7 * 24 * time.Hour)

	// Find all pending hours logs older than 7 days
	oldHoursLogs, err := s.hoursRepo.FindPendingHoursOlderThan7Days(ctx, cutoffDate)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to find pending hours older than 7 days")
		return 0, fmt.Errorf("failed to find old pending hours: %w", err)
	}

	if len(oldHoursLogs) == 0 {
		s.logger.WithContext(ctx).Info("No pending hours logs older than 7 days found")
		return 0, nil
	}

	autoVerifiedCount := 0
	autoVerifiedAt := time.Now()

	// Auto-verify each old hours log
	for _, hoursLog := range oldHoursLogs {
		// Auto-verify hours
		if err := s.hoursRepo.AutoVerifyHours(ctx, hoursLog.ID, autoVerifiedAt); err != nil {
			s.logger.WithContext(ctx).
				WithField("hours_log_id", hoursLog.ID.String()).
				Error("Failed to auto-verify hours log")
			continue // Continue with next log, don't fail the entire operation
		}

		// Get registration details to get volunteer profile ID
		registration, err := s.registrationService.GetRegistration(ctx, hoursLog.RegistrationID)
		if err != nil {
			s.logger.WithContext(ctx).
				WithField("registration_id", hoursLog.RegistrationID.String()).
				Error("Failed to get registration for auto-verified hours")
			continue
		}

		// Increment volunteer's total hours
		if err := s.volunteerService.IncrementTotalHours(ctx, registration.VolunteerProfileID, hoursLog.Hours); err != nil {
			s.logger.WithContext(ctx).
				WithField("volunteer_profile_id", registration.VolunteerProfileID.String()).
				Error("Failed to increment volunteer total hours")
			// Continue anyway
		}

		// Update registration hours status
		if err := s.registrationService.UpdateRegistrationHours(ctx, hoursLog.RegistrationID, hoursLog.Hours, string(models.HoursStatusVerified)); err != nil {
			s.logger.WithContext(ctx).Error("Failed to update registration hours status")
			// Continue anyway
		}

		autoVerifiedCount++
	}

	s.logger.WithContext(ctx).
		WithField("auto_verified_count", autoVerifiedCount).
		WithField("total_found", len(oldHoursLogs)).
		Info("Auto-verification completed")

	return autoVerifiedCount, nil
}

// GetHoursLog retrieves a specific hours log by ID
func (s *hoursService) GetHoursLog(ctx context.Context, hoursLogID uuid.UUID) (*models.HoursLog, error) {
	if hoursLogID == uuid.Nil {
		return nil, fmt.Errorf("%w: hours log ID is required", ErrInvalidHoursData)
	}

	return s.hoursRepo.FindHoursLogByID(ctx, hoursLogID)
}

// GetHoursLogsByRegistration retrieves all hours logs for a registration
func (s *hoursService) GetHoursLogsByRegistration(ctx context.Context, registrationID uuid.UUID) ([]*models.HoursLog, error) {
	if registrationID == uuid.Nil {
		return nil, fmt.Errorf("%w: registration ID is required", ErrInvalidHoursData)
	}

	return s.hoursRepo.FindHoursLogsByRegistration(ctx, registrationID)
}

// GetHoursLogsByVolunteer retrieves all hours logs for a volunteer
func (s *hoursService) GetHoursLogsByVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.HoursLog, error) {
	if volunteerProfileID == uuid.Nil {
		return nil, fmt.Errorf("%w: volunteer profile ID is required", ErrInvalidHoursData)
	}

	return s.hoursRepo.FindHoursLogsByVolunteer(ctx, volunteerProfileID)
}

// GetPendingHoursForVolunteer retrieves pending hours logs for a volunteer
func (s *hoursService) GetPendingHoursForVolunteer(ctx context.Context, volunteerProfileID uuid.UUID) ([]*models.HoursLog, error) {
	if volunteerProfileID == uuid.Nil {
		return nil, fmt.Errorf("%w: volunteer profile ID is required", ErrInvalidHoursData)
	}

	// Get all hours logs for volunteer
	allLogs, err := s.hoursRepo.FindHoursLogsByVolunteer(ctx, volunteerProfileID)
	if err != nil {
		return nil, err
	}

	// Filter for pending status
	pendingLogs := make([]*models.HoursLog, 0)
	for _, log := range allLogs {
		if log.IsPending() {
			pendingLogs = append(pendingLogs, log)
		}
	}

	return pendingLogs, nil
}
