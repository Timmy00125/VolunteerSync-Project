package models

import "errors"

// Message-related errors
var (
	// ErrInvalidSenderID is returned when the sender ID is invalid
	ErrInvalidSenderID = errors.New("invalid sender ID")
	// ErrEmptyMessageContent is returned when the message content is empty
	ErrEmptyMessageContent = errors.New("message content is required")
	// ErrInvalidMessageType is returned when the message type is invalid
	ErrInvalidMessageType = errors.New("invalid message type")
	// ErrBroadcastWithoutOpportunity is returned when a broadcast message has no opportunity
	ErrBroadcastWithoutOpportunity = errors.New("broadcast message must have an opportunity ID")
)

// Notification-related errors
var (
	// ErrInvalidRecipientID is returned when the recipient ID is invalid
	ErrInvalidRecipientID = errors.New("invalid recipient ID")
	// ErrEmptyNotificationContent is returned when the notification content is empty
	ErrEmptyNotificationContent = errors.New("notification content is required")
	// ErrInvalidNotificationType is returned when the notification type is invalid
	ErrInvalidNotificationType = errors.New("invalid notification type")
	// ErrInvalidNotificationPriority is returned when the notification priority is invalid
	ErrInvalidNotificationPriority = errors.New("invalid notification priority")
	// ErrInvalidDeliveryMethod is returned when the delivery method is invalid
	ErrInvalidDeliveryMethod = errors.New("invalid delivery method")
)
