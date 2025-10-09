package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/organizations/repositories"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/google/uuid"
)

// Custom errors for service operations
var (
	// ErrInvalidOrganizationName is returned when the organization name is invalid
	ErrInvalidOrganizationName = errors.New("organization name is required")
	// ErrInvalidOrganizationEmail is returned when the organization email is invalid
	ErrInvalidOrganizationEmail = errors.New("organization email is invalid")
	// ErrOrganizationNotFound is returned when an organization cannot be found
	ErrOrganizationNotFound = repositories.ErrOrganizationNotFound
	// ErrOrganizationAlreadyExists is returned when an organization with the same name exists
	ErrOrganizationAlreadyExists = repositories.ErrOrganizationAlreadyExists
	// ErrUnauthorized is returned when a user doesn't have permission for an operation
	ErrUnauthorized = errors.New("unauthorized to perform this action")
)

// GeocodingService defines the interface for geocoding addresses to coordinates
// This will be implemented in the geocoding package (Task T133)
type GeocodingService interface {
	GeocodeAddress(ctx context.Context, address string) (lat, lng float64, err error)
}

// OrganizationService encapsulates organization business logic, providing methods for
// organization lifecycle workflows such as creation, updates, listing, and search.
// Handlers should depend on this interface to keep HTTP transport concerns isolated
// from domain logic.
type OrganizationService interface {
	// CreateOrganization creates a new organization with auto-verification and geocoding
	// As per FR-010 and FR-015: organizations are auto-verified on creation
	CreateOrganization(ctx context.Context, input CreateOrganizationInput, creatorUserID uuid.UUID) (*models.Organization, error)

	// GetOrganization retrieves an organization by ID (public endpoint)
	GetOrganization(ctx context.Context, orgID uuid.UUID) (*models.Organization, error)

	// GetOrganizationBySlug retrieves an organization by its slug (public endpoint)
	GetOrganizationBySlug(ctx context.Context, slug string) (*models.Organization, error)

	// UpdateOrganization updates an organization (admin only)
	// Validates that the requesting user is an admin of the organization
	UpdateOrganization(ctx context.Context, orgID uuid.UUID, input UpdateOrganizationInput, userID uuid.UUID) (*models.Organization, error)

	// ListOrganizations retrieves a paginated list of organizations with filters
	// Supports search by name, filtering by location and causes
	ListOrganizations(ctx context.Context, filters OrganizationListFilters) (*OrganizationListResponse, error)

	// DeleteOrganization soft deletes an organization (admin only)
	DeleteOrganization(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) error

	// InviteMember invites a user to join an organization (admin only, FR-014)
	// This will be implemented when the Organization_Member module is ready
	// InviteMember(ctx context.Context, orgID, userID, inviterID uuid.UUID, role string) error

	// RemoveMember removes a user from an organization (admin only)
	// RemoveMember(ctx context.Context, orgID, userID, removerID uuid.UUID) error
}

// CreateOrganizationInput represents the input for creating a new organization
type CreateOrganizationInput struct {
	Name             string
	MissionStatement *string
	Description      *string
	Website          *string
	Email            string
	Phone            *string
	AddressLine1     *string
	AddressLine2     *string
	City             *string
	State            *string
	PostalCode       *string
	Country          string
	LogoURL          *string
	BannerURL        *string
	CauseIDs         []uuid.UUID // Cause categories for the organization
}

// UpdateOrganizationInput represents the input for updating an organization
type UpdateOrganizationInput struct {
	Name             *string
	MissionStatement *string
	Description      *string
	Website          *string
	Email            *string
	Phone            *string
	AddressLine1     *string
	AddressLine2     *string
	City             *string
	State            *string
	PostalCode       *string
	Country          *string
	LogoURL          *string
	BannerURL        *string
	CauseIDs         []uuid.UUID
}

// OrganizationListFilters represents filters for listing organizations
type OrganizationListFilters struct {
	Search   string
	City     string
	State    string
	CauseIDs []uuid.UUID
	Page     int
	Limit    int
}

