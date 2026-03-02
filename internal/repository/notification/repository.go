package repository

import (
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository handles database operations for Notifications
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// FindByUserID retrieves notifications for a specific user
func (r *NotificationRepository) FindByUserID(userID uuid.UUID, limit int) ([]model.Notification, error) {
	var notifs []model.Notification
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&notifs).Error
	return notifs, err
}

// CountUnread returns the number of unread notifications for a user
func (r *NotificationRepository) CountUnread(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).Where("user_id = ? AND is_read = false", userID).Count(&count).Error
	return count, err
}

// Create inserts a new notification
func (r *NotificationRepository) Create(notif *model.Notification) error {
	return r.db.Create(notif).Error
}

// MarkAsRead sets a notification as read
func (r *NotificationRepository) MarkAsRead(id uuid.UUID) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}
