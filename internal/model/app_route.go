package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppRoute maps a proxy path to an internal target service
type AppRoute struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	PathPrefix  string         `gorm:"size:255;not null;uniqueIndex" json:"pathPrefix"` // e.g., "/hr-app"
	TargetURL   string         `gorm:"size:255;not null" json:"targetUrl"`  // e.g., "http://internal-hr.local:8080"
	Icon        string         `gorm:"size:255" json:"icon"` 
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate is a GORM hook to set the UUID before creating a record
func (a *AppRoute) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New()
	return
}
