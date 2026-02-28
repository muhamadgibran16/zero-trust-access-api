package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog represents an access log entry for the Zero Trust Audit feature
type AuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID    string    `gorm:"size:255;index" json:"userId"` // Can be system or empty for unauthenticated
	Action    string    `gorm:"size:255" json:"action"`
	Method    string    `gorm:"size:10" json:"method"`
	Path      string    `gorm:"size:255" json:"path"`
	IPAddress string    `gorm:"size:255" json:"ipAddress"`
	UserAgent string    `gorm:"text" json:"userAgent"`
	Status    int       `json:"status"`
	Details   string    `gorm:"type:text" json:"details"` // JSON string or text for extra info (e.g. device posture failure)
	CreatedAt time.Time `json:"createdAt"`
}

// BeforeCreate is a GORM hook to set the UUID before creating
func (a *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New()
	return
}
