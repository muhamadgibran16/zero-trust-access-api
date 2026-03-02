package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PasswordReset represents a password reset token record
type PasswordReset struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	Token     string         `gorm:"size:255;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time      `gorm:"not null" json:"expiresAt"`
	Used      bool           `gorm:"default:false" json:"used"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate sets UUID before inserting
func (p *PasswordReset) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}
