package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/middleware"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/models"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/communications/services"
	apperrors "github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/errors"
	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/pkg/logger"
)

// CommunicationsHandler exposes HTTP handlers for communications (messages and notifications)
type CommunicationsHandler struct {
	service services.CommunicationsService
	log     *logger.Logger
}

// NewCommunicationsHandler constructs a CommunicationsHandler with required dependencies
func NewCommunicationsHandler(service services.CommunicationsService, log *logger.Logger) (*CommunicationsHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("communications handler requires communications service")
	}

	if log == nil {
		log = logger.Get()
	}

	return &CommunicationsHandler{
		service: service,
		log:     log,
	}, nil
}

// RegisterRoutes wires communications routes under the provided router group
func (h *CommunicationsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}

	// All routes require authentication
	// These would be added to a group with auth middleware
	// authenticated := rg.Group("")
	// authenticated.Use(authMiddleware)
	rg.POST("/messages", h.CreateMessage)                         // POST /messages
	rg.GET("/notifications", h.ListNotifications)                 // GET /notifications
	rg.PATCH("/notifications/:id/read", h.MarkNotificationAsRead) // PATCH /notifications/:id/read
	rg.GET("/notifications/unread-count", h.GetUnreadCount)       // GET /notifications/unread-count
}

// Request/Response DTOs

// createMessageRequest represents the request body for creating a message
type createMessageRequest struct {
	MessageType   string   `json:"message_type" binding:"required"` // "direct" or "broadcast"
	RecipientIDs  []string `json:"recipient_ids,omitempty"`         // Required for direct messages
	OpportunityID *string  `json:"opportunity_id,omitempty"`        // Required for broadcast messages
	Subject       *string  `json:"subject,omitempty"`
	Content       string   `json:"content" binding:"required"`
}

// messageResponse represents a message in responses
type messageResponse struct {
	ID            string  `json:"id"`
	SenderID      string  `json:"sender_id"`
	OpportunityID *string `json:"opportunity_id,omitempty"`
	MessageType   string  `json:"message_type"`
	Subject       *string `json:"subject,omitempty"`
	Content       string  `json:"content"`
	SentAt        string  `json:"sent_at"`
}

// notificationResponse represents a notification in responses
type notificationResponse struct {
	ID                string  `json:"id"`
	RecipientID       string  `json:"recipient_id"`
	NotificationType  string  `json:"notification_type"`
	Title             string  `json:"title"`
	Content           string  `json:"content"`
	ActionURL         *string `json:"action_url,omitempty"`
	Priority          string  `json:"priority"`
	RelatedEntityType *string `json:"related_entity_type,omitempty"`
	RelatedEntityID   *string `json:"related_entity_id,omitempty"`
	ReadAt            *string `json:"read_at,omitempty"`
	DeliveredAt       *string `json:"delivered_at,omitempty"`
	DeliveryMethod    string  `json:"delivery_method"`
	SentAt            string  `json:"sent_at"`
	CreatedAt         string  `json:"created_at"`
}

