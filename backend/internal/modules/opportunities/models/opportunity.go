package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OpportunityStatus represents the status of an opportunity
type OpportunityStatus string

const (
	// OpportunityStatusDraft represents a draft opportunity not yet published
	OpportunityStatusDraft OpportunityStatus = "draft"
	// OpportunityStatusPublished represents a published opportunity visible to volunteers
	OpportunityStatusPublished OpportunityStatus = "published"
	// OpportunityStatusCancelled represents a cancelled opportunity
	OpportunityStatusCancelled OpportunityStatus = "cancelled"
	// OpportunityStatusCompleted represents a completed opportunity
	OpportunityStatusCompleted OpportunityStatus = "completed"
)

// RecurrencePattern represents the recurrence pattern for opportunities
type RecurrencePattern string

const (
	// RecurrencePatternDaily represents daily recurrence
	RecurrencePatternDaily RecurrencePattern = "daily"
	// RecurrencePatternWeekly represents weekly recurrence
	RecurrencePatternWeekly RecurrencePattern = "weekly"
	// RecurrencePatternMonthly represents monthly recurrence
	RecurrencePatternMonthly RecurrencePattern = "monthly"
	// RecurrencePatternCustom represents custom recurrence
	RecurrencePatternCustom RecurrencePattern = "custom"
)

