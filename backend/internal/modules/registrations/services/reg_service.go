package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/registrations/repositories"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/google/uuid"
)

// Custom errors for service operations
var (
	// ErrInvalidRegistrationData is returned when registration data is invalid
	ErrInvalidRegistrationData = errors.New("invalid registration data")
	// ErrRegistrationNotFound is returned when a registration cannot be found
	ErrRegistrationNotFound = repositories.ErrRegistrationNotFound
	// ErrRegistrationAlreadyExists is returned when a registration already exists
	ErrRegistrationAlreadyExists = repositories.ErrRegistrationAlreadyExists
	// ErrOpportunityAtCapacity is returned when an opportunity has reached its capacity
	ErrOpportunityAtCapacity = errors.New("opportunity has reached capacity")
	// ErrRegistrationOverlap is returned when a volunteer has an overlapping registration
	ErrRegistrationOverlap = errors.New("volunteer has an overlapping event registration")
	// ErrLateCancellation is returned when a cancellation is within 24 hours of the event
	ErrLateCancellation = errors.New("cancellation is within 24 hours of event start")
	// ErrUnauthorized is returned when a user doesn't have permission for an operation
	ErrUnauthorized = errors.New("unauthorized to perform this action")
	// ErrInvalidStatus is returned when a registration status transition is invalid
	ErrInvalidStatus = errors.New("invalid registration status transition")
)

// OpportunityService defines the interface for opportunity operations
// This will be injected from the opportunities module
type OpportunityService interface {
	GetOpportunity(ctx context.Context, id uuid.UUID) (*OpportunityDetails, error)
}

// OpportunityDetails represents opportunity data needed by registration service
type OpportunityDetails struct {
	ID                   uuid.UUID
	OrganizationID       uuid.UUID
	Title                string
	StartDate            time.Time
	StartTime            time.Time
	EndDate              time.Time
	EndTime              time.Time
	Capacity             int
	CurrentRegistrations int
	Status               string
	Location             string
	Timezone             string
}

// NotificationService defines the interface for sending notifications
// This will be implemented in the notifications module
type NotificationService interface {
	SendRegistrationConfirmation(ctx context.Context, volunteerID, opportunityID uuid.UUID) error
	SendCancellationNotification(ctx context.Context, volunteerID, opportunityID uuid.UUID) error
	SendLateCancellationWarning(ctx context.Context, volunteerID, opportunityID uuid.UUID) error
	SendWaitlistConfirmation(ctx context.Context, volunteerID, opportunityID uuid.UUID) error
	SendWaitlistPromotion(ctx context.Context, volunteerID, opportunityID uuid.UUID) error
}

// RegistrationService encapsulates registration business logic
// Provides methods for volunteer event registration, cancellation, check-in, and calendar generation
type RegistrationService interface {
	// RegisterVolunteer registers a volunteer for an opportunity
	// Checks capacity, duplicates, overlaps, and adds to waitlist if full
	RegisterVolunteer(ctx context.Context, input RegisterVolunteerInput) (*models.Registration, error)

	// CancelRegistration cancels a volunteer's registration
	// Issues late cancellation warning if within 24 hours
	CancelRegistration(ctx context.Context, registrationID, volunteerID uuid.UUID, reason *string) error

	// CheckInVolunteer checks in a volunteer for an event
	CheckInVolunteer(ctx context.Context, registrationID, coordinatorID uuid.UUID) error

	// GetRegistration retrieves a specific registration by ID
	GetRegistration(ctx context.Context, registrationID uuid.UUID) (*models.Registration, error)

	// ListVolunteerRegistrations lists all registrations for a volunteer
	ListVolunteerRegistrations(ctx context.Context, volunteerProfileID uuid.UUID, filters RegistrationFilters) ([]*models.Registration, error)

	// ListOpportunityRegistrations lists all registrations for an opportunity
	ListOpportunityRegistrations(ctx context.Context, opportunityID uuid.UUID, filters RegistrationFilters) ([]*models.Registration, error)

	// GenerateCalendarFile generates an .ics calendar file for a registration
	GenerateCalendarFile(ctx context.Context, registrationID uuid.UUID) ([]byte, error)

	// PromoteFromWaitlist promotes a waitlisted volunteer to confirmed status
	// Called when a confirmed volunteer cancels
	PromoteFromWaitlist(ctx context.Context, opportunityID uuid.UUID) error

	// UpdateHoursInformation updates the hours worked and status for a registration
	// Called by the hours service when hours are logged or verified
	UpdateHoursInformation(ctx context.Context, registrationID uuid.UUID, hours float64, hoursStatus string) error
}

