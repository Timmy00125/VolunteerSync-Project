package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/repositories"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
	"github.com/google/uuid"
)

// Custom errors for service operations
var (
	// ErrInvalidMessageType is returned when the message type is invalid
	ErrInvalidMessageType = errors.New("invalid message type")
	// ErrMissingRecipients is returned when a direct message has no recipients
	ErrMissingRecipients = errors.New("direct message must have at least one recipient")
	// ErrMissingOpportunityID is returned when a broadcast message has no opportunity ID
	ErrMissingOpportunityID = errors.New("broadcast message must have an opportunity ID")
	// ErrEmptyMessageContent is returned when the message content is empty
	ErrEmptyMessageContent = errors.New("message content is required")
	// ErrNotificationNotFound is returned when a notification cannot be found
	ErrNotificationNotFound = repositories.ErrNotificationNotFound
	// ErrUnauthorized is returned when a user doesn't have permission for an operation
	ErrUnauthorized = errors.New("unauthorized to perform this action")
)

// RegistrationRepository defines the interface for fetching registration data
// This is needed to get volunteers registered for an opportunity (broadcast messages)
type RegistrationRepository interface {
	FindVolunteersByOpportunity(ctx context.Context, opportunityID uuid.UUID) ([]uuid.UUID, error)
}

