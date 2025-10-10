package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrOrganizationNotFound is returned when an organization cannot be found
	ErrOrganizationNotFound = errors.New("organization not found")
	// ErrOrganizationAlreadyExists is returned when attempting to create an organization with an existing slug
	ErrOrganizationAlreadyExists = errors.New("organization with this name already exists")
	// ErrInvalidOrganizationID is returned when the provided organization ID is invalid
	ErrInvalidOrganizationID = errors.New("invalid organization ID")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// OrganizationFilters represents filters for listing organizations
type OrganizationFilters struct {
	Search   string      // Search by name
	City     string      // Filter by city
	State    string      // Filter by state
	CauseIDs []uuid.UUID // Filter by cause categories
	Page     int         // Page number (1-indexed)
	Limit    int         // Items per page
}

// PaginatedOrganizations represents a paginated list of organizations
type PaginatedOrganizations struct {
	Organizations []models.Organization
	TotalItems    int
	TotalPages    int
	CurrentPage   int
	HasNext       bool
	HasPrev       bool
}

// OrganizationRepository defines the interface for organization data access
// Following Clean Architecture principles, this interface allows for dependency injection
// and makes the service layer independent of database implementation details
type OrganizationRepository interface {
	// CreateOrganization creates a new organization in the database
	// Returns the created organization or an error if the operation fails
	CreateOrganization(ctx context.Context, org *models.Organization) error

	// FindOrganizationByID retrieves an organization by its unique identifier
	// Returns ErrOrganizationNotFound if no organization exists with the given ID
	FindOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)

	// FindOrganizationBySlug retrieves an organization by its URL-friendly slug
	// Returns ErrOrganizationNotFound if no organization exists with the given slug
	FindOrganizationBySlug(ctx context.Context, slug string) (*models.Organization, error)

	// ListOrganizations retrieves a paginated list of organizations with optional filters
	// Supports search by name, filtering by location and causes
	ListOrganizations(ctx context.Context, filters OrganizationFilters) (*PaginatedOrganizations, error)

	// UpdateOrganization updates an existing organization
	// Only updates the fields that are provided (partial updates supported)
	UpdateOrganization(ctx context.Context, org *models.Organization) error

	// DeleteOrganization soft deletes an organization by its ID
	DeleteOrganization(ctx context.Context, id uuid.UUID) error

	// AddMember adds a user as a member of an organization with a specific role
	AddMember(ctx context.Context, member *models.OrganizationMember) error

	// RemoveMember removes a user from an organization's team
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error

	// ListMembers retrieves all members of an organization with their user details
	ListMembers(ctx context.Context, orgID uuid.UUID) ([]models.OrganizationMember, error)

	// FindMemberByOrgAndUser retrieves a membership record by organization and user ID
	FindMemberByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error)

	// IsMember checks if a user is a member of an organization (any role)
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)

	// GetMemberRole retrieves the role of a user in an organization
	// Returns the role or empty string if not a member
	GetMemberRole(ctx context.Context, orgID, userID uuid.UUID) (models.OrganizationMemberRole, error)

	// IncrementVolunteerCount increases the total volunteer count for an organization
	IncrementVolunteerCount(ctx context.Context, orgID uuid.UUID, count int) error

	// IncrementHoursCount increases the total hours count for an organization
	IncrementHoursCount(ctx context.Context, orgID uuid.UUID, hours float64) error

	// UpdateAverageRating updates the average rating for an organization
	UpdateAverageRating(ctx context.Context, orgID uuid.UUID, rating float64) error

	// FindUserByEmail looks up a user ID by email address
	// Used for inviting team members
	FindUserByEmail(ctx context.Context, email string) (uuid.UUID, error)

	// FindOrganizationsByUserID retrieves all organizations a user is a member of
	FindOrganizationsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Organization, error)
}

// gormOrganizationRepository is the GORM implementation of OrganizationRepository
type gormOrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new instance of OrganizationRepository using GORM
func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &gormOrganizationRepository{
		db: db,
	}
}