// RegisterVolunteerInput represents the input for registering a volunteer
type RegisterVolunteerInput struct {
	OpportunityID      uuid.UUID
	VolunteerProfileID uuid.UUID
}

// RegistrationFilters represents filters for listing registrations
type RegistrationFilters struct {
	Status    *models.RegistrationStatus
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// registrationService is the implementation of RegistrationService
type registrationService struct {
	registrationRepo    repositories.RegistrationRepository
	opportunityService  OpportunityService
	notificationService NotificationService
	logger              logger.Logger
}

// NewRegistrationService creates a new instance of RegistrationService
func NewRegistrationService(
	registrationRepo repositories.RegistrationRepository,
	opportunityService OpportunityService,
	notificationService NotificationService,
	logger logger.Logger,
) RegistrationService {
	return &registrationService{
		registrationRepo:    registrationRepo,
		opportunityService:  opportunityService,
		notificationService: notificationService,
		logger:              logger,
	}
}

// RegisterVolunteer registers a volunteer for an opportunity
func (s *registrationService) RegisterVolunteer(ctx context.Context, input RegisterVolunteerInput) (*models.Registration, error) {
	// Validate input
	if input.OpportunityID == uuid.Nil {
		return nil, fmt.Errorf("%w: opportunity ID is required", ErrInvalidRegistrationData)
	}
	if input.VolunteerProfileID == uuid.Nil {
		return nil, fmt.Errorf("%w: volunteer profile ID is required", ErrInvalidRegistrationData)
	}

	// Get opportunity details
	opportunity, err := s.opportunityService.GetOpportunity(ctx, input.OpportunityID)
	if err != nil {
		s.logger.WithContext(ctx).WithField("opportunity_id", input.OpportunityID.String()).Error("Failed to get opportunity")
		return nil, fmt.Errorf("failed to get opportunity: %w", err)
	}

	// Check if opportunity is published and not cancelled/completed
	if opportunity.Status != "published" {
		return nil, fmt.Errorf("%w: opportunity is not available for registration", ErrInvalidRegistrationData)
	}

	// Check for duplicate registration
	existingReg, err := s.registrationRepo.FindRegistrationByVolunteerAndOpportunity(ctx, input.VolunteerProfileID, input.OpportunityID)
	if err != nil && !errors.Is(err, repositories.ErrRegistrationNotFound) {
		s.logger.WithContext(ctx).Error("Failed to check existing registration")
		return nil, fmt.Errorf("failed to check existing registration: %w", err)
	}
	if existingReg != nil && !existingReg.IsCancelled() {
		return nil, ErrRegistrationAlreadyExists
	}

	// Check for overlapping registrations
	overlaps, err := s.checkForOverlaps(ctx, input.VolunteerProfileID, opportunity.StartDate, opportunity.StartTime, opportunity.EndDate, opportunity.EndTime)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to check for overlapping registrations")
		return nil, fmt.Errorf("failed to check for overlaps: %w", err)
	}
	if len(overlaps) > 0 {
		s.logger.WithContext(ctx).Warn("Volunteer has overlapping registrations")
		return nil, ErrRegistrationOverlap
	}

	// Check capacity
	confirmedCount, err := s.registrationRepo.CountConfirmedRegistrationsByOpportunity(ctx, input.OpportunityID)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to count confirmed registrations")
		return nil, fmt.Errorf("failed to check capacity: %w", err)
	}

	// Determine initial status based on capacity
	initialStatus := models.RegistrationStatusConfirmed
	if int(confirmedCount) >= opportunity.Capacity {
		initialStatus = models.RegistrationStatusWaitlisted
		s.logger.WithContext(ctx).Info("Opportunity at capacity, adding to waitlist")
	}

	// Create registration
	registration := &models.Registration{
		OpportunityID:      input.OpportunityID,
		VolunteerProfileID: input.VolunteerProfileID,
		Status:             initialStatus,
		RegisteredAt:       time.Now(),
	}

	err = s.registrationRepo.CreateRegistration(ctx, registration)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to create registration")
		return nil, fmt.Errorf("failed to create registration: %w", err)
	}

	// Send appropriate notification
	if initialStatus == models.RegistrationStatusWaitlisted {
		if s.notificationService != nil {
			_ = s.notificationService.SendWaitlistConfirmation(ctx, input.VolunteerProfileID, input.OpportunityID)
		}
	} else {
		if s.notificationService != nil {
			_ = s.notificationService.SendRegistrationConfirmation(ctx, input.VolunteerProfileID, input.OpportunityID)
		}
	}

	s.logger.WithContext(ctx).Info("Volunteer registered successfully")
	return registration, nil
}