// CommunicationsService encapsulates communications business logic
type CommunicationsService interface {
	SendDirectMessage(ctx context.Context, input SendDirectMessageInput) (*models.Message, error)
	SendBroadcastMessage(ctx context.Context, input SendBroadcastMessageInput) (*models.Message, error)
	CreateNotification(ctx context.Context, input CreateNotificationInput) (*models.Notification, error)
	SendEventReminders(ctx context.Context, opportunityID uuid.UUID, reminderType ReminderType) error
	GetUserNotifications(ctx context.Context, filters GetNotificationsFilters) (*NotificationListResponse, error)
	MarkNotificationRead(ctx context.Context, notificationID, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	BulkCreateNotifications(ctx context.Context, notifications []CreateNotificationInput) error
}

// SendDirectMessageInput represents input for sending a direct message
type SendDirectMessageInput struct {
	SenderID     uuid.UUID
	RecipientIDs []uuid.UUID
	Subject      *string
	Content      string
}

// SendBroadcastMessageInput represents input for sending a broadcast message
type SendBroadcastMessageInput struct {
	SenderID      uuid.UUID
	OpportunityID uuid.UUID
	Subject       *string
	Content       string
}

// CreateNotificationInput represents input for creating a notification
type CreateNotificationInput struct {
	RecipientID       uuid.UUID
	NotificationType  models.NotificationType
	Title             string
	Content           string
	ActionURL         *string
	Priority          models.NotificationPriority
	RelatedEntityType *string
	RelatedEntityID   *uuid.UUID
	DeliveryMethod    models.NotificationDeliveryMethod
}

// ReminderType represents the type of event reminder
type ReminderType string

const (
	// ReminderType24Hours is sent 24 hours before event
	ReminderType24Hours ReminderType = "24h"
	// ReminderType2Hours is sent 2 hours before event
	ReminderType2Hours ReminderType = "2h"
)

// GetNotificationsFilters represents filters for retrieving notifications
type GetNotificationsFilters struct {
	UserID           uuid.UUID
	Unread           *bool
	NotificationType *models.NotificationType
	Priority         *models.NotificationPriority
	Page             int
	Limit            int
}

// NotificationListResponse represents a paginated list of notifications
type NotificationListResponse struct {
	Notifications []models.Notification `json:"notifications"`
	Pagination    PaginationInfo        `json:"pagination"`
	UnreadCount   int                   `json:"unread_count"`
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

// communicationsService is the concrete implementation of CommunicationsService
type communicationsService struct {
	repo    repositories.CommunicationsRepository
	regRepo RegistrationRepository // Optional for fetching event volunteers
	logger  *logger.Logger
}

// NewCommunicationsService creates a new instance of CommunicationsService
func NewCommunicationsService(
	repo repositories.CommunicationsRepository,
	regRepo RegistrationRepository,
	logger *logger.Logger,
) CommunicationsService {
	return &communicationsService{
		repo:    repo,
		regRepo: regRepo,
		logger:  logger,
	}
}

// SendDirectMessage sends a message to specific recipients
func (s *communicationsService) SendDirectMessage(ctx context.Context, input SendDirectMessageInput) (*models.Message, error) {
	if input.SenderID == uuid.Nil {
		return nil, fmt.Errorf("sender ID is required")
	}

	if len(input.RecipientIDs) == 0 {
		return nil, ErrMissingRecipients
	}

	if input.Content == "" {
		return nil, ErrEmptyMessageContent
	}

	message := &models.Message{
		SenderID:    input.SenderID,
		MessageType: models.MessageTypeDirect,
		Subject:     input.Subject,
		Content:     input.Content,
		SentAt:      time.Now(),
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		s.logger.ErrorWithErr("Failed to create message", err)
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	recipients := make([]models.MessageRecipient, 0, len(input.RecipientIDs))
	for _, recipientID := range input.RecipientIDs {
		recipients = append(recipients, models.MessageRecipient{
			MessageID:   message.ID,
			RecipientID: recipientID,
		})
	}

	if err := s.repo.CreateMessageRecipients(ctx, recipients); err != nil {
		s.logger.ErrorWithErr("Failed to create message recipients", err)
		return nil, fmt.Errorf("failed to create message recipients: %w", err)
	}

	s.logger.Info("Direct message sent successfully")

	return message, nil
}

// SendBroadcastMessage sends a message to all volunteers registered for an opportunity
func (s *communicationsService) SendBroadcastMessage(ctx context.Context, input SendBroadcastMessageInput) (*models.Message, error) {
	if input.SenderID == uuid.Nil {
		return nil, fmt.Errorf("sender ID is required")
	}

	if input.OpportunityID == uuid.Nil {
		return nil, ErrMissingOpportunityID
	}

	if input.Content == "" {
		return nil, ErrEmptyMessageContent
	}

	if s.regRepo == nil {
		return nil, fmt.Errorf("registration repository not available")
	}

	volunteerIDs, err := s.regRepo.FindVolunteersByOpportunity(ctx, input.OpportunityID)
	if err != nil {
		s.logger.ErrorWithErr("Failed to fetch volunteers for opportunity", err)
		return nil, fmt.Errorf("failed to fetch volunteers: %w", err)
	}

	if len(volunteerIDs) == 0 {
		s.logger.Warn("No volunteers found for broadcast message")
		return nil, fmt.Errorf("no volunteers registered for this opportunity")
	}

	message := &models.Message{
		SenderID:      input.SenderID,
		OpportunityID: &input.OpportunityID,
		MessageType:   models.MessageTypeBroadcast,
		Subject:       input.Subject,
		Content:       input.Content,
		SentAt:        time.Now(),
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		s.logger.ErrorWithErr("Failed to create broadcast message", err)
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	recipients := make([]models.MessageRecipient, 0, len(volunteerIDs))
	for _, volunteerID := range volunteerIDs {
		recipients = append(recipients, models.MessageRecipient{
			MessageID:   message.ID,
			RecipientID: volunteerID,
		})
	}

	if err := s.repo.CreateMessageRecipients(ctx, recipients); err != nil {
		s.logger.ErrorWithErr("Failed to create broadcast message recipients", err)
		return nil, fmt.Errorf("failed to create message recipients: %w", err)
	}

	s.logger.Info("Broadcast message sent successfully")

	return message, nil
}

// CreateNotification creates a notification for a user
func (s *communicationsService) CreateNotification(ctx context.Context, input CreateNotificationInput) (*models.Notification, error) {
	if input.RecipientID == uuid.Nil {
		return nil, fmt.Errorf("recipient ID is required")
	}

	if input.Title == "" || input.Content == "" {
		return nil, fmt.Errorf("title and content are required")
	}

	notification := &models.Notification{
		RecipientID:       input.RecipientID,
		NotificationType:  input.NotificationType,
		Title:             input.Title,
		Content:           input.Content,
		ActionURL:         input.ActionURL,
		Priority:          input.Priority,
		RelatedEntityType: input.RelatedEntityType,
		RelatedEntityID:   input.RelatedEntityID,
		DeliveryMethod:    input.DeliveryMethod,
		SentAt:            time.Now(),
	}

	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}
	if notification.DeliveryMethod == "" {
		notification.DeliveryMethod = models.NotificationDeliveryMethodInApp
	}

	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		s.logger.ErrorWithErr("Failed to create notification", err)
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("Notification created successfully")

	return notification, nil
}

// SendEventReminders sends reminder notifications to all volunteers registered for an event
func (s *communicationsService) SendEventReminders(ctx context.Context, opportunityID uuid.UUID, reminderType ReminderType) error {
	if opportunityID == uuid.Nil {
		return fmt.Errorf("opportunity ID is required")
	}

	if s.regRepo == nil {
		return fmt.Errorf("registration repository not available")
	}

	volunteerIDs, err := s.regRepo.FindVolunteersByOpportunity(ctx, opportunityID)
	if err != nil {
		s.logger.ErrorWithErr("Failed to fetch volunteers for event reminder", err)
		return fmt.Errorf("failed to fetch volunteers: %w", err)
	}

	if len(volunteerIDs) == 0 {
		s.logger.Info("No volunteers to remind")
		return nil
	}

	var title, content string
	var priority models.NotificationPriority

	switch reminderType {
	case ReminderType24Hours:
		title = "Event Reminder - 24 Hours"
		content = "Your volunteer event starts in 24 hours. Don't forget to check in!"
		priority = models.NotificationPriorityNormal
	case ReminderType2Hours:
		title = "Event Reminder - 2 Hours"
		content = "Your volunteer event starts in 2 hours. See you soon!"
		priority = models.NotificationPriorityHigh
	default:
		return fmt.Errorf("invalid reminder type: %s", reminderType)
	}

	notifications := make([]CreateNotificationInput, 0, len(volunteerIDs))
	for _, volunteerID := range volunteerIDs {
		notifications = append(notifications, CreateNotificationInput{
			RecipientID:       volunteerID,
			NotificationType:  models.NotificationTypeEventReminder,
			Title:             title,
			Content:           content,
			Priority:          priority,
			RelatedEntityType: stringPtr("opportunity"),
			RelatedEntityID:   &opportunityID,
			DeliveryMethod:    models.NotificationDeliveryMethodInApp,
		})
	}

	if err := s.BulkCreateNotifications(ctx, notifications); err != nil {
		s.logger.ErrorWithErr("Failed to send event reminders", err)
		return fmt.Errorf("failed to send event reminders: %w", err)
	}

	s.logger.Info("Event reminders sent successfully")

	return nil
}

// GetUserNotifications retrieves paginated notifications for a user with filters
func (s *communicationsService) GetUserNotifications(ctx context.Context, filters GetNotificationsFilters) (*NotificationListResponse, error) {
	if filters.UserID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}

	repoFilters := repositories.NotificationFilters{
		RecipientID:      filters.UserID,
		Unread:           filters.Unread,
		NotificationType: filters.NotificationType,
		Priority:         filters.Priority,
		Page:             filters.Page,
		Limit:            filters.Limit,
	}

	result, err := s.repo.FindNotificationsByUser(ctx, repoFilters)
	if err != nil {
		s.logger.ErrorWithErr("Failed to fetch notifications", err)
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	return &NotificationListResponse{
		Notifications: result.Notifications,
		Pagination: PaginationInfo{
			Page:       result.CurrentPage,
			Limit:      filters.Limit,
			TotalPages: result.TotalPages,
			TotalItems: result.TotalItems,
			HasNext:    result.HasNext,
			HasPrev:    result.HasPrev,
		},
		UnreadCount: result.UnreadCount,
	}, nil
}

// MarkNotificationRead marks a notification as read
func (s *communicationsService) MarkNotificationRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	if notificationID == uuid.Nil {
		return fmt.Errorf("notification ID is required")
	}

	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	notification, err := s.repo.FindNotificationByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotificationNotFound) {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to fetch notification: %w", err)
	}

	if notification.RecipientID != userID {
		return ErrUnauthorized
	}

	if err := s.repo.MarkNotificationAsRead(ctx, notificationID); err != nil {
		s.logger.ErrorWithErr("Failed to mark notification as read", err)
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	s.logger.Info("Notification marked as read")

	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *communicationsService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	if userID == uuid.Nil {
		return 0, fmt.Errorf("user ID is required")
	}

	count, err := s.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		s.logger.ErrorWithErr("Failed to get unread count", err)
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// BulkCreateNotifications creates multiple notifications at once
func (s *communicationsService) BulkCreateNotifications(ctx context.Context, inputs []CreateNotificationInput) error {
	if len(inputs) == 0 {
		return fmt.Errorf("notifications list cannot be empty")
	}

	notifications := make([]models.Notification, 0, len(inputs))
	for _, input := range inputs {
		if input.RecipientID == uuid.Nil {
			return fmt.Errorf("recipient ID is required")
		}
		if input.Title == "" || input.Content == "" {
			return fmt.Errorf("title and content are required")
		}

		notification := models.Notification{
			RecipientID:       input.RecipientID,
			NotificationType:  input.NotificationType,
			Title:             input.Title,
			Content:           input.Content,
			ActionURL:         input.ActionURL,
			Priority:          input.Priority,
			RelatedEntityType: input.RelatedEntityType,
			RelatedEntityID:   input.RelatedEntityID,
			DeliveryMethod:    input.DeliveryMethod,
			SentAt:            time.Now(),
		}

		if notification.Priority == "" {
			notification.Priority = models.NotificationPriorityNormal
		}
		if notification.DeliveryMethod == "" {
			notification.DeliveryMethod = models.NotificationDeliveryMethodInApp
		}

		notifications = append(notifications, notification)
	}

	if err := s.repo.BulkCreateNotifications(ctx, notifications); err != nil {
		s.logger.ErrorWithErr("Failed to bulk create notifications", err)
		return fmt.Errorf("failed to bulk create notifications: %w", err)
	}

	s.logger.Info("Bulk notifications created successfully")

	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
