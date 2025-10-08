package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VerificationStatus represents the verification status of an organization
type VerificationStatus string

const (
	// VerificationStatusVerified represents a verified organization
	VerificationStatusVerified VerificationStatus = "verified"
	// VerificationStatusUnverified represents an unverified organization
	VerificationStatusUnverified VerificationStatus = "unverified"
)

// Organization represents a nonprofit organization or community group managing volunteer programs
// Organizations can post opportunities, manage teams, and track volunteer impact
type Organization struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name               string             `gorm:"type:varchar(200);not null;index" json:"name"`
	Slug               string             `gorm:"type:varchar(200);uniqueIndex;not null" json:"slug"`
	MissionStatement   *string            `gorm:"type:text" json:"mission_statement,omitempty"`
	Description        *string            `gorm:"type:text" json:"description,omitempty"`
	Website            *string            `gorm:"type:varchar(255)" json:"website,omitempty"`
	Email              string             `gorm:"type:varchar(255);not null" json:"email"`
	Phone              *string            `gorm:"type:varchar(20)" json:"phone,omitempty"`
	AddressLine1       *string            `gorm:"type:varchar(255)" json:"address_line_1,omitempty"`
	AddressLine2       *string            `gorm:"type:varchar(255)" json:"address_line_2,omitempty"`
	City               *string            `gorm:"type:varchar(100)" json:"city,omitempty"`
	State              *string            `gorm:"type:varchar(50)" json:"state,omitempty"`
	PostalCode         *string            `gorm:"type:varchar(20)" json:"postal_code,omitempty"`
	Country            string             `gorm:"type:varchar(100);not null;default:'United States'" json:"country"`
	Latitude           *float64           `gorm:"type:decimal(10,7)" json:"latitude,omitempty"`
	Longitude          *float64           `gorm:"type:decimal(10,7)" json:"longitude,omitempty"`
	LogoURL            *string            `gorm:"type:varchar(500)" json:"logo_url,omitempty"`
	BannerURL          *string            `gorm:"type:varchar(500)" json:"banner_url,omitempty"`
	VerificationStatus VerificationStatus `gorm:"type:varchar(20);not null;default:'verified'" json:"verification_status"`
	VerifiedAt         *time.Time         `gorm:"type:timestamp" json:"verified_at,omitempty"`
	TotalVolunteers    int                `gorm:"not null;default:0" json:"total_volunteers"`
	TotalHours         float64            `gorm:"type:decimal(10,2);not null;default:0" json:"total_hours"`
	AvgRating          *float64           `gorm:"type:decimal(3,2)" json:"avg_rating,omitempty"`
	CreatedAt          time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for the Organization model
func (Organization) TableName() string {
	return "organizations"
}

// BeforeCreate hook to generate UUID and slug if not provided
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}

	// Generate slug from name if not provided
	if o.Slug == "" {
		o.Slug = GenerateSlug(o.Name)
	}

	// Auto-verify organization on creation (FR-015: V1 auto-verifies all orgs)
	if o.VerificationStatus == "" {
		o.VerificationStatus = VerificationStatusVerified
		now := time.Now()
		o.VerifiedAt = &now
	}

	return nil
}

// BeforeUpdate hook to regenerate slug if name changes
func (o *Organization) BeforeUpdate(tx *gorm.DB) error {
	// Check if name is being updated
	if tx.Statement.Changed("Name") {
		o.Slug = GenerateSlug(o.Name)
	}
	return nil
}

// GenerateSlug creates a URL-friendly slug from the organization name
// Converts to lowercase, replaces spaces/special chars with hyphens, removes consecutive hyphens
// Example: "Hope Community Center" -> "hope-community-center"
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	// Keep only alphanumeric characters and hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// If slug is empty after processing, generate a random one
	if slug == "" {
		slug = fmt.Sprintf("org-%s", uuid.New().String()[:8])
	}

	return slug
}

// HasAddress returns true if the organization has a complete address
func (o *Organization) HasAddress() bool {
	return o.AddressLine1 != nil && *o.AddressLine1 != "" &&
		o.City != nil && *o.City != "" &&
		o.State != nil && *o.State != "" &&
		o.PostalCode != nil && *o.PostalCode != ""
}

// IsGeocodingRequired returns true if the organization has an address but no coordinates
func (o *Organization) IsGeocodingRequired() bool {
	return o.HasAddress() && (o.Latitude == nil || o.Longitude == nil)
}

// GetFullAddress returns the complete address as a single string for geocoding
func (o *Organization) GetFullAddress() string {
	if !o.HasAddress() {
		return ""
	}

	parts := []string{}

	if o.AddressLine1 != nil && *o.AddressLine1 != "" {
		parts = append(parts, *o.AddressLine1)
	}

	if o.AddressLine2 != nil && *o.AddressLine2 != "" {
		parts = append(parts, *o.AddressLine2)
	}

	if o.City != nil && *o.City != "" {
		parts = append(parts, *o.City)
	}

	if o.State != nil && *o.State != "" {
		parts = append(parts, *o.State)
	}

	if o.PostalCode != nil && *o.PostalCode != "" {
		parts = append(parts, *o.PostalCode)
	}

	parts = append(parts, o.Country)

	return strings.Join(parts, ", ")
}

// SetCoordinates sets the latitude and longitude for the organization
func (o *Organization) SetCoordinates(lat, lng float64) {
	o.Latitude = &lat
	o.Longitude = &lng
}

// IsVerified returns true if the organization is verified
func (o *Organization) IsVerified() bool {
	return o.VerificationStatus == VerificationStatusVerified
}

// Verify marks the organization as verified with the current timestamp
func (o *Organization) Verify() {
	o.VerificationStatus = VerificationStatusVerified
	now := time.Now()
	o.VerifiedAt = &now
}

// IncrementVolunteers increases the total volunteer count by the given amount
func (o *Organization) IncrementVolunteers(count int) {
	o.TotalVolunteers += count
}

// IncrementHours increases the total hours count by the given amount
func (o *Organization) IncrementHours(hours float64) {
	o.TotalHours += hours
}

// UpdateRating updates the average rating for the organization
func (o *Organization) UpdateRating(rating float64) {
	o.AvgRating = &rating
}
