package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BadgeType represents the type of achievement badge
type BadgeType string

const (
	// BadgeTypeSystem represents a system-wide badge awarded automatically
	BadgeTypeSystem BadgeType = "system"
	// BadgeTypeOrganizationCustom represents a custom badge created by an organization
	BadgeTypeOrganizationCustom BadgeType = "organization_custom"
)

// CriteriaType represents the criteria for earning a badge
type CriteriaType string

const (
	// CriteriaTypeHoursMilestone represents badges earned based on total hours (e.g., 10 hours, 50 hours)
	CriteriaTypeHoursMilestone CriteriaType = "hours_milestone"
	// CriteriaTypeEventsMilestone represents badges earned based on number of events (e.g., first event, 10 events)
	CriteriaTypeEventsMilestone CriteriaType = "events_milestone"
	// CriteriaTypeConsistency represents badges earned for consistent volunteering (e.g., 3 months consistent)
	CriteriaTypeConsistency CriteriaType = "consistency"
	// CriteriaTypeCustom represents custom criteria for organization-specific badges
	CriteriaTypeCustom CriteriaType = "custom"
)

// Achievement represents a recognition badge for volunteer milestones and custom organization awards
// System badges are awarded automatically based on criteria (FR-073)
// Organization custom badges can be awarded manually by coordinators (FR-075)
type Achievement struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid;index" json:"organization_id,omitempty"` // NULL for system badges
	Name           string         `gorm:"type:varchar(100);not null" json:"name"`
	Description    string         `gorm:"type:text;not null" json:"description"`
	IconURL        *string        `gorm:"type:varchar(500)" json:"icon_url,omitempty"`
	BadgeType      BadgeType      `gorm:"type:varchar(50);not null;index" json:"badge_type"`
	CriteriaType   *CriteriaType  `gorm:"type:varchar(50)" json:"criteria_type,omitempty"`
	CriteriaValue  *int           `gorm:"type:integer" json:"criteria_value,omitempty"` // Threshold value (e.g., 50 for 50 hours)
	CreatedAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support

	// Relationships
	VolunteerAchievements []VolunteerAchievement `gorm:"foreignKey:AchievementID" json:"volunteer_achievements,omitempty"`
}

// TableName specifies the table name for the Achievement model
func (Achievement) TableName() string {
	return "achievements"
}

// BeforeCreate hook to generate UUID if not provided
func (a *Achievement) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsSystemBadge checks if this is a system badge (not organization-specific)
func (a *Achievement) IsSystemBadge() bool {
	return a.BadgeType == BadgeTypeSystem
}

// IsOrganizationBadge checks if this is an organization-specific custom badge
func (a *Achievement) IsOrganizationBadge() bool {
	return a.BadgeType == BadgeTypeOrganizationCustom
}

// HasAutoCriteria checks if this badge has automatic awarding criteria
func (a *Achievement) HasAutoCriteria() bool {
	return a.CriteriaType != nil && *a.CriteriaType != CriteriaTypeCustom
}

