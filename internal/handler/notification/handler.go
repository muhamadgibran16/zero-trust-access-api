package notification

import (
	"fmt"

	service "github.com/gibran/go-gin-boilerplate/internal/service/notification"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles notification-related requests
type Handler struct {
	service *service.NotificationService
}

// NewHandler creates a new Notification Handler
func NewHandler(s *service.NotificationService) *Handler {
	return &Handler{service: s}
}

// GetNotifications returns the current user's notifications
func (h *Handler) GetNotifications(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := uuid.Parse(fmt.Sprintf("%v", userID))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	notifs, err := h.service.GetUserNotifications(id, 20)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch notifications")
		return
	}

	response.Success(c, "Notifications loaded", notifs)
}

// GetUnreadCount returns the unread notification count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := uuid.Parse(fmt.Sprintf("%v", userID))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	count, err := h.service.GetUnreadCount(id)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch unread count")
		return
	}

	response.Success(c, "Unread count", gin.H{"count": count})
}

// MarkAsRead marks a specific notification as read
func (h *Handler) MarkAsRead(c *gin.Context) {
	notifID := c.Param("id")
	id, err := uuid.Parse(notifID)
	if err != nil {
		response.BadRequest(c, "Invalid notification ID")
		return
	}

	if err := h.service.MarkAsRead(id); err != nil {
		response.InternalServerError(c, "Failed to mark notification as read")
		return
	}

	response.Success(c, "Notification marked as read", nil)
}
