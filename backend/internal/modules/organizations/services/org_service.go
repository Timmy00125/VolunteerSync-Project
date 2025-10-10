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

	// GetTeamMembers retrieves all team members of an organization
	GetTeamMembers(ctx context.Context, orgID uuid.UUID, requestingUserID uuid.UUID) ([]TeamMemberOutput, error)

	// InviteMember invites a user to join an organization as a team member (admin only, FR-014)
	InviteMember(ctx context.Context, input InviteMemberInput, inviterID uuid.UUID) error

	// RemoveMember removes a user from an organization team (admin only)
	RemoveMember(ctx context.Context, orgID, userID, removerID uuid.UUID) error

	// GetDashboard retrieves dashboard metrics for an organization (admin/coordinator only)
	// Includes metrics, upcoming events, and recent registrations
	GetDashboard(ctx context.Context, orgID uuid.UUID, requestingUserID uuid.UUID) (*OrganizationDashboard, error)

	// GetUserOrganizations retrieves all organizations the user is a member of
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.Organization, error)
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

// TeamMemberOutput represents a team member with user details
type TeamMemberOutput struct {
	ID        string                        `json:"id"`
	UserID    string                        `json:"user_id"`
	FirstName string                        `json:"first_name"`
	LastName  string                        `json:"last_name"`
	Email     string                        `json:"email"`
	Role      models.OrganizationMemberRole `json:"role"`
	JoinedAt  time.Time                     `json:"joined_at"`
}

// InviteMemberInput represents the input for inviting a team member
type InviteMemberInput struct {
	OrganizationID uuid.UUID
	Email          string
	Role           models.OrganizationMemberRole
}

// OrganizationDashboard represents dashboard data for an organization
type OrganizationDashboard struct {
	Organization        *models.Organization `json:"organization"`
	Metrics             DashboardMetrics     `json:"metrics"`
	UpcomingEvents      []UpcomingEventInfo  `json:"upcoming_events"`
	RecentRegistrations []RecentRegistration `json:"recent_registrations"`
}

// DashboardMetrics represents key metrics for an organization dashboard
type DashboardMetrics struct {
	TotalVolunteers        int     `json:"total_volunteers"`
	ActiveVolunteers       int     `json:"active_volunteers"`
	TotalHours             float64 `json:"total_hours"`
	HoursThisMonth         float64 `json:"hours_this_month"`
	TotalEvents            int     `json:"total_events"`
	UpcomingEvents         int     `json:"upcoming_events"`
	EventsThisMonth        int     `json:"events_this_month"`
	VolunteerRetentionRate float64 `json:"volunteer_retention_rate"`
}

// UpcomingEventInfo represents an upcoming event for dashboard display
type UpcomingEventInfo struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Date            time.Time `json:"date"`
	StartTime       time.Time `json:"start_time"`
	RegisteredCount int       `json:"registered_count"`
	Capacity        int       `json:"capacity"`
	Status          string    `json:"status"`
}

// RecentRegistration represents a recent volunteer registration
type RecentRegistration struct {
	ID               uuid.UUID `json:"id"`
	VolunteerName    string    `json:"volunteer_name"`
	OpportunityTitle string    `json:"opportunity_title"`
	RegisteredAt     time.Time `json:"registered_at"`
	Status           string    `json:"status"`
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

// GetTeamMembers retrieves all team members of an organization
func (s *organizationService) GetTeamMembers(ctx context.Context, orgID uuid.UUID, requestingUserID uuid.UUID) ([]TeamMemberOutput, error) {
	// Validate that the requesting user is a member (admin or coordinator) of the organization
	role, err := s.repo.GetMemberRole(ctx, orgID, requestingUserID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID.String(),
			"user_id": requestingUserID.String(),
			"error":   err.Error(),
		}).Error("Failed to check user membership")
		return nil, apperrors.NewInternalServerError("failed to verify membership")
	}

	if role == "" {
		return nil, apperrors.NewUnauthorizedError("not a member of this organization")
	}

	// Get all members
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id": orgID.String(),
			"error":  err.Error(),
		}).Error("Failed to list organization members")
		return nil, apperrors.NewInternalServerError("failed to retrieve team members")
	}

	// Note: For now, we're returning member data without user details (first_name, last_name, email)
	// These would need to be fetched from the auth/users module via a cross-module call
	// For now, we'll return what we have
	output := make([]TeamMemberOutput, len(members))
	for i, member := range members {
		output[i] = TeamMemberOutput{
			ID:        member.ID.String(),
			UserID:    member.UserID.String(),
			FirstName: "", // TODO: Fetch from users module
			LastName:  "", // TODO: Fetch from users module
			Email:     "", // TODO: Fetch from users module
			Role:      member.Role,
			JoinedAt:  member.JoinedAt,
		}
	}

	return output, nil
}