// OrganizationListResponse represents a paginated list of organizations
type OrganizationListResponse struct {
	Organizations []models.Organization `json:"data"`
	Pagination    PaginationInfo        `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalPages int  `json:"total_pages"`
	TotalItems int  `json:"total_items"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// organizationService is the concrete implementation of OrganizationService
type organizationService struct {
	repo             repositories.OrganizationRepository
	geocodingService GeocodingService // Optional, can be nil
	logger           *logger.Logger
}

// NewOrganizationService creates a new instance of OrganizationService
func NewOrganizationService(
	repo repositories.OrganizationRepository,
	geocodingService GeocodingService,
	logger *logger.Logger,
) OrganizationService {
	return &organizationService{
		repo:             repo,
		geocodingService: geocodingService,
		logger:           logger,
	}
}

// CreateOrganization creates a new organization with auto-verification and geocoding
func (s *organizationService) CreateOrganization(
	ctx context.Context,
	input CreateOrganizationInput,
	creatorUserID uuid.UUID,
) (*models.Organization, error) {
	// Validate input
	if err := s.validateCreateInput(input); err != nil {
		return nil, apperrors.NewValidationError(err.Error(), nil)
	}

	// Create organization model
	org := &models.Organization{
		Name:             input.Name,
		MissionStatement: input.MissionStatement,
		Description:      input.Description,
		Website:          input.Website,
		Email:            input.Email,
		Phone:            input.Phone,
		AddressLine1:     input.AddressLine1,
		AddressLine2:     input.AddressLine2,
		City:             input.City,
		State:            input.State,
		PostalCode:       input.PostalCode,
		Country:          input.Country,
		LogoURL:          input.LogoURL,
		BannerURL:        input.BannerURL,
	}

	// Set default country if not provided
	if org.Country == "" {
		org.Country = "United States"
	}

	// Auto-verify organization (FR-015: V1 auto-verifies all organizations on creation)
	// This is also handled in the BeforeCreate hook, but we set it explicitly here for clarity
	org.Verify()

	// Geocode address if provided and geocoding service is available
	if s.geocodingService != nil && org.HasAddress() {
		if err := s.geocodeOrganization(ctx, org); err != nil {
			// Log the error but don't fail the creation - geocoding is optional
			s.logger.WithFields(map[string]interface{}{
				"org_name": org.Name,
				"error":    err.Error(),
			}).Warn("Failed to geocode organization address")
		}
	}

	// Create organization in database
	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		if errors.Is(err, repositories.ErrOrganizationAlreadyExists) {
			return nil, apperrors.NewConflictError("An organization with this name already exists")
		}
		s.logger.WithFields(map[string]interface{}{
			"org_name": org.Name,
			"error":    err.Error(),
		}).Error("Failed to create organization")
		return nil, apperrors.NewInternalServerError("Failed to create organization")
	}

	// Create organization member record for the creator as admin (FR-014)
	now := time.Now()
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         creatorUserID,
		Role:           models.OrgRoleAdmin,
		JoinedAt:       now,
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		// Log the error but don't fail the organization creation
		// The organization admin can manually add the creator later
		s.logger.WithFields(map[string]interface{}{
			"org_id":          org.ID,
			"creator_user_id": creatorUserID,
			"error":           err.Error(),
		}).Error("Failed to create organization member record for creator")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"org_id":          org.ID,
			"creator_user_id": creatorUserID,
			"role":            models.OrgRoleAdmin,
		}).Info("Organization creator added as admin")
	}

	s.logger.WithFields(map[string]interface{}{
		"org_id":          org.ID,
		"org_name":        org.Name,
		"creator_user_id": creatorUserID,
	}).Info("Organization created successfully")

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *organizationService) GetOrganization(ctx context.Context, orgID uuid.UUID) (*models.Organization, error) {
	if orgID == uuid.Nil {
		return nil, apperrors.NewValidationError("Invalid organization ID", nil)
	}

	org, err := s.repo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, repositories.ErrOrganizationNotFound) {
			return nil, apperrors.NewNotFoundError("Organization")
		}
		s.logger.WithFields(map[string]interface{}{
			"org_id": orgID,
			"error":  err.Error(),
		}).Error("Failed to get organization")
		return nil, apperrors.NewInternalServerError("Failed to retrieve organization")
	}

	return org, nil
}

