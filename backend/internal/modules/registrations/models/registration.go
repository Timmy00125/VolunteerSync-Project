package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegistrationStatus represents the status of a volunteer registration
type RegistrationStatus string

const (
	// RegistrationStatusConfirmed represents a confirmed registration
	RegistrationStatusConfirmed RegistrationStatus = "confirmed"
	// RegistrationStatusCancelled represents a cancelled registration
	RegistrationStatusCancelled RegistrationStatus = "cancelled"
	// RegistrationStatusWaitlisted represents a waitlisted registration (opportunity at capacity)
	RegistrationStatusWaitlisted RegistrationStatus = "waitlisted"
	// RegistrationStatusCompleted represents a completed registration (event finished)
	RegistrationStatusCompleted RegistrationStatus = "completed"
)

// HoursStatus represents the status of logged volunteer hours
type HoursStatus string

const (
	// HoursStatusPending represents hours logged by coordinator, awaiting volunteer verification
	HoursStatusPending HoursStatus = "pending"
	// HoursStatusVerified represents hours verified by volunteer
	HoursStatusVerified HoursStatus = "verified"
	// HoursStatusDisputed represents hours disputed by volunteer
	HoursStatusDisputed HoursStatus = "disputed"
)

// Registration represents a volunteer's sign-up for a specific opportunity
// This is the junction table linking volunteers to opportunities (events)
type Registration struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OpportunityID      uuid.UUID          `gorm:"type:uuid;not null;index:idx_registrations_opportunity" json:"opportunity_id"`
	VolunteerProfileID uuid.UUID          `gorm:"type:uuid;not null;index:idx_registrations_volunteer" json:"volunteer_profile_id"`
	Status             RegistrationStatus `gorm:"type:varchar(20);not null;default:'confirmed';index" json:"status"`
	RegisteredAt       time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"registered_at"`
	CheckedInAt        *time.Time         `gorm:"type:timestamp" json:"checked_in_at,omitempty"`
	CancelledAt        *time.Time         `gorm:"type:timestamp" json:"cancelled_at,omitempty"`
	CancellationReason *string            `gorm:"type:text" json:"cancellation_reason,omitempty"`
	HoursWorked        *float64           `gorm:"type:decimal(5,2)" json:"hours_worked,omitempty"`
	HoursStatus        *HoursStatus       `gorm:"type:varchar(20)" json:"hours_status,omitempty"`
	HoursLoggedAt      *time.Time         `gorm:"type:timestamp" json:"hours_logged_at,omitempty"`
	HoursVerifiedAt    *time.Time         `gorm:"type:timestamp" json:"hours_verified_at,omitempty"`
	VolunteerRating    *int               `gorm:"type:int" json:"volunteer_rating,omitempty"`
	VolunteerReview    *string            `gorm:"type:text" json:"volunteer_review,omitempty"`
	ReviewSubmittedAt  *time.Time         `gorm:"type:timestamp" json:"review_submitted_at,omitempty"`
	CoordinatorNotes   *string            `gorm:"type:text" json:"coordinator_notes,omitempty"`
	CreatedAt          time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for the Registration model
func (Registration) TableName() string {
	return "registrations"
}

// BeforeCreate hook to generate UUID if not provided
func (r *Registration) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.RegisteredAt.IsZero() {
		r.RegisteredAt = time.Now()
	}
	return nil
}

// IsConfirmed checks if the registration is confirmed
func (r *Registration) IsConfirmed() bool {
	return r.Status == RegistrationStatusConfirmed
}

// IsWaitlisted checks if the registration is waitlisted
func (r *Registration) IsWaitlisted() bool {
	return r.Status == RegistrationStatusWaitlisted
}

// IsCancelled checks if the registration is cancelled
func (r *Registration) IsCancelled() bool {
	return r.Status == RegistrationStatusCancelled
}

// IsCompleted checks if the registration is completed
func (r *Registration) IsCompleted() bool {
	return r.Status == RegistrationStatusCompleted
}

// HasCheckedIn checks if the volunteer has checked in
func (r *Registration) HasCheckedIn() bool {
	return r.CheckedInAt != nil
}

// HasHoursLogged checks if hours have been logged for this registration
func (r *Registration) HasHoursLogged() bool {
	return r.HoursWorked != nil && *r.HoursWorked > 0
}

// AreHoursVerified checks if logged hours have been verified by the volunteer
func (r *Registration) AreHoursVerified() bool {
	return r.HoursStatus != nil && *r.HoursStatus == HoursStatusVerified
}

// AreHoursDisputed checks if logged hours have been disputed by the volunteer
func (r *Registration) AreHoursDisputed() bool {
	return r.HoursStatus != nil && *r.HoursStatus == HoursStatusDisputed
}

// AreHoursPending checks if logged hours are pending verification
func (r *Registration) AreHoursPending() bool {
	return r.HoursStatus != nil && *r.HoursStatus == HoursStatusPending
}

// HasReview checks if the volunteer has submitted a review
func (r *Registration) HasReview() bool {
	return r.ReviewSubmittedAt != nil
}

// CanSubmitReview checks if the volunteer can submit a review
// Review only allowed if status = completed and volunteer attended
func (r *Registration) CanSubmitReview() bool {
	return r.IsCompleted() && r.HasCheckedIn()
}

// IsLateCancellation checks if the cancellation is within 24 hours of the event
// This requires the event start time to be passed in
func (r *Registration) IsLateCancellation(eventStartTime time.Time) bool {
	if r.CancelledAt == nil {
		return false
	}
	hoursUntilEvent := eventStartTime.Sub(*r.CancelledAt).Hours()
	return hoursUntilEvent < 24 && hoursUntilEvent > 0
}