// CancelRegistration cancels a volunteer's registration
func (s *registrationService) CancelRegistration(ctx context.Context, registrationID, volunteerID uuid.UUID, reason *string) error {
	// Validate input
	if registrationID == uuid.Nil {
		return fmt.Errorf("%w: registration ID is required", ErrInvalidRegistrationData)
	}

	// Get registration
	registration, err := s.registrationRepo.FindRegistrationByID(ctx, registrationID)
	if err != nil {
		return err
	}

	// Verify ownership
	if registration.VolunteerProfileID != volunteerID {
		return ErrUnauthorized
	}

	// Check if already cancelled
	if registration.IsCancelled() {
		return fmt.Errorf("%w: registration is already cancelled", ErrInvalidStatus)
	}

	// Get opportunity details to check for late cancellation
	opportunity, err := s.opportunityService.GetOpportunity(ctx, registration.OpportunityID)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to get opportunity for cancellation check")
		return fmt.Errorf("failed to get opportunity: %w", err)
	}

	cancelledAt := time.Now()

	// Check for late cancellation (within 24 hours)
	isLate := registration.IsLateCancellation(opportunity.StartDate)
	if isLate {
		s.logger.WithContext(ctx).Warn("Late cancellation detected")
		if s.notificationService != nil {
			_ = s.notificationService.SendLateCancellationWarning(ctx, volunteerID, registration.OpportunityID)
		}
	}

	// Cancel the registration
	err = s.registrationRepo.CancelRegistration(ctx, registrationID, reason, cancelledAt)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to cancel registration")
		return fmt.Errorf("failed to cancel registration: %w", err)
	}

	// If it was a confirmed registration, try to promote from waitlist
	if registration.IsConfirmed() {
		_ = s.PromoteFromWaitlist(ctx, registration.OpportunityID)
	}

	s.logger.WithContext(ctx).Info("Registration cancelled successfully")
	return nil
}

// CheckInVolunteer checks in a volunteer for an event
func (s *registrationService) CheckInVolunteer(ctx context.Context, registrationID, coordinatorID uuid.UUID) error {
	// Validate input
	if registrationID == uuid.Nil {
		return fmt.Errorf("%w: registration ID is required", ErrInvalidRegistrationData)
	}

	// Get registration
	registration, err := s.registrationRepo.FindRegistrationByID(ctx, registrationID)
	if err != nil {
		return err
	}

	// Check if registration is confirmed
	if !registration.IsConfirmed() {
		return fmt.Errorf("%w: only confirmed registrations can be checked in", ErrInvalidStatus)
	}

	// Check if already checked in
	if registration.HasCheckedIn() {
		return fmt.Errorf("%w: volunteer is already checked in", ErrInvalidStatus)
	}

	// Perform check-in
	checkedInAt := time.Now()
	err = s.registrationRepo.CheckIn(ctx, registrationID, checkedInAt)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to check in volunteer")
		return fmt.Errorf("failed to check in: %w", err)
	}

	s.logger.WithContext(ctx).Info("Volunteer checked in successfully")
	return nil
}