// GetOrganizationBySlug retrieves an organization by its slug
func (s *organizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	if slug == "" {
		return nil, apperrors.NewValidationError("Slug cannot be empty", nil)
	}

	org, err := s.repo.FindOrganizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repositories.ErrOrganizationNotFound) {
			return nil, apperrors.NewNotFoundError("Organization")
		}
		s.logger.WithFields(map[string]interface{}{
			"slug":  slug,
			"error": err.Error(),
		}).Error("Failed to get organization by slug")
		return nil, apperrors.NewInternalServerError("Failed to retrieve organization")
	}

	return org, nil
}

// UpdateOrganization updates an organization (admin only)
func (s *organizationService) UpdateOrganization(
	ctx context.Context,
	orgID uuid.UUID,
	input UpdateOrganizationInput,
	userID uuid.UUID,
) (*models.Organization, error) {
	// Validate input
	if orgID == uuid.Nil {
		return nil, apperrors.NewValidationError("Invalid organization ID", nil)
	}

	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("Invalid user ID", nil)
	}

	// Get existing organization
	org, err := s.repo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, repositories.ErrOrganizationNotFound) {
			return nil, apperrors.NewNotFoundError("Organization")
		}
		return nil, apperrors.NewInternalServerError("Failed to retrieve organization")
	}

	// Verify that the user is an admin of this organization
	// Super admins (if they exist in the user system) should also be allowed
	// We check the user role from context in the handler/middleware layer
	role, err := s.repo.GetMemberRole(ctx, orgID, userID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID,
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user role in organization")
		return nil, apperrors.NewInternalServerError("Failed to verify permissions")
	}

	// Only organization admins can update organization details
	if role != models.OrgRoleAdmin {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID,
			"user_id": userID,
			"role":    role,
		}).Warn("User attempted to update organization without admin privileges")
		return nil, apperrors.NewForbiddenError("Only organization administrators can update organization details")
	}

	// Apply updates (only update fields that are provided)
	addressChanged := false

	if input.Name != nil && *input.Name != "" {
		org.Name = *input.Name
	}
	if input.MissionStatement != nil {
		org.MissionStatement = input.MissionStatement
	}
	if input.Description != nil {
		org.Description = input.Description
	}
	if input.Website != nil {
		org.Website = input.Website
	}
	if input.Email != nil && *input.Email != "" {
		if err := validateEmail(*input.Email); err != nil {
			return nil, apperrors.NewValidationError("Invalid email address", nil)
		}
		org.Email = *input.Email
	}
	if input.Phone != nil {
		org.Phone = input.Phone
	}
	if input.AddressLine1 != nil {
		org.AddressLine1 = input.AddressLine1
		addressChanged = true
	}
	if input.AddressLine2 != nil {
		org.AddressLine2 = input.AddressLine2
		addressChanged = true
	}
	if input.City != nil {
		org.City = input.City
		addressChanged = true
	}
	if input.State != nil {
		org.State = input.State
		addressChanged = true
	}
	if input.PostalCode != nil {
		org.PostalCode = input.PostalCode
		addressChanged = true
	}
	if input.Country != nil && *input.Country != "" {
		org.Country = *input.Country
		addressChanged = true
	}
	if input.LogoURL != nil {
		org.LogoURL = input.LogoURL
	}
	if input.BannerURL != nil {
		org.BannerURL = input.BannerURL
	}

	// Re-geocode if address changed
	if addressChanged && s.geocodingService != nil && org.HasAddress() {
		if err := s.geocodeOrganization(ctx, org); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"org_id": org.ID,
				"error":  err.Error(),
			}).Warn("Failed to geocode updated organization address")
		}
	}

	// Update organization in database
	if err := s.repo.UpdateOrganization(ctx, org); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id": orgID,
			"error":  err.Error(),
		}).Error("Failed to update organization")
		return nil, apperrors.NewInternalServerError("Failed to update organization")
	}

	s.logger.WithFields(map[string]interface{}{
		"org_id":  org.ID,
		"user_id": userID,
	}).Info("Organization updated successfully")

	return org, nil
}

