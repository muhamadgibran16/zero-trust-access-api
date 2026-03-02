package service

import (
	"log"

	"github.com/gibran/go-gin-boilerplate/internal/model"
	repository "github.com/gibran/go-gin-boilerplate/internal/repository/notification"
	userRepo "github.com/gibran/go-gin-boilerplate/internal/repository/user"
	"github.com/google/uuid"
)

// NotificationService handles notification business logic
type NotificationService struct {
	repo     *repository.NotificationRepository
	userRepo *userRepo.UserRepository
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(repo *repository.NotificationRepository, uRepo *userRepo.UserRepository) *NotificationService {
	return &NotificationService{repo: repo, userRepo: uRepo}
}

// GetUserNotifications returns notifications for a user
func (s *NotificationService) GetUserNotifications(userID uuid.UUID, limit int) ([]model.Notification, error) {
	return s.repo.FindByUserID(userID, limit)
}

// GetUnreadCount returns unread notification count for a user
func (s *NotificationService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	return s.repo.CountUnread(userID)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(id uuid.UUID) error {
	return s.repo.MarkAsRead(id)
}

// NotifyAdmins creates a notification for all admin users
func (s *NotificationService) NotifyAdmins(title, message, notifType string) {
	admins, _, err := s.userRepo.FindAll(1, 100)
	if err != nil {
		log.Printf("[NotificationService] Failed to fetch admins: %v", err)
		return
	}

	for _, admin := range admins {
		if admin.Role != model.RoleAdmin {
			continue
		}
		notif := &model.Notification{
			UserID:  admin.ID,
			Title:   title,
			Message: message,
			Type:    notifType,
		}
		if err := s.repo.Create(notif); err != nil {
			log.Printf("[NotificationService] Failed to create notification for %s: %v", admin.Email, err)
		}
	}
}

// NotifyUser creates a notification for a specific user
func (s *NotificationService) NotifyUser(userID uuid.UUID, title, message, notifType string) {
	notif := &model.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
	}
	if err := s.repo.Create(notif); err != nil {
		log.Printf("[NotificationService] Failed to create notification: %v", err)
	}
}
