package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Custom errors for repository operations
var (
	// ErrMessageNotFound is returned when a message cannot be found
	ErrMessageNotFound = errors.New("message not found")
	// ErrNotificationNotFound is returned when a notification cannot be found
	ErrNotificationNotFound = errors.New("notification not found")
	// ErrDatabaseOperation is returned when a database operation fails
	ErrDatabaseOperation = errors.New("database operation failed")
)

// NotificationFilters represents filters for listing notifications
type NotificationFilters struct {
	RecipientID      uuid.UUID                    // Required: User receiving notifications
	Unread           *bool                        // Optional: Filter by read status
	NotificationType *models.NotificationType     // Optional: Filter by type
	Priority         *models.NotificationPriority // Optional: Filter by priority
	Page             int                          // Page number (1-indexed)
	Limit            int                          // Items per page
}

// PaginatedNotifications represents a paginated list of notifications
type PaginatedNotifications struct {
	Notifications []models.Notification
	TotalItems    int
	TotalPages    int
	CurrentPage   int
	HasNext       bool
	HasPrev       bool
	UnreadCount   int // Total unread count across all pages
}

// CommunicationsRepository defines the interface for communications data access
// Following Clean Architecture principles, this interface allows for dependency injection
// and makes the service layer independent of database implementation details
type CommunicationsRepository interface {
	// CreateMessage creates a new message in the database
	// Returns the created message or an error if the operation fails
	CreateMessage(ctx context.Context, message *models.Message) error

	// CreateMessageRecipients creates multiple message recipients in a single transaction
	// Used for direct messages to multiple recipients
	CreateMessageRecipients(ctx context.Context, recipients []models.MessageRecipient) error

	// FindMessageByID retrieves a message by its unique identifier
	// Returns ErrMessageNotFound if no message exists with the given ID
	FindMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error)

	// FindMessagesByRecipient retrieves all messages for a specific recipient
	FindMessagesByRecipient(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]models.Message, int, error)

	// MarkMessageAsRead marks a message as read for a specific recipient
	MarkMessageAsRead(ctx context.Context, messageID, recipientID uuid.UUID) error

	// GetUnreadMessageCount returns the count of unread messages for a recipient
	GetUnreadMessageCount(ctx context.Context, recipientID uuid.UUID) (int, error)

	// CreateNotification creates a new notification in the database
	// Returns the created notification or an error if the operation fails
	CreateNotification(ctx context.Context, notification *models.Notification) error

	// FindNotificationByID retrieves a notification by its unique identifier
	// Returns ErrNotificationNotFound if no notification exists with the given ID
	FindNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)

	// FindNotificationsByUser retrieves notifications for a user with filters and pagination
	// Supports filtering by read status, type, and priority
	FindNotificationsByUser(ctx context.Context, filters NotificationFilters) (*PaginatedNotifications, error)

	// MarkNotificationAsRead marks a notification as read by setting the ReadAt timestamp
	MarkNotificationAsRead(ctx context.Context, notificationID uuid.UUID) error

	// GetUnreadCount returns the count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)

	// DeleteOldNotifications deletes notifications older than a specified date
	// Used for data retention policies (cleanup job)
	DeleteOldNotifications(ctx context.Context, beforeDate string) error

	// BulkCreateNotifications creates multiple notifications in a single transaction
	// Used for broadcast notifications to multiple users
	BulkCreateNotifications(ctx context.Context, notifications []models.Notification) error
}

// gormCommunicationsRepository is the GORM implementation of CommunicationsRepository
type gormCommunicationsRepository struct {
	db *gorm.DB
}

// NewCommunicationsRepository creates a new instance of CommunicationsRepository using GORM
func NewCommunicationsRepository(db *gorm.DB) CommunicationsRepository {
	return &gormCommunicationsRepository{
		db: db,
	}
}

// CreateMessage creates a new message in the database
func (r *gormCommunicationsRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	if err := message.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// CreateMessageRecipients creates multiple message recipients in a single transaction
func (r *gormCommunicationsRepository) CreateMessageRecipients(ctx context.Context, recipients []models.MessageRecipient) error {
	if len(recipients) == 0 {
		return fmt.Errorf("recipients list cannot be empty")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, recipient := range recipients {
			if err := tx.Create(&recipient).Error; err != nil {
				return fmt.Errorf("failed to create recipient: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create message recipients: %w", err)
	}

	return nil
}

// FindMessageByID retrieves a message by its unique identifier
func (r *gormCommunicationsRepository) FindMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid message ID")
	}

	var message models.Message
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&message)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrMessageNotFound
		}
		return nil, fmt.Errorf("failed to find message: %w", result.Error)
	}

	return &message, nil
}

