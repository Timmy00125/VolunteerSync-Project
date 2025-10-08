package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PreferredTime represents the preferred time of day for volunteering
type PreferredTime string

const (
	// PreferredTimeMorning represents morning time preference
	PreferredTimeMorning PreferredTime = "morning"
	// PreferredTimeAfternoon represents afternoon time preference
	PreferredTimeAfternoon PreferredTime = "afternoon"
	// PreferredTimeEvening represents evening time preference
	PreferredTimeEvening PreferredTime = "evening"
	// PreferredTimeFlexible represents flexible time preference
	PreferredTimeFlexible PreferredTime = "flexible"
)

// VolunteerProfile represents extended profile data for volunteer users
// This includes availability preferences, skills, interests, and privacy settings
type VolunteerProfile struct {
	ID                       uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID                   uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	ProfilePhotoURL          *string        `gorm:"type:varchar(500)" json:"profile_photo_url,omitempty"`
	Biography                *string        `gorm:"type:text" json:"biography,omitempty"`
	Location                 *string        `gorm:"type:varchar(255);index" json:"location,omitempty"`
	Latitude                 *float64       `gorm:"type:decimal(10,7);index:idx_volunteer_profiles_lat_lng" json:"latitude,omitempty"`
	Longitude                *float64       `gorm:"type:decimal(10,7);index:idx_volunteer_profiles_lat_lng" json:"longitude,omitempty"`
	AvailabilityMonday       bool           `gorm:"not null;default:false" json:"availability_monday"`
	AvailabilityTuesday      bool           `gorm:"not null;default:false" json:"availability_tuesday"`
	AvailabilityWednesday    bool           `gorm:"not null;default:false" json:"availability_wednesday"`
	AvailabilityThursday     bool           `gorm:"not null;default:false" json:"availability_thursday"`
	AvailabilityFriday       bool           `gorm:"not null;default:false" json:"availability_friday"`
	AvailabilitySaturday     bool           `gorm:"not null;default:false" json:"availability_saturday"`
	AvailabilitySunday       bool           `gorm:"not null;default:false" json:"availability_sunday"`
	PreferredTime            *PreferredTime `gorm:"type:varchar(20)" json:"preferred_time,omitempty"`
	TotalHours               float64        `gorm:"type:decimal(10,2);not null;default:0" json:"total_hours"`
	TotalEvents              int            `gorm:"not null;default:0" json:"total_events"`
	EmergencyContactName     *string        `gorm:"type:varchar(200)" json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone    *string        `gorm:"type:varchar(20)" json:"emergency_contact_phone,omitempty"`
	PrivacyShowHours         bool           `gorm:"not null;default:true" json:"privacy_show_hours"`
	PrivacyShowEvents        bool           `gorm:"not null;default:true" json:"privacy_show_events"`
	PrivacyShowOrganizations bool           `gorm:"not null;default:true" json:"privacy_show_organizations"`
	NotificationInApp        bool           `gorm:"not null;default:true" json:"notification_in_app"`
	NotificationBrowserPush  bool           `gorm:"not null;default:false" json:"notification_browser_push"`
	CreatedAt                time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt                time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for the VolunteerProfile model
func (VolunteerProfile) TableName() string {
	return "volunteer_profiles"
}

// BeforeCreate hook to generate UUID if not provided
func (vp *VolunteerProfile) BeforeCreate(tx *gorm.DB) error {
	if vp.ID == uuid.Nil {
		vp.ID = uuid.New()
	}
	return nil
}

// GetAvailabilityMap returns a map of day names to their availability status
// Useful for displaying availability in a structured format
func (vp *VolunteerProfile) GetAvailabilityMap() map[string]bool {
	return map[string]bool{
		"monday":    vp.AvailabilityMonday,
		"tuesday":   vp.AvailabilityTuesday,
		"wednesday": vp.AvailabilityWednesday,
		"thursday":  vp.AvailabilityThursday,
		"friday":    vp.AvailabilityFriday,
		"saturday":  vp.AvailabilitySaturday,
		"sunday":    vp.AvailabilitySunday,
	}
}

// SetAvailability sets availability for all days from a map
// Useful for bulk updates of availability
func (vp *VolunteerProfile) SetAvailability(availability map[string]bool) {
	if monday, ok := availability["monday"]; ok {
		vp.AvailabilityMonday = monday
	}
	if tuesday, ok := availability["tuesday"]; ok {
		vp.AvailabilityTuesday = tuesday
	}
	if wednesday, ok := availability["wednesday"]; ok {
		vp.AvailabilityWednesday = wednesday
	}
	if thursday, ok := availability["thursday"]; ok {
		vp.AvailabilityThursday = thursday
	}
	if friday, ok := availability["friday"]; ok {
		vp.AvailabilityFriday = friday
	}
	if saturday, ok := availability["saturday"]; ok {
		vp.AvailabilitySaturday = saturday
	}
	if sunday, ok := availability["sunday"]; ok {
		vp.AvailabilitySunday = sunday
	}
}

// IsAvailableOnDay checks if the volunteer is available on a specific day
// Returns true if available, false otherwise
func (vp *VolunteerProfile) IsAvailableOnDay(day string) bool {
	switch day {
	case "monday":
		return vp.AvailabilityMonday
	case "tuesday":
		return vp.AvailabilityTuesday
	case "wednesday":
		return vp.AvailabilityWednesday
	case "thursday":
		return vp.AvailabilityThursday
	case "friday":
		return vp.AvailabilityFriday
	case "saturday":
		return vp.AvailabilitySaturday
	case "sunday":
		return vp.AvailabilitySunday
	default:
		return false
	}
}

// HasLocation checks if the volunteer has set their location with coordinates
func (vp *VolunteerProfile) HasLocation() bool {
	return vp.Latitude != nil && vp.Longitude != nil
}

// IncrementHours adds hours to the volunteer's total hours
// Used when hours are verified
func (vp *VolunteerProfile) IncrementHours(hours float64) {
	vp.TotalHours += hours
}

// IncrementEvents increments the total number of events attended
func (vp *VolunteerProfile) IncrementEvents() {
	vp.TotalEvents++
}