// paginationResponse represents pagination metadata
type paginationResponse struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalPages int  `json:"total_pages"`
	TotalItems int  `json:"total_items"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// CreateMessage handles POST /messages
// Creates a direct message or broadcast message
func (h *CommunicationsHandler) CreateMessage(c *gin.Context) {
	var req createMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid request payload").WithError(err))
		return
	}

	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Validate message type and create appropriate message
	switch req.MessageType {
	case "direct":
		// Direct message requires recipient IDs
		if len(req.RecipientIDs) == 0 {
			h.respondWithError(c, apperrors.NewBadRequestError("direct message requires at least one recipient"))
			return
		}

		// Parse recipient IDs
		recipientIDs := make([]uuid.UUID, 0, len(req.RecipientIDs))
		for _, idStr := range req.RecipientIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				h.respondWithError(c, apperrors.NewValidationError("invalid recipient ID format", map[string]interface{}{
					"recipient_id": idStr,
				}))
				return
			}
			recipientIDs = append(recipientIDs, id)
		}

		// Send direct message
		input := services.SendDirectMessageInput{
			SenderID:     userUUID,
			RecipientIDs: recipientIDs,
			Subject:      req.Subject,
			Content:      req.Content,
		}

		message, err := h.service.SendDirectMessage(ctx, input)
		if err != nil {
			h.respondWithError(c, h.mapServiceError(err))
			return
		}

		c.JSON(http.StatusCreated, h.toMessageResponse(message))

	case "broadcast":
		// Broadcast message requires opportunity ID
		if req.OpportunityID == nil || *req.OpportunityID == "" {
			h.respondWithError(c, apperrors.NewBadRequestError("broadcast message requires opportunity_id"))
			return
		}

		opportunityID, err := uuid.Parse(*req.OpportunityID)
		if err != nil {
			h.respondWithError(c, apperrors.NewValidationError("invalid opportunity ID format", map[string]interface{}{
				"opportunity_id": *req.OpportunityID,
			}))
			return
		}

		// Send broadcast message
		input := services.SendBroadcastMessageInput{
			SenderID:      userUUID,
			OpportunityID: opportunityID,
			Subject:       req.Subject,
			Content:       req.Content,
		}

		message, err := h.service.SendBroadcastMessage(ctx, input)
		if err != nil {
			h.respondWithError(c, h.mapServiceError(err))
			return
		}

		c.JSON(http.StatusCreated, h.toMessageResponse(message))

	default:
		h.respondWithError(c, apperrors.NewBadRequestError("invalid message_type: must be 'direct' or 'broadcast'"))
	}
}

// ListNotifications handles GET /notifications
// Retrieves paginated notifications for the authenticated user with filters
func (h *CommunicationsHandler) ListNotifications(c *gin.Context) {
	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Parse unread filter
	var unread *bool
	if unreadStr := c.Query("unread"); unreadStr != "" {
		if unreadStr == "true" {
			t := true
			unread = &t
		} else if unreadStr == "false" {
			f := false
			unread = &f
		}
	}

	// Parse notification type filter
	var notificationType *models.NotificationType
	if typeStr := c.Query("type"); typeStr != "" {
		nt := models.NotificationType(typeStr)
		notificationType = &nt
	}

	// Parse priority filter
	var priority *models.NotificationPriority
	if priorityStr := c.Query("priority"); priorityStr != "" {
		p := models.NotificationPriority(priorityStr)
		priority = &p
	}

	// Build filters
	filters := services.GetNotificationsFilters{
		UserID:           userUUID,
		Unread:           unread,
		NotificationType: notificationType,
		Priority:         priority,
		Page:             page,
		Limit:            limit,
	}

	// Fetch notifications
	result, err := h.service.GetUserNotifications(ctx, filters)
	if err != nil {
		h.respondWithError(c, h.mapServiceError(err))
		return
	}

	// Convert to response
	notifications := make([]notificationResponse, len(result.Notifications))
	for i, n := range result.Notifications {
		notifications[i] = h.toNotificationResponse(&n)
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"pagination": paginationResponse{
			Page:       result.Pagination.Page,
			Limit:      result.Pagination.Limit,
			TotalPages: result.Pagination.TotalPages,
			TotalItems: result.Pagination.TotalItems,
			HasNext:    result.Pagination.HasNext,
			HasPrev:    result.Pagination.HasPrev,
		},
		"unread_count": result.UnreadCount,
	})
}

// MarkNotificationAsRead handles PATCH /notifications/:id/read
// Marks a notification as read for the authenticated user
func (h *CommunicationsHandler) MarkNotificationAsRead(c *gin.Context) {
	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	// Parse notification ID from URL
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		h.respondWithError(c, apperrors.NewBadRequestError("invalid notification ID format"))
		return
	}

	ctx := c.Request.Context()

	// Mark notification as read
	if err := h.service.MarkNotificationRead(ctx, notificationID, userUUID); err != nil {
		h.respondWithError(c, h.mapServiceError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read successfully",
	})
}

// GetUnreadCount handles GET /notifications/unread-count
// Returns the count of unread notifications for the authenticated user
func (h *CommunicationsHandler) GetUnreadCount(c *gin.Context) {
	// Get authenticated user UUID from context (set by auth and context enrichment middleware)
	userUUID := middleware.MustGetUserUUID(c)

	ctx := c.Request.Context()

	count, err := h.service.GetUnreadCount(ctx, userUUID)
	if err != nil {
		h.respondWithError(c, h.mapServiceError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// Helper methods

// respondWithError sends an error response
func (h *CommunicationsHandler) respondWithError(c *gin.Context, err *apperrors.AppError) {
	h.log.ErrorWithErr("Request failed", err)
	c.JSON(err.HTTPStatus, gin.H{
		"error": gin.H{
			"message": err.Message,
			"code":    err.Code,
			"details": err.Details,
		},
	})
}

// mapServiceError maps service errors to appropriate HTTP errors
func (h *CommunicationsHandler) mapServiceError(err error) *apperrors.AppError {
	switch err {
	case services.ErrMissingRecipients:
		return apperrors.NewBadRequestError("direct message must have at least one recipient")
	case services.ErrMissingOpportunityID:
		return apperrors.NewBadRequestError("broadcast message must have an opportunity ID")
	case services.ErrEmptyMessageContent:
		return apperrors.NewBadRequestError("message content is required")
	case services.ErrNotificationNotFound:
		return apperrors.NewNotFoundError("notification")
	case services.ErrUnauthorized:
		return apperrors.NewForbiddenError("you don't have permission to access this resource")
	default:
		return apperrors.NewInternalServerError("an unexpected error occurred").WithError(err)
	}
}

// toMessageResponse converts a message model to response DTO
func (h *CommunicationsHandler) toMessageResponse(message *models.Message) messageResponse {
	var opportunityID *string
	if message.OpportunityID != nil {
		idStr := message.OpportunityID.String()
		opportunityID = &idStr
	}

	return messageResponse{
		ID:            message.ID.String(),
		SenderID:      message.SenderID.String(),
		OpportunityID: opportunityID,
		MessageType:   string(message.MessageType),
		Subject:       message.Subject,
		Content:       message.Content,
		SentAt:        message.SentAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// toNotificationResponse converts a notification model to response DTO
func (h *CommunicationsHandler) toNotificationResponse(notification *models.Notification) notificationResponse {
	var readAt, deliveredAt *string
	if notification.ReadAt != nil {
		t := notification.ReadAt.Format("2006-01-02T15:04:05Z07:00")
		readAt = &t
	}
	if notification.DeliveredAt != nil {
		t := notification.DeliveredAt.Format("2006-01-02T15:04:05Z07:00")
		deliveredAt = &t
	}

	var relatedEntityID *string
	if notification.RelatedEntityID != nil {
		idStr := notification.RelatedEntityID.String()
		relatedEntityID = &idStr
	}

	return notificationResponse{
		ID:                notification.ID.String(),
		RecipientID:       notification.RecipientID.String(),
		NotificationType:  string(notification.NotificationType),
		Title:             notification.Title,
		Content:           notification.Content,
		ActionURL:         notification.ActionURL,
		Priority:          string(notification.Priority),
		RelatedEntityType: notification.RelatedEntityType,
		RelatedEntityID:   relatedEntityID,
		ReadAt:            readAt,
		DeliveredAt:       deliveredAt,
		DeliveryMethod:    string(notification.DeliveryMethod),
		SentAt:            notification.SentAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:         notification.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
