package repository

import (
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"gorm.io/gorm"
)

// AuditLogRepository handles operations for the AuditLog model
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// FindAll retrieves a list of audit logs with pagination
func (r *AuditLogRepository) FindAll(offset int, limit int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	// Get total count
	if err := r.db.Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results ordered by newest first
	if err := r.db.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
