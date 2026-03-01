package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Device represents a registered corporate device for trusted MDM access
type Device struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	MacAddress string         `gorm:"size:255;not null;uniqueIndex" json:"macAddress"` // Used as unique identifier for bare-metal
	Name       string         `gorm:"size:255;not null" json:"name"`
	CertThumb  string         `gorm:"size:64;uniqueIndex" json:"certThumbprint"` // Mock thumbprint for cryptographic auth
	IsApproved bool           `gorm:"default:false" json:"isApproved"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	
	User       User           `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate is a GORM hook to set the UUID before creating a record
func (d *Device) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New()
	return
}