// GetRegistration retrieves a specific registration by ID
func (s *registrationService) GetRegistration(ctx context.Context, registrationID uuid.UUID) (*models.Registration, error) {
	if registrationID == uuid.Nil {
		return nil, fmt.Errorf("%w: registration ID is required", ErrInvalidRegistrationData)
	}

	return s.registrationRepo.FindRegistrationByID(ctx, registrationID)
}

// ListVolunteerRegistrations lists all registrations for a volunteer
func (s *registrationService) ListVolunteerRegistrations(ctx context.Context, volunteerProfileID uuid.UUID, filters RegistrationFilters) ([]*models.Registration, error) {
	if volunteerProfileID == uuid.Nil {
		return nil, fmt.Errorf("%w: volunteer profile ID is required", ErrInvalidRegistrationData)
	}

	return s.registrationRepo.FindRegistrationsByVolunteer(ctx, volunteerProfileID)
}

// ListOpportunityRegistrations lists all registrations for an opportunity
func (s *registrationService) ListOpportunityRegistrations(ctx context.Context, opportunityID uuid.UUID, filters RegistrationFilters) ([]*models.Registration, error) {
	if opportunityID == uuid.Nil {
		return nil, fmt.Errorf("%w: opportunity ID is required", ErrInvalidRegistrationData)
	}

	return s.registrationRepo.FindRegistrationsByOpportunity(ctx, opportunityID)
}

// GenerateCalendarFile generates an .ics calendar file for a registration
func (s *registrationService) GenerateCalendarFile(ctx context.Context, registrationID uuid.UUID) ([]byte, error) {
	// Get registration
	registration, err := s.registrationRepo.FindRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	// Get opportunity details
	opportunity, err := s.opportunityService.GetOpportunity(ctx, registration.OpportunityID)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to get opportunity for calendar generation")
		return nil, fmt.Errorf("failed to get opportunity: %w", err)
	}

	// Generate iCalendar format
	icsContent := s.generateICalendarContent(registration, opportunity)

	s.logger.WithContext(ctx).Info("Calendar file generated")
	return []byte(icsContent), nil
}

// PromoteFromWaitlist promotes a waitlisted volunteer to confirmed status
func (s *registrationService) PromoteFromWaitlist(ctx context.Context, opportunityID uuid.UUID) error {
	// Get waitlisted registrations (ordered by registration date)
	waitlisted, err := s.registrationRepo.FindWaitlistedRegistrations(ctx, opportunityID)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to get waitlisted registrations")
		return err
	}

	if len(waitlisted) == 0 {
		// No one on waitlist
		return nil
	}

	// Get opportunity details to check capacity
	opportunity, err := s.opportunityService.GetOpportunity(ctx, opportunityID)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to get opportunity for waitlist promotion")
		return err
	}

	// Count current confirmed registrations
	confirmedCount, err := s.registrationRepo.CountConfirmedRegistrationsByOpportunity(ctx, opportunityID)
	if err != nil {
		s.logger.WithContext(ctx).Error("Failed to count confirmed registrations")
		return err
	}

	// Promote first waitlisted if there's space
	if int(confirmedCount) < opportunity.Capacity {
		firstWaitlisted := waitlisted[0]
		err = s.registrationRepo.UpdateRegistrationStatus(ctx, firstWaitlisted.ID, models.RegistrationStatusConfirmed)
		if err != nil {
			s.logger.WithContext(ctx).Error("Failed to promote from waitlist")
			return err
		}

		// Send promotion notification
		if s.notificationService != nil {
			_ = s.notificationService.SendWaitlistPromotion(ctx, firstWaitlisted.VolunteerProfileID, opportunityID)
		}

		s.logger.WithContext(ctx).Info("Volunteer promoted from waitlist")
	}

	return nil
}

