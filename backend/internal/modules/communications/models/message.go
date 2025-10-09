package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageType represents the type of message
type MessageType string

const (
	// MessageTypeDirect represents a direct message sent to specific recipients
	MessageTypeDirect MessageType = "direct"
	// MessageTypeBroadcast represents a broadcast message sent to all event registrants
	MessageTypeBroadcast MessageType = "broadcast"
)

// Message represents a communication record between users
// Messages can be direct (to specific recipients) or broadcast (to all event volunteers)
type Message struct {
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SenderID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"sender_id"`
	OpportunityID *uuid.UUID  `gorm:"type:uuid;index" json:"opportunity_id,omitempty"`
	MessageType   MessageType `gorm:"type:varchar(20);not null" json:"message_type"`
	Subject       *string     `gorm:"type:varchar(255)" json:"subject,omitempty"`
	Content       string      `gorm:"type:text;not null" json:"content"`
	SentAt        time.Time   `gorm:"not null;index" json:"sent_at"`
	CreatedAt     time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName specifies the table name for the Message model
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate hook to generate UUID and set sent_at timestamp
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	// Set sent_at to current time if not set
	if m.SentAt.IsZero() {
		m.SentAt = time.Now()
	}

	return nil
}

// Validate checks if the message data is valid
func (m *Message) Validate() error {
	if m.SenderID == uuid.Nil {
		return ErrInvalidSenderID
	}

	if m.Content == "" {
		return ErrEmptyMessageContent
	}

	if m.MessageType != MessageTypeDirect && m.MessageType != MessageTypeBroadcast {
		return ErrInvalidMessageType
	}

	// Broadcast messages must have an opportunity_id
	if m.MessageType == MessageTypeBroadcast && m.OpportunityID == nil {
		return ErrBroadcastWithoutOpportunity
	}

	return nil
}