// InviteMember invites a user to join an organization as a team member (admin only)
func (s *organizationService) InviteMember(ctx context.Context, input InviteMemberInput, inviterID uuid.UUID) error {
	// Validate that the inviter is an admin of the organization
	role, err := s.repo.GetMemberRole(ctx, input.OrganizationID, inviterID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  input.OrganizationID.String(),
			"user_id": inviterID.String(),
			"error":   err.Error(),
		}).Error("Failed to check inviter membership")
		return apperrors.NewInternalServerError("failed to verify membership")
	}

	if role != models.OrgRoleAdmin {
		return apperrors.NewForbiddenError("only admins can invite team members")
	}

	// Validate input
	if err := validateEmail(input.Email); err != nil {
		return apperrors.NewValidationError("invalid email format", nil)
	}

	if input.Role != models.OrgRoleAdmin && input.Role != models.OrgRoleCoordinator {
		return apperrors.NewValidationError("invalid role: must be 'admin' or 'coordinator'", nil)
	}

	// Look up user by email
	userID, err := s.repo.FindUserByEmail(ctx, input.Email)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"email": input.Email,
			"error": err.Error(),
		}).Error("Failed to lookup user by email")
		return apperrors.NewValidationError("no user found with this email address", nil)
	}

	// Check if the user is already a member
	existingRole, err := s.repo.GetMemberRole(ctx, input.OrganizationID, userID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  input.OrganizationID.String(),
			"user_id": userID.String(),
			"error":   err.Error(),
		}).Error("Failed to check existing membership")
		return apperrors.NewInternalServerError("failed to verify membership status")
	}

	if existingRole != "" {
		return apperrors.NewConflictError("user is already a member of this organization")
	}

	// Add the user as a member
	member := &models.OrganizationMember{
		OrganizationID: input.OrganizationID,
		UserID:         userID,
		Role:           input.Role,
		JoinedAt:       time.Now(),
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  input.OrganizationID.String(),
			"user_id": userID.String(),
			"error":   err.Error(),
		}).Error("Failed to add member")
		return apperrors.NewInternalServerError("failed to add team member")
	}

	s.logger.WithFields(map[string]interface{}{
		"org_id":  input.OrganizationID.String(),
		"user_id": userID.String(),
		"role":    input.Role,
		"email":   input.Email,
	}).Info("Team member added successfully")

	return nil
}