// VolunteerAchievement represents a junction table tracking which volunteers earned which badges
// Badge cannot be earned twice by same volunteer (unique constraint on volunteer_profile_id + achievement_id)
// Congratulatory notification sent when badge earned (FR-076)
type VolunteerAchievement struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	VolunteerProfileID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_volunteer_achievement" json:"volunteer_profile_id"`
	AchievementID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_volunteer_achievement;index:idx_achievement" json:"achievement_id"`
	EarnedAt           time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"earned_at"`
	AwardedByUserID    *uuid.UUID `gorm:"type:uuid" json:"awarded_by_user_id,omitempty"` // User who awarded (for custom badges)
	CreatedAt          time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relationships
	Achievement Achievement `gorm:"foreignKey:AchievementID" json:"achievement,omitempty"`
}

// TableName specifies the table name for the VolunteerAchievement model
func (VolunteerAchievement) TableName() string {
	return "volunteer_achievements"
}

// BeforeCreate hook to generate UUID if not provided
func (va *VolunteerAchievement) BeforeCreate(tx *gorm.DB) error {
	if va.ID == uuid.Nil {
		va.ID = uuid.New()
	}
	return nil
}

// System Badge Definitions
// These are commonly used system badges that can be seeded into the database
var (
	// Hours Milestones
	SystemBadgeFirstHour = Achievement{
		Name:          "First Hour",
		Description:   "Completed your first hour of volunteering!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeHoursMilestone),
		CriteriaValue: ptrInt(1),
	}

	SystemBadge10Hours = Achievement{
		Name:          "10 Hours",
		Description:   "Volunteered for 10 hours - making a real impact!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeHoursMilestone),
		CriteriaValue: ptrInt(10),
	}

	SystemBadge25Hours = Achievement{
		Name:          "25 Hours",
		Description:   "Reached 25 hours of volunteering - you're amazing!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeHoursMilestone),
		CriteriaValue: ptrInt(25),
	}

	SystemBadge50Hours = Achievement{
		Name:          "50 Hours",
		Description:   "Half a century of hours! Your dedication is inspiring.",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeHoursMilestone),
		CriteriaValue: ptrInt(50),
	}

	SystemBadge100Hours = Achievement{
		Name:          "100 Hours",
		Description:   "100 hours of volunteering! You're a community hero.",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeHoursMilestone),
		CriteriaValue: ptrInt(100),
	}

	// Events Milestones
	SystemBadgeFirstEvent = Achievement{
		Name:          "First Event",
		Description:   "Attended your first volunteer event - welcome!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeEventsMilestone),
		CriteriaValue: ptrInt(1),
	}

	SystemBadge5Events = Achievement{
		Name:          "5 Events",
		Description:   "Participated in 5 volunteer events - you're committed!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeEventsMilestone),
		CriteriaValue: ptrInt(5),
	}

	SystemBadge10Events = Achievement{
		Name:          "10 Events",
		Description:   "10 events completed! You're a regular volunteer.",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeEventsMilestone),
		CriteriaValue: ptrInt(10),
	}

	SystemBadge25Events = Achievement{
		Name:          "25 Events",
		Description:   "25 events - you're truly making a difference!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeEventsMilestone),
		CriteriaValue: ptrInt(25),
	}

	// Consistency Badges
	SystemBadge3MonthsConsistent = Achievement{
		Name:          "3 Months Consistent",
		Description:   "Volunteered consistently for 3 months straight!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeConsistency),
		CriteriaValue: ptrInt(3), // 3 months
	}

	SystemBadge6MonthsConsistent = Achievement{
		Name:          "6 Months Consistent",
		Description:   "Half a year of consistent volunteering - incredible dedication!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeConsistency),
		CriteriaValue: ptrInt(6), // 6 months
	}

	SystemBadge1YearConsistent = Achievement{
		Name:          "1 Year Consistent",
		Description:   "One full year of consistent volunteering - you're a champion!",
		BadgeType:     BadgeTypeSystem,
		CriteriaType:  ptrCriteriaType(CriteriaTypeConsistency),
		CriteriaValue: ptrInt(12), // 12 months
	}
)

// GetSystemBadges returns all predefined system badges for seeding
func GetSystemBadges() []Achievement {
	return []Achievement{
		SystemBadgeFirstHour,
		SystemBadge10Hours,
		SystemBadge25Hours,
		SystemBadge50Hours,
		SystemBadge100Hours,
		SystemBadgeFirstEvent,
		SystemBadge5Events,
		SystemBadge10Events,
		SystemBadge25Events,
		SystemBadge3MonthsConsistent,
		SystemBadge6MonthsConsistent,
		SystemBadge1YearConsistent,
	}
}

// Helper functions to create pointers
func ptrCriteriaType(ct CriteriaType) *CriteriaType {
	return &ct
}

func ptrInt(i int) *int {
	return &i
}
