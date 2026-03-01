package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PolicyTypeDenyIP   = "DENY_IP"
	PolicyTypeAllowIP  = "ALLOW_IP"
	PolicyTypeRequire  = "REQUIRE_ROLE"
	PolicyTypeTime     = "TIME_RESTRICT"
	PolicyTypeGeo      = "GEO_RESTRICT"
)

// PolicyRule defines dynamic access rules evaluated by the zero trust engine
type PolicyRule struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Type      string    `gorm:"size:50;not null;index" json:"type"`
	Value     string    `gorm:"size:255;not null" json:"value"` // e.g. "198.51.100.1"
	Resource  string    `gorm:"size:255;index" json:"resource"` // optional prefix path, e.g. "/api/v1/admin". Empty means global.
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate is a GORM hook to set the UUID before creating
func (p *PolicyRule) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}
