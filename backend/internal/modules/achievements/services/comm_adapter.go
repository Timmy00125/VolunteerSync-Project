package services

import (
	"context"

	commModels "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/models"
	commServices "github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/services"
)

// communicationsServiceAdapter adapts the communications module's service to satisfy
// the CommunicationsService interface expected by the achievements service.
// This maintains module boundaries while enabling cross-module communication.
type communicationsServiceAdapter struct {
	commService commServices.CommunicationsService
}

// NewCommunicationsServiceAdapter creates a new adapter for the communications service
func NewCommunicationsServiceAdapter(commService commServices.CommunicationsService) CommunicationsService {
	return &communicationsServiceAdapter{
		commService: commService,
	}
}

// CreateNotification creates a notification by adapting the input to the communications service format
func (a *communicationsServiceAdapter) CreateNotification(ctx context.Context, input CreateNotificationInput) error {
	// Map our input to the communications service input
	commInput := commServices.CreateNotificationInput{
		RecipientID:       input.RecipientID,
		NotificationType:  mapNotificationType(input.NotificationType),
		Title:             input.Title,
		Content:           input.Content,
		ActionURL:         input.ActionURL,
		Priority:          mapNotificationPriority(input.Priority),
		RelatedEntityType: input.RelatedEntityType,
		RelatedEntityID:   input.RelatedEntityID,
		DeliveryMethod:    mapDeliveryMethod(input.DeliveryMethod),
	}

	// Create the notification
	_, err := a.commService.CreateNotification(ctx, commInput)
	return err
}

// mapNotificationType maps string notification type to the communications model type
func mapNotificationType(notificationType string) commModels.NotificationType {
	switch notificationType {
	case "achievement_earned":
		return commModels.NotificationTypeAchievementEarned
	case "event_reminder":
		return commModels.NotificationTypeEventReminder
	case "hours_logged":
		return commModels.NotificationTypeHoursLogged
	case "hours_verified":
		return commModels.NotificationTypeHoursVerified
	case "hours_disputed":
		return commModels.NotificationTypeHoursDisputed
	case "message_received":
		return commModels.NotificationTypeMessageReceived
	case "waitlist_notification":
		return commModels.NotificationTypeWaitlistNotification
	case "registration_confirmed":
		return commModels.NotificationTypeRegistrationConfirmed
	case "event_cancelled":
		return commModels.NotificationTypeEventCancelled
	default:
		return commModels.NotificationTypeRegistrationConfirmed // default fallback
	}
}

// mapNotificationPriority maps string priority to the communications model priority
func mapNotificationPriority(priority string) commModels.NotificationPriority {
	switch priority {
	case "low":
		return commModels.NotificationPriorityLow
	case "normal", "medium":
		return commModels.NotificationPriorityNormal
	case "high":
		return commModels.NotificationPriorityHigh
	case "critical", "urgent":
		return commModels.NotificationPriorityCritical
	default:
		return commModels.NotificationPriorityNormal
	}
}

// mapDeliveryMethod maps string delivery method to the communications model delivery method
func mapDeliveryMethod(deliveryMethod string) commModels.NotificationDeliveryMethod {
	switch deliveryMethod {
	case "in_app":
		return commModels.NotificationDeliveryMethodInApp
	case "browser_push":
		return commModels.NotificationDeliveryMethodBrowserPush
	default:
		return commModels.NotificationDeliveryMethodInApp
	}
}