// RemoveMember removes a user from an organization team (admin only)
func (s *organizationService) RemoveMember(ctx context.Context, orgID, userID, removerID uuid.UUID) error {
	// Validate that the remover is an admin of the organization
	role, err := s.repo.GetMemberRole(ctx, orgID, removerID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID.String(),
			"user_id": removerID.String(),
			"error":   err.Error(),
		}).Error("Failed to check remover membership")
		return apperrors.NewInternalServerError("failed to verify membership")
	}

	if role != models.OrgRoleAdmin {
		return apperrors.NewForbiddenError("only admins can remove team members")
	}

	// Check that the user being removed is not the last admin
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id": orgID.String(),
			"error":  err.Error(),
		}).Error("Failed to list members")
		return apperrors.NewInternalServerError("failed to verify member status")
	}

	// Count admins and check if we're removing the last one
	adminCount := 0
	isRemovingAdmin := false
	for _, member := range members {
		if member.Role == models.OrgRoleAdmin {
			adminCount++
			if member.UserID == userID {
				isRemovingAdmin = true
			}
		}
	}

	if isRemovingAdmin && adminCount <= 1 {
		return apperrors.NewValidationError("cannot remove the last admin from an organization", nil)
	}

	// Remove the member
	if err := s.repo.RemoveMember(ctx, orgID, userID); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID.String(),
			"user_id": userID.String(),
			"error":   err.Error(),
		}).Error("Failed to remove member")
		return apperrors.NewInternalServerError("failed to remove team member")
	}

	s.logger.WithFields(map[string]interface{}{
		"org_id":     orgID.String(),
		"user_id":    userID.String(),
		"removed_by": removerID.String(),
	}).Info("Team member removed successfully")

	return nil
}

// GetDashboard retrieves dashboard metrics for an organization
func (s *organizationService) GetDashboard(ctx context.Context, orgID uuid.UUID, requestingUserID uuid.UUID) (*OrganizationDashboard, error) {
	// Validate that the requesting user is a member of the organization
	role, err := s.repo.GetMemberRole(ctx, orgID, requestingUserID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"org_id":  orgID.String(),
			"user_id": requestingUserID.String(),
			"error":   err.Error(),
		}).Error("Failed to check user membership")
		return nil, apperrors.NewInternalServerError("failed to verify membership")
	}

	if role == "" {
		return nil, apperrors.NewUnauthorizedError("not a member of this organization")
	}

	// Get organization details
	org, err := s.repo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, repositories.ErrOrganizationNotFound) {
			return nil, apperrors.NewNotFoundError("Organization")
		}
		s.logger.WithFields(map[string]interface{}{
			"org_id": orgID.String(),
			"error":  err.Error(),
		}).Error("Failed to get organization")
		return nil, apperrors.NewInternalServerError("Failed to retrieve organization")
	}

	// Initialize dashboard with basic data
	dashboard := &OrganizationDashboard{
		Organization: org,
		Metrics: DashboardMetrics{
			TotalVolunteers:        org.TotalVolunteers,
			ActiveVolunteers:       0, // TODO: Calculate from recent registrations
			TotalHours:             org.TotalHours,
			HoursThisMonth:         0, // TODO: Calculate from hours logs
			TotalEvents:            0, // TODO: Count from opportunities
			UpcomingEvents:         0, // TODO: Count upcoming opportunities
			EventsThisMonth:        0, // TODO: Count events this month
			VolunteerRetentionRate: 0, // TODO: Calculate retention rate
		},
		UpcomingEvents:      []UpcomingEventInfo{},  // TODO: Fetch from opportunities module
		RecentRegistrations: []RecentRegistration{}, // TODO: Fetch from registrations module
	}

	// Note: Full implementation would require cross-module calls to:
	// 1. Opportunities module - for upcoming events count and details
	// 2. Registrations module - for recent registrations and active volunteers
	// 3. Hours module - for hours this month calculation
	// These are currently placeholders and will be implemented with proper adapters

	return dashboard, nil
}

// GetUserOrganizations retrieves all organizations the user is a member of
func (s *organizationService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	if userID == uuid.Nil {
		return nil, apperrors.NewValidationError("Invalid user ID", nil)
	}

	orgs, err := s.repo.FindOrganizationsByUserID(ctx, userID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id": userID.String(),
			"error":   err.Error(),
		}).Error("Failed to get user organizations")
		return nil, apperrors.NewInternalServerError("Failed to retrieve user organizations")
	}

	return orgs, nil
}
