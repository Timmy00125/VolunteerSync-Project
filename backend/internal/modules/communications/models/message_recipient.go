package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRecipient represents a recipient of a direct message
// This is a junction table that tracks read status for each recipient
type MessageRecipient struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MessageID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"message_id"`
	RecipientID uuid.UUID  `gorm:"type:uuid;not null;index" json:"recipient_id"`
	ReadAt      *time.Time `gorm:"index" json:"read_at,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName specifies the table name for the MessageRecipient model
func (MessageRecipient) TableName() string {
	return "message_recipients"
}

// BeforeCreate hook to generate UUID
func (mr *MessageRecipient) BeforeCreate(tx *gorm.DB) error {
	if mr.ID == uuid.Nil {
		mr.ID = uuid.New()
	}
	return nil
}

// MarkAsRead marks the message as read by setting the ReadAt timestamp
func (mr *MessageRecipient) MarkAsRead() {
	now := time.Now()
	mr.ReadAt = &now
}

// IsRead returns true if the message has been read
func (mr *MessageRecipient) IsRead() bool {
	return mr.ReadAt != nil
}