// ListOrganizations retrieves a paginated list of organizations with filters
func (s *organizationService) ListOrganizations(
	ctx context.Context,
	filters OrganizationListFilters,
) (*OrganizationListResponse, error) {
	// Convert service filters to repository filters
	repoFilters := repositories.OrganizationFilters{
		Search:   filters.Search,
		City:     filters.City,
		State:    filters.State,
		CauseIDs: filters.CauseIDs,
		Page:     filters.Page,
		Limit:    filters.Limit,
	}

	// Get paginated results from repository
	result, err := s.repo.ListOrganizations(ctx, repoFilters)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to list organizations")
		return nil, apperrors.NewInternalServerError("Failed to retrieve organizations")
	}

	// Convert repository response to service response
	return &OrganizationListResponse{
		Organizations: result.Organizations,
		Pagination: PaginationInfo{
			Page:       result.CurrentPage,
			Limit:      filters.Limit,
			TotalPages: result.TotalPages,
			TotalItems: result.TotalItems,
			HasNext:    result.HasNext,
			HasPrev:    result.HasPrev,
		},
	}, nil
}

// DeleteOrganization soft deletes an organization (admin only)
func (s *organizationService) DeleteOrganization(
	ctx context.Context,
	orgID uuid.UUID,
	userID uuid.UUID,
) error {
	if orgID == uuid.Nil {
		return apperrors.NewValidationError("Invalid organization ID", nil)
	}

	if userID == uuid.Nil {
		return apperrors.NewValidationError("Invalid user ID", nil)
	}

	// Verify that the user is an admin of this organization
	// Only organization admins can delete organizations
	role, err := s.repo.GetMemberRole(ctx, orgID, userID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID,
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user role in organization")
		return apperrors.NewInternalServerError("Failed to verify permissions")
	}

	// Only organization admins can delete organizations
	if role != models.OrgRoleAdmin {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID,
			"user_id": userID,
			"role":    role,
		}).Warn("User attempted to delete organization without admin privileges")
		return apperrors.NewForbiddenError("Only organization administrators can delete organizations")
	}

	// Delete organization
	if err := s.repo.DeleteOrganization(ctx, orgID); err != nil {
		if errors.Is(err, repositories.ErrOrganizationNotFound) {
			return apperrors.NewNotFoundError("Organization")
		}
		s.logger.WithFields(map[string]interface{}{
			"org_id": orgID,
			"error":  err.Error(),
		}).Error("Failed to delete organization")
		return apperrors.NewInternalServerError("Failed to delete organization")
	}

	s.logger.WithFields(map[string]interface{}{
		"org_id":  orgID,
		"user_id": userID,
	}).Info("Organization deleted successfully")

	return nil
}

// Helper methods

// validateCreateInput validates the input for creating an organization
func (s *organizationService) validateCreateInput(input CreateOrganizationInput) error {
	if input.Name == "" {
		return ErrInvalidOrganizationName
	}

	if input.Email == "" {
		return ErrInvalidOrganizationEmail
	}

	if err := validateEmail(input.Email); err != nil {
		return err
	}

	return nil
}

// validateEmail validates an email address format
func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// geocodeOrganization geocodes the organization's address and sets the coordinates
func (s *organizationService) geocodeOrganization(ctx context.Context, org *models.Organization) error {
	if s.geocodingService == nil {
		return fmt.Errorf("geocoding service not available")
	}

	if !org.HasAddress() {
		return fmt.Errorf("organization does not have a complete address")
	}

	// Get full address string
	fullAddress := org.GetFullAddress()

	// Add timeout to geocoding request
	geocodeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Geocode the address
	lat, lng, err := s.geocodingService.GeocodeAddress(geocodeCtx, fullAddress)
	if err != nil {
		return fmt.Errorf("geocoding failed: %w", err)
	}

	// Set coordinates
	org.SetCoordinates(lat, lng)

	s.logger.WithFields(map[string]interface{}{
		"org_name": org.Name,
		"lat":      lat,
		"lng":      lng,
	}).Debug("Organization geocoded successfully")

	return nil
}