// CreateOrganization creates a new organization in the database
func (r *gormOrganizationRepository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	// Validate input
	if org == nil {
		return fmt.Errorf("organization cannot be nil")
	}

	if org.Name == "" {
		return fmt.Errorf("organization name is required")
	}

	if org.Email == "" {
		return fmt.Errorf("organization email is required")
	}

	// Use a transaction to ensure atomic operation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if organization with the same slug already exists
		// The slug is auto-generated from name in BeforeCreate hook
		slug := models.GenerateSlug(org.Name)
		var existingOrg models.Organization
		result := tx.Where("slug = ?", slug).First(&existingOrg)

		if result.Error == nil {
			// Organization found, slug already exists
			return ErrOrganizationAlreadyExists
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Unexpected error during lookup
			return fmt.Errorf("failed to check existing organization: %w", result.Error)
		}

		// Create the organization
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// FindOrganizationByID retrieves an organization by its unique identifier
func (r *gormOrganizationRepository) FindOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidOrganizationID
	}

	var org models.Organization
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&org)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to find organization: %w", result.Error)
	}

	return &org, nil
}

// FindOrganizationBySlug retrieves an organization by its URL-friendly slug
func (r *gormOrganizationRepository) FindOrganizationBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	var org models.Organization
	result := r.db.WithContext(ctx).Where("slug = ?", slug).First(&org)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to find organization: %w", result.Error)
	}

	return &org, nil
}

// ListOrganizations retrieves a paginated list of organizations with optional filters
func (r *gormOrganizationRepository) ListOrganizations(ctx context.Context, filters OrganizationFilters) (*PaginatedOrganizations, error) {
	// Set default pagination values
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 {
		filters.Limit = 10
	}
	if filters.Limit > 100 {
		filters.Limit = 100 // Max limit to prevent performance issues
	}

	// Build query
	query := r.db.WithContext(ctx).Model(&models.Organization{})

	// Apply search filter (name)
	if filters.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Search+"%")
	}

	// Apply location filters
	if filters.City != "" {
		query = query.Where("city = ?", filters.City)
	}
	if filters.State != "" {
		query = query.Where("state = ?", filters.State)
	}

	// Apply cause filter (requires join with organization_causes table)
	// This will be implemented when the junction table is available
	// if len(filters.CauseIDs) > 0 {
	// 	query = query.Joins("JOIN organization_causes ON organization_causes.organization_id = organizations.id").
	// 		Where("organization_causes.cause_id IN ?", filters.CauseIDs).
	// 		Distinct()
	// }

	// Count total items
	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Calculate pagination
	offset := (filters.Page - 1) * filters.Limit
	totalPages := int((totalItems + int64(filters.Limit) - 1) / int64(filters.Limit))

	// Fetch paginated results
	var organizations []models.Organization
	result := query.
		Offset(offset).
		Limit(filters.Limit).
		Order("created_at DESC").
		Find(&organizations)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", result.Error)
	}

	return &PaginatedOrganizations{
		Organizations: organizations,
		TotalItems:    int(totalItems),
		TotalPages:    totalPages,
		CurrentPage:   filters.Page,
		HasNext:       filters.Page < totalPages,
		HasPrev:       filters.Page > 1,
	}, nil
}

// UpdateOrganization updates an existing organization
func (r *gormOrganizationRepository) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	if org == nil {
		return fmt.Errorf("organization cannot be nil")
	}

	if org.ID == uuid.Nil {
		return ErrInvalidOrganizationID
	}

	// Check if organization exists
	var existingOrg models.Organization
	if err := r.db.WithContext(ctx).Where("id = ?", org.ID).First(&existingOrg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrganizationNotFound
		}
		return fmt.Errorf("failed to check organization existence: %w", err)
	}

	// Update the organization
	result := r.db.WithContext(ctx).Model(org).Updates(org)
	if result.Error != nil {
		return fmt.Errorf("failed to update organization: %w", result.Error)
	}

	return nil
}

// DeleteOrganization soft deletes an organization by its ID
func (r *gormOrganizationRepository) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidOrganizationID
	}

	// Check if organization exists
	var org models.Organization
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrganizationNotFound
		}
		return fmt.Errorf("failed to check organization existence: %w", err)
	}

	// Soft delete the organization
	result := r.db.WithContext(ctx).Delete(&org)
	if result.Error != nil {
		return fmt.Errorf("failed to delete organization: %w", result.Error)
	}

	return nil
}