// FindMessagesByRecipient retrieves all messages for a specific recipient with pagination
func (r *gormCommunicationsRepository) FindMessagesByRecipient(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]models.Message, int, error) {
	if recipientID == uuid.Nil {
		return nil, 0, fmt.Errorf("invalid recipient ID")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var messages []models.Message
	var total int64

	// Query to get messages where user is a recipient
	query := r.db.WithContext(ctx).
		Joins("JOIN message_recipients ON message_recipients.message_id = messages.id").
		Where("message_recipients.recipient_id = ?", recipientID).
		Order("messages.sent_at DESC")

	// Get total count
	if err := query.Model(&models.Message{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find messages: %w", err)
	}

	return messages, int(total), nil
}

// MarkMessageAsRead marks a message as read for a specific recipient
func (r *gormCommunicationsRepository) MarkMessageAsRead(ctx context.Context, messageID, recipientID uuid.UUID) error {
	if messageID == uuid.Nil || recipientID == uuid.Nil {
		return fmt.Errorf("invalid message ID or recipient ID")
	}

	result := r.db.WithContext(ctx).
		Model(&models.MessageRecipient{}).
		Where("message_id = ? AND recipient_id = ? AND read_at IS NULL", messageID, recipientID).
		Update("read_at", gorm.Expr("NOW()"))

	if result.Error != nil {
		return fmt.Errorf("failed to mark message as read: %w", result.Error)
	}

	return nil
}

// GetUnreadMessageCount returns the count of unread messages for a recipient
func (r *gormCommunicationsRepository) GetUnreadMessageCount(ctx context.Context, recipientID uuid.UUID) (int, error) {
	if recipientID == uuid.Nil {
		return 0, fmt.Errorf("invalid recipient ID")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.MessageRecipient{}).
		Where("recipient_id = ? AND read_at IS NULL", recipientID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}

	return int(count), nil
}

// CreateNotification creates a new notification in the database
func (r *gormCommunicationsRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	if err := notification.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// FindNotificationByID retrieves a notification by its unique identifier
func (r *gormCommunicationsRepository) FindNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid notification ID")
	}

	var notification models.Notification
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&notification)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to find notification: %w", result.Error)
	}

	return &notification, nil
}

// FindNotificationsByUser retrieves notifications for a user with filters and pagination
func (r *gormCommunicationsRepository) FindNotificationsByUser(ctx context.Context, filters NotificationFilters) (*PaginatedNotifications, error) {
	if filters.RecipientID == uuid.Nil {
		return nil, fmt.Errorf("recipient ID is required")
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}

	offset := (filters.Page - 1) * filters.Limit

	var notifications []models.Notification
	var totalItems int64

	// Build base query
	query := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("recipient_id = ?", filters.RecipientID)

	// Apply filters
	if filters.Unread != nil && *filters.Unread {
		query = query.Where("read_at IS NULL")
	}

	if filters.NotificationType != nil {
		query = query.Where("notification_type = ?", *filters.NotificationType)
	}

	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}

	// Get total count
	if err := query.Count(&totalItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get paginated results
	if err := query.Order("sent_at DESC").Offset(offset).Limit(filters.Limit).Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to find notifications: %w", err)
	}

	// Get total unread count (across all pages)
	unreadCount, err := r.GetUnreadCount(ctx, filters.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread count: %w", err)
	}

	totalPages := int(totalItems) / filters.Limit
	if int(totalItems)%filters.Limit > 0 {
		totalPages++
	}

	return &PaginatedNotifications{
		Notifications: notifications,
		TotalItems:    int(totalItems),
		TotalPages:    totalPages,
		CurrentPage:   filters.Page,
		HasNext:       filters.Page < totalPages,
		HasPrev:       filters.Page > 1,
		UnreadCount:   unreadCount,
	}, nil
}

// MarkNotificationAsRead marks a notification as read by setting the ReadAt timestamp
func (r *gormCommunicationsRepository) MarkNotificationAsRead(ctx context.Context, notificationID uuid.UUID) error {
	if notificationID == uuid.Nil {
		return fmt.Errorf("invalid notification ID")
	}

	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND read_at IS NULL", notificationID).
		Update("read_at", gorm.Expr("NOW()"))

	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *gormCommunicationsRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	if userID == uuid.Nil {
		return 0, fmt.Errorf("invalid user ID")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("recipient_id = ? AND read_at IS NULL", userID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return int(count), nil
}

// DeleteOldNotifications deletes notifications older than a specified date
func (r *gormCommunicationsRepository) DeleteOldNotifications(ctx context.Context, beforeDate string) error {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", beforeDate).
		Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete old notifications: %w", result.Error)
	}

	return nil
}

// BulkCreateNotifications creates multiple notifications in a single transaction
func (r *gormCommunicationsRepository) BulkCreateNotifications(ctx context.Context, notifications []models.Notification) error {
	if len(notifications) == 0 {
		return fmt.Errorf("notifications list cannot be empty")
	}

	// Validate all notifications before creating
	for i := range notifications {
		if err := notifications[i].Validate(); err != nil {
			return fmt.Errorf("validation failed for notification %d: %w", i, err)
		}
	}

	// Use batch insert for better performance
	err := r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error
	if err != nil {
		return fmt.Errorf("failed to bulk create notifications: %w", err)
	}

	return nil
}