// UpdateHoursInformation updates the hours worked and status for a registration
// Called by the hours service when hours are logged or verified
func (s *registrationService) UpdateHoursInformation(ctx context.Context, registrationID uuid.UUID, hours float64, hoursStatus string) error {
	// Validate input
	if registrationID == uuid.Nil {
		return fmt.Errorf("%w: registration ID is required", ErrInvalidRegistrationData)
	}

	if hours < 0 {
		return fmt.Errorf("%w: hours must be non-negative", ErrInvalidRegistrationData)
	}

	// Get registration to verify it exists
	registration, err := s.registrationRepo.FindRegistrationByID(ctx, registrationID)
	if err != nil {
		return err
	}

	// Parse hours status
	var status models.HoursStatus
	switch hoursStatus {
	case "pending":
		status = models.HoursStatusPending
	case "verified":
		status = models.HoursStatusVerified
	case "disputed":
		status = models.HoursStatusDisputed
	default:
		return fmt.Errorf("%w: invalid hours status '%s'", ErrInvalidRegistrationData, hoursStatus)
	}

	// Log the hours
	now := time.Now()
	if err := s.registrationRepo.LogHours(ctx, registrationID, hours, status, now); err != nil {
		s.logger.WithContext(ctx).WithField("registration_id", registrationID.String()).Error("Failed to log hours")
		return fmt.Errorf("failed to log hours: %w", err)
	}

	s.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"registration_id": registrationID.String(),
		"hours":           hours,
		"status":          hoursStatus,
	}).Info("Hours information updated on registration")

	// Send notification if hours are logged
	if hoursStatus == "pending" && s.notificationService != nil {
		_ = s.notificationService.SendRegistrationConfirmation(ctx, registration.VolunteerProfileID, registration.OpportunityID)
	}

	return nil
}

// checkForOverlaps checks if a volunteer has overlapping event registrations
func (s *registrationService) checkForOverlaps(ctx context.Context, volunteerProfileID uuid.UUID, startDate time.Time, startTime time.Time, endDate time.Time, endTime time.Time) ([]*models.Registration, error) {
	// Get all confirmed registrations for the volunteer
	registrations, err := s.registrationRepo.FindOverlappingRegistrations(ctx, volunteerProfileID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Filter for actual time overlaps (need to compare with opportunity times)
	// This is a simplified version - full implementation would need opportunity data
	var overlapping []*models.Registration
	for _, reg := range registrations {
		// Get opportunity details to check actual time overlap
		opp, err := s.opportunityService.GetOpportunity(ctx, reg.OpportunityID)
		if err != nil {
			continue // Skip if we can't get opportunity details
		}

		// Check if time ranges overlap
		if s.timeRangesOverlap(startDate, endDate, opp.StartDate, opp.EndDate) {
			overlapping = append(overlapping, reg)
		}
	}

	return overlapping, nil
}

// timeRangesOverlap checks if two time ranges overlap
func (s *registrationService) timeRangesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

// generateICalendarContent generates iCalendar (.ics) file content
func (s *registrationService) generateICalendarContent(registration *models.Registration, opportunity *OpportunityDetails) string {
	// Basic iCalendar format (RFC 5545)
	var ics strings.Builder

	ics.WriteString("BEGIN:VCALENDAR\r\n")
	ics.WriteString("VERSION:2.0\r\n")
	ics.WriteString("PRODID:-//VolunteerSync//Event Calendar//EN\r\n")
	ics.WriteString("CALSCALE:GREGORIAN\r\n")
	ics.WriteString("METHOD:PUBLISH\r\n")
	ics.WriteString("BEGIN:VEVENT\r\n")
	ics.WriteString(fmt.Sprintf("UID:%s@volunteersync.org\r\n", registration.ID.String()))
	ics.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z")))
	ics.WriteString(fmt.Sprintf("DTSTART:%s\r\n", opportunity.StartDate.UTC().Format("20060102T150405Z")))
	ics.WriteString(fmt.Sprintf("DTEND:%s\r\n", opportunity.EndDate.UTC().Format("20060102T150405Z")))
	ics.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(opportunity.Title)))
	ics.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICalText(opportunity.Location)))
	ics.WriteString(fmt.Sprintf("STATUS:%s\r\n", registration.Status))
	ics.WriteString("END:VEVENT\r\n")
	ics.WriteString("END:VCALENDAR\r\n")

	return ics.String()
}

// escapeICalText escapes special characters in iCalendar text fields
func escapeICalText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
}