// IncrementVolunteerCount increases the total volunteer count for an organization
func (r *gormOrganizationRepository) IncrementVolunteerCount(ctx context.Context, orgID uuid.UUID, count int) error {
	if orgID == uuid.Nil {
		return ErrInvalidOrganizationID
	}

	result := r.db.WithContext(ctx).Model(&models.Organization{}).
		Where("id = ?", orgID).
		UpdateColumn("total_volunteers", gorm.Expr("total_volunteers + ?", count))

	if result.Error != nil {
		return fmt.Errorf("failed to increment volunteer count: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

// IncrementHoursCount increases the total hours count for an organization
func (r *gormOrganizationRepository) IncrementHoursCount(ctx context.Context, orgID uuid.UUID, hours float64) error {
	if orgID == uuid.Nil {
		return ErrInvalidOrganizationID
	}

	result := r.db.WithContext(ctx).Model(&models.Organization{}).
		Where("id = ?", orgID).
		UpdateColumn("total_hours", gorm.Expr("total_hours + ?", hours))

	if result.Error != nil {
		return fmt.Errorf("failed to increment hours count: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

// UpdateAverageRating updates the average rating for an organization
func (r *gormOrganizationRepository) UpdateAverageRating(ctx context.Context, orgID uuid.UUID, rating float64) error {
	if orgID == uuid.Nil {
		return ErrInvalidOrganizationID
	}

	result := r.db.WithContext(ctx).Model(&models.Organization{}).
		Where("id = ?", orgID).
		Update("avg_rating", rating)

	if result.Error != nil {
		return fmt.Errorf("failed to update average rating: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

// AddMember adds a user as a member of an organization with a specific role
func (r *gormOrganizationRepository) AddMember(ctx context.Context, member *models.OrganizationMember) error {
	if member == nil {
		return fmt.Errorf("member cannot be nil")
	}

	if member.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if member.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	// Check if membership already exists
	var existingMember models.OrganizationMember
	result := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", member.OrganizationID, member.UserID).
		First(&existingMember)

	if result.Error == nil {
		// Membership already exists
		return fmt.Errorf("user is already a member of this organization")
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Unexpected error during lookup
		return fmt.Errorf("failed to check existing membership: %w", result.Error)
	}

	// Create the membership
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

// RemoveMember removes a user from an organization (soft delete)
func (r *gormOrganizationRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	if orgID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	// Find the membership
	var member models.OrganizationMember
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("membership not found")
		}
		return fmt.Errorf("failed to find membership: %w", err)
	}

	// Soft delete the membership
	result := r.db.WithContext(ctx).Delete(&member)
	if result.Error != nil {
		return fmt.Errorf("failed to remove member: %w", result.Error)
	}

	return nil
}

// ListMembers retrieves all members of an organization with their user details
func (r *gormOrganizationRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]models.OrganizationMember, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization ID is required")
	}

	var members []models.OrganizationMember
	result := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("created_at ASC").
		Find(&members)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list members: %w", result.Error)
	}

	return members, nil
}

// FindMemberByOrgAndUser retrieves a membership record by organization and user ID
func (r *gormOrganizationRepository) FindMemberByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("organization ID is required")
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	var member models.OrganizationMember
	result := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("membership not found")
		}
		return nil, fmt.Errorf("failed to find membership: %w", result.Error)
	}

	return &member, nil
}

// IsMember checks if a user is a member of an organization (any role)
func (r *gormOrganizationRepository) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	if orgID == uuid.Nil || userID == uuid.Nil {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}

	return count > 0, nil
}

// GetMemberRole retrieves the role of a user in an organization
// Returns the role or empty string if not a member
func (r *gormOrganizationRepository) GetMemberRole(ctx context.Context, orgID, userID uuid.UUID) (models.OrganizationMemberRole, error) {
	if orgID == uuid.Nil || userID == uuid.Nil {
		return "", nil
	}

	var member models.OrganizationMember
	result := r.db.WithContext(ctx).
		Select("role").
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", nil // Not a member
		}
		return "", fmt.Errorf("failed to get member role: %w", result.Error)
	}

	return member.Role, nil
}

// FindUserByEmail looks up a user ID by email address
// Used for inviting team members
func (r *gormOrganizationRepository) FindUserByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	if email == "" {
		return uuid.Nil, fmt.Errorf("email is required")
	}

	var user struct {
		ID uuid.UUID `gorm:"column:id"`
	}

	result := r.db.WithContext(ctx).
		Table("users").
		Select("id").
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return uuid.Nil, fmt.Errorf("user not found with email: %s", email)
		}
		return uuid.Nil, fmt.Errorf("failed to lookup user: %w", result.Error)
	}

	return user.ID, nil
}

// FindOrganizationsByUserID retrieves all organizations a user is a member of
func (r *gormOrganizationRepository) FindOrganizationsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	var orgs []models.Organization

	// Join with organization_members table
	result := r.db.WithContext(ctx).
		Joins("INNER JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to find organizations: %w", result.Error)
	}

	return orgs, nil
}
