package model

import (
	"time"

	"github.com/google/uuid"
)

// TokenBlacklist represents an invalidated JWT JTI (Token ID) or revoked User sessions
type TokenBlacklist struct {
	ID        uint      `gorm:"primaryKey"`
	TokenID   string    `gorm:"index;type:varchar(255)" json:"tokenId"` // For single session logout
	UserID    uuid.UUID `gorm:"index;type:uuid" json:"userId"`          // For full user revocation
	ExpiresAt time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}
