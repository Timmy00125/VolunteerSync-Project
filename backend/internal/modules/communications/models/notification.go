package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// NotificationTypeRegistrationConfirmed is sent when a volunteer registers for an event
	NotificationTypeRegistrationConfirmed NotificationType = "registration_confirmed"
	// NotificationTypeEventReminder is sent before an event starts
	NotificationTypeEventReminder NotificationType = "event_reminder"
	// NotificationTypeHoursLogged is sent when volunteer hours are logged
	NotificationTypeHoursLogged NotificationType = "hours_logged"
	// NotificationTypeMessageReceived is sent when a message is received
	NotificationTypeMessageReceived NotificationType = "message_received"
	// NotificationTypeEventCancelled is sent when an event is cancelled
	NotificationTypeEventCancelled NotificationType = "event_cancelled"
	// NotificationTypeWaitlistNotification is sent when a spot opens up from waitlist
	NotificationTypeWaitlistNotification NotificationType = "waitlist_notification"
	// NotificationTypeAchievementEarned is sent when a volunteer earns an achievement
	NotificationTypeAchievementEarned NotificationType = "achievement_earned"
	// NotificationTypeHoursDisputed is sent when volunteer hours are disputed
	NotificationTypeHoursDisputed NotificationType = "hours_disputed"
	// NotificationTypeHoursVerified is sent when volunteer hours are verified
	NotificationTypeHoursVerified NotificationType = "hours_verified"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	// NotificationPriorityLow represents low priority notifications
	NotificationPriorityLow NotificationPriority = "low"
	// NotificationPriorityNormal represents normal priority notifications
	NotificationPriorityNormal NotificationPriority = "normal"
	// NotificationPriorityHigh represents high priority notifications
	NotificationPriorityHigh NotificationPriority = "high"
	// NotificationPriorityCritical represents critical priority notifications (always shown)
	NotificationPriorityCritical NotificationPriority = "critical"
)

// NotificationDeliveryMethod represents how the notification was delivered
type NotificationDeliveryMethod string

const (
	// NotificationDeliveryMethodInApp represents in-app notifications
	NotificationDeliveryMethodInApp NotificationDeliveryMethod = "in_app"
	// NotificationDeliveryMethodBrowserPush represents browser push notifications
	NotificationDeliveryMethodBrowserPush NotificationDeliveryMethod = "browser_push"
)

// Notification represents a system-generated notification for a user
// Notifications are created for various events like registrations, reminders, messages, etc.
type Notification struct {
	ID                uuid.UUID                  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RecipientID       uuid.UUID                  `gorm:"type:uuid;not null;index:idx_notification_recipient" json:"recipient_id"`
	NotificationType  NotificationType           `gorm:"type:varchar(50);not null;index" json:"notification_type"`
	Title             string                     `gorm:"type:varchar(255);not null" json:"title"`
	Content           string                     `gorm:"type:text;not null" json:"content"`
	ActionURL         *string                    `gorm:"type:varchar(500)" json:"action_url,omitempty"`
	Priority          NotificationPriority       `gorm:"type:varchar(20);not null;default:'normal'" json:"priority"`
	RelatedEntityType *string                    `gorm:"type:varchar(50)" json:"related_entity_type,omitempty"`
	RelatedEntityID   *uuid.UUID                 `gorm:"type:uuid" json:"related_entity_id,omitempty"`
	ReadAt            *time.Time                 `gorm:"index:idx_notification_read" json:"read_at,omitempty"`
	DeliveredAt       *time.Time                 `gorm:"type:timestamp" json:"delivered_at,omitempty"`
	DeliveryMethod    NotificationDeliveryMethod `gorm:"type:varchar(20);not null;default:'in_app'" json:"delivery_method"`
	SentAt            time.Time                  `gorm:"not null;index" json:"sent_at"`
	CreatedAt         time.Time                  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time                  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName specifies the table name for the Notification model
func (Notification) TableName() string {
	return "notifications"
}

// BeforeCreate hook to generate UUID and set sent_at timestamp
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}

	// Set sent_at to current time if not set
	if n.SentAt.IsZero() {
		n.SentAt = time.Now()
	}

	// Set default priority if not set
	if n.Priority == "" {
		n.Priority = NotificationPriorityNormal
	}

	// Set default delivery method if not set
	if n.DeliveryMethod == "" {
		n.DeliveryMethod = NotificationDeliveryMethodInApp
	}

	return nil
}

// Validate checks if the notification data is valid
func (n *Notification) Validate() error {
	if n.RecipientID == uuid.Nil {
		return ErrInvalidRecipientID
	}

	if n.Title == "" || n.Content == "" {
		return ErrEmptyNotificationContent
	}

	// Validate notification type
	validTypes := []NotificationType{
		NotificationTypeRegistrationConfirmed,
		NotificationTypeEventReminder,
		NotificationTypeHoursLogged,
		NotificationTypeMessageReceived,
		NotificationTypeEventCancelled,
		NotificationTypeWaitlistNotification,
		NotificationTypeAchievementEarned,
		NotificationTypeHoursDisputed,
		NotificationTypeHoursVerified,
	}

	validType := false
	for _, vt := range validTypes {
		if n.NotificationType == vt {
			validType = true
			break
		}
	}
	if !validType {
		return ErrInvalidNotificationType
	}

	// Validate priority
	validPriorities := []NotificationPriority{
		NotificationPriorityLow,
		NotificationPriorityNormal,
		NotificationPriorityHigh,
		NotificationPriorityCritical,
	}

	validPriority := false
	for _, vp := range validPriorities {
		if n.Priority == vp {
			validPriority = true
			break
		}
	}
	if !validPriority {
		return ErrInvalidNotificationPriority
	}

	// Validate delivery method
	validMethods := []NotificationDeliveryMethod{
		NotificationDeliveryMethodInApp,
		NotificationDeliveryMethodBrowserPush,
	}

	validMethod := false
	for _, vm := range validMethods {
		if n.DeliveryMethod == vm {
			validMethod = true
			break
		}
	}
	if !validMethod {
		return ErrInvalidDeliveryMethod
	}

	return nil
}

// MarkAsRead marks the notification as read by setting the ReadAt timestamp
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.ReadAt = &now
}

// IsRead returns true if the notification has been read
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

// MarkAsDelivered marks the notification as delivered
func (n *Notification) MarkAsDelivered() {
	now := time.Now()
	n.DeliveredAt = &now
}

// IsDelivered returns true if the notification has been delivered
func (n *Notification) IsDelivered() bool {
	return n.DeliveredAt != nil
}

// IsCritical returns true if the notification is critical priority
func (n *Notification) IsCritical() bool {
	return n.Priority == NotificationPriorityCritical
}
