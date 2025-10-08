package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// HoursLog represents a detailed audit trail of volunteer hours worked and verification
// This is an immutable audit log - no deletes, only status updates (FR-054)
// Follows the principle that logged hours should never be deleted for audit trail compliance
type HoursLog struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RegistrationID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_hours_registration" json:"registration_id"`
	Hours            float64        `gorm:"type:decimal(5,2);not null" json:"hours"`
	LoggedByUserID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_hours_logged_by" json:"logged_by_user_id"`
	Status           HoursStatus    `gorm:"type:varchar(20);not null;index" json:"status"`
	CoordinatorNotes *string        `gorm:"type:text" json:"coordinator_notes,omitempty"`
	VolunteerNotes   *string        `gorm:"type:text" json:"volunteer_notes,omitempty"`
	DisputeReason    *string        `gorm:"type:text" json:"dispute_reason,omitempty"`
	DisputedAt       *time.Time     `gorm:"type:timestamp" json:"disputed_at,omitempty"`
	ResolvedAt       *time.Time     `gorm:"type:timestamp" json:"resolved_at,omitempty"`
	ResolutionNotes  *string        `gorm:"type:text" json:"resolution_notes,omitempty"`
	LoggedAt         time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"logged_at"`
	VerifiedAt       *time.Time     `gorm:"type:timestamp" json:"verified_at,omitempty"`
	AutoVerifiedAt   *time.Time     `gorm:"type:timestamp" json:"auto_verified_at,omitempty"`
	CreatedAt        time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support (though should rarely be used per FR-054)
}

// TableName specifies the table name for the HoursLog model
func (HoursLog) TableName() string {
	return "hours_logs"
}

// BeforeCreate hook to generate UUID if not provided and set initial timestamps
func (h *HoursLog) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	if h.LoggedAt.IsZero() {
		h.LoggedAt = time.Now()
	}
	// Default status to pending if not set
	if h.Status == "" {
		h.Status = HoursStatusPending
	}
	return nil
}

// IsPending checks if the hours log is in pending status
func (h *HoursLog) IsPending() bool {
	return h.Status == HoursStatusPending
}

// IsVerified checks if the hours log is verified
func (h *HoursLog) IsVerified() bool {
	return h.Status == HoursStatusVerified
}

// IsDisputed checks if the hours log is disputed
func (h *HoursLog) IsDisputed() bool {
	return h.Status == HoursStatusDisputed
}

// IsAutoVerified checks if the hours were auto-verified after 7 days
func (h *HoursLog) IsAutoVerified() bool {
	return h.AutoVerifiedAt != nil
}

// IsOlderThan7Days checks if the hours log has been pending for more than 7 days
// Used by the auto-verification cron job (FR-049)
func (h *HoursLog) IsOlderThan7Days() bool {
	return h.IsPending() && time.Since(h.LoggedAt) > 7*24*time.Hour
}

// CanBeVerified checks if the hours log can be verified by the volunteer
func (h *HoursLog) CanBeVerified() bool {
	return h.IsPending()
}

// CanBeDisputed checks if the hours log can be disputed by the volunteer
func (h *HoursLog) CanBeDisputed() bool {
	return h.IsPending()
}

// IsResolved checks if a disputed hours log has been resolved
func (h *HoursLog) IsResolved() bool {
	return h.IsDisputed() && h.ResolvedAt != nil
}

// ValidateHours validates that hours are positive (reject negative values per FR edge case)
func (h *HoursLog) ValidateHours() error {
	if h.Hours <= 0 {
		return ErrInvalidHours
	}
	return nil
}

// Custom errors for hours log validation
var (
	// ErrInvalidHours is returned when hours are zero or negative
	ErrInvalidHours = gorm.ErrInvalidData
)