// Opportunity represents a volunteer event or shift
// Opportunities are posted by organizations and can be one-time or recurring events
type Opportunity struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrganizationID       uuid.UUID          `gorm:"type:uuid;not null;index" json:"organization_id"`
	CreatedByUserID      uuid.UUID          `gorm:"type:uuid;not null;index" json:"created_by_user_id"`
	Title                string             `gorm:"type:varchar(200);not null;index" json:"title"`
	Description          string             `gorm:"type:text;not null" json:"description"`
	Status               OpportunityStatus  `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`
	StartDate            time.Time          `gorm:"type:date;not null;index:idx_opportunity_dates" json:"start_date"`
	StartTime            time.Time          `gorm:"type:time;not null" json:"start_time"`
	EndDate              time.Time          `gorm:"type:date;not null" json:"end_date"`
	EndTime              time.Time          `gorm:"type:time;not null" json:"end_time"`
	Timezone             string             `gorm:"type:varchar(50);not null" json:"timezone"`
	AddressLine1         string             `gorm:"type:varchar(255);not null" json:"address_line_1"`
	AddressLine2         *string            `gorm:"type:varchar(255)" json:"address_line_2,omitempty"`
	City                 string             `gorm:"type:varchar(100);not null;index:idx_opportunity_location" json:"city"`
	State                string             `gorm:"type:varchar(50);not null;index:idx_opportunity_location" json:"state"`
	PostalCode           string             `gorm:"type:varchar(20);not null" json:"postal_code"`
	Country              string             `gorm:"type:varchar(100);not null;default:'United States'" json:"country"`
	Latitude             *float64           `gorm:"type:decimal(10,7);index:idx_opportunity_geo" json:"latitude,omitempty"`
	Longitude            *float64           `gorm:"type:decimal(10,7);index:idx_opportunity_geo" json:"longitude,omitempty"`
	Capacity             int                `gorm:"not null" json:"capacity"`
	CurrentRegistrations int                `gorm:"not null;default:0" json:"current_registrations"`
	MinAge               *int               `gorm:"type:int" json:"min_age,omitempty"`
	IsRecurring          bool               `gorm:"not null;default:false" json:"is_recurring"`
	RecurrencePattern    *RecurrencePattern `gorm:"type:varchar(50)" json:"recurrence_pattern,omitempty"`
	RecurrenceEndDate    *time.Time         `gorm:"type:date" json:"recurrence_end_date,omitempty"`
	ParentOpportunityID  *uuid.UUID         `gorm:"type:uuid;index" json:"parent_opportunity_id,omitempty"`
	PublishedAt          *time.Time         `gorm:"type:timestamp" json:"published_at,omitempty"`
	CancelledAt          *time.Time         `gorm:"type:timestamp" json:"cancelled_at,omitempty"`
	CancellationReason   *string            `gorm:"type:text" json:"cancellation_reason,omitempty"`
	CompletedAt          *time.Time         `gorm:"type:timestamp" json:"completed_at,omitempty"`
	AutoCompleteAt       *time.Time         `gorm:"type:timestamp" json:"auto_complete_at,omitempty"`
	CreatedAt            time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt            time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt            gorm.DeletedAt     `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for the Opportunity model
func (Opportunity) TableName() string {
	return "opportunities"
}

// BeforeCreate hook to generate UUID and set auto-complete date
func (o *Opportunity) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}

	// Set auto-complete date to 7 days after end date
	if o.AutoCompleteAt == nil && !o.EndDate.IsZero() {
		autoCompleteDate := o.EndDate.AddDate(0, 0, 7)
		o.AutoCompleteAt = &autoCompleteDate
	}

	// Set published_at if status is published and not already set
	if o.Status == OpportunityStatusPublished && o.PublishedAt == nil {
		now := time.Now()
		o.PublishedAt = &now
	}

	return nil
}

// BeforeUpdate hook to handle status changes and validate
func (o *Opportunity) BeforeUpdate(tx *gorm.DB) error {
	// Set published_at when status changes to published
	if tx.Statement.Changed("Status") {
		if o.Status == OpportunityStatusPublished && o.PublishedAt == nil {
			now := time.Now()
			o.PublishedAt = &now
		}
		if o.Status == OpportunityStatusCancelled && o.CancelledAt == nil {
			now := time.Now()
			o.CancelledAt = &now
		}
		if o.Status == OpportunityStatusCompleted && o.CompletedAt == nil {
			now := time.Now()
			o.CompletedAt = &now
		}
	}

	// Update auto-complete date if end date changes
	if tx.Statement.Changed("EndDate") && !o.EndDate.IsZero() {
		autoCompleteDate := o.EndDate.AddDate(0, 0, 7)
		o.AutoCompleteAt = &autoCompleteDate
	}

	return nil
}

// Validate checks if the opportunity data is valid
func (o *Opportunity) Validate() error {
	// Title validation
	if o.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(o.Title) > 200 {
		return fmt.Errorf("title must be 200 characters or less")
	}

	// Description validation
	if o.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Date/time validation
	if o.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if o.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if o.EndDate.Before(o.StartDate) {
		return fmt.Errorf("end date must be on or after start date")
	}

	// Location validation
	if o.AddressLine1 == "" {
		return fmt.Errorf("address line 1 is required")
	}
	if o.City == "" {
		return fmt.Errorf("city is required")
	}
	if o.State == "" {
		return fmt.Errorf("state is required")
	}
	if o.PostalCode == "" {
		return fmt.Errorf("postal code is required")
	}

	// Capacity validation
	if o.Capacity <= 0 {
		return fmt.Errorf("capacity must be greater than 0")
	}

	// Timezone validation
	if o.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}

	// Recurrence validation
	if o.IsRecurring {
		if o.RecurrencePattern == nil {
			return fmt.Errorf("recurrence pattern is required for recurring opportunities")
		}
		if o.RecurrenceEndDate == nil {
			return fmt.Errorf("recurrence end date is required for recurring opportunities")
		}
		if o.RecurrenceEndDate.Before(o.StartDate) {
			return fmt.Errorf("recurrence end date must be on or after start date")
		}
	}

	// Min age validation
	if o.MinAge != nil && *o.MinAge < 0 {
		return fmt.Errorf("minimum age cannot be negative")
	}

	return nil
}

// IsFull returns true if the opportunity is at capacity
func (o *Opportunity) IsFull() bool {
	return o.CurrentRegistrations >= o.Capacity
}

// IsPastEvent returns true if the opportunity start date is in the past
func (o *Opportunity) IsPastEvent() bool {
	return o.StartDate.Before(time.Now())
}

// CanEdit returns true if the opportunity can be edited
// Past events cannot be edited except for hour logging
func (o *Opportunity) CanEdit() bool {
	return !o.IsPastEvent() || o.Status == OpportunityStatusDraft
}

// CanRegister returns true if volunteers can register for this opportunity
func (o *Opportunity) CanRegister() bool {
	return o.Status == OpportunityStatusPublished && !o.IsFull() && !o.IsPastEvent()
}

// ShouldAutoComplete returns true if the opportunity should be auto-completed
func (o *Opportunity) ShouldAutoComplete() bool {
	return o.Status == OpportunityStatusPublished &&
		o.AutoCompleteAt != nil &&
		time.Now().After(*o.AutoCompleteAt)
}
