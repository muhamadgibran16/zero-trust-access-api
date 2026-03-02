package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	NotifTypeDeviceRequest  = "DEVICE_REQUEST"
	NotifTypeHighRisk       = "HIGH_RISK"
	NotifTypeSessionRevoked = "SESSION_REVOKED"
)

// Notification represents an in-app notification
type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Type      string         `gorm:"size:50;not null" json:"type"`
	IsRead    bool           `gorm:"default:false" json:"isRead"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate sets UUID before inserting
func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	n.ID = uuid.New()
	return
}
