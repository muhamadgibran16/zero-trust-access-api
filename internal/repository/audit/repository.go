package repository

import (
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"gorm.io/gorm"
)

// AuditLogWithUser is a response DTO that includes the username
type AuditLogWithUser struct {
	model.AuditLog
	Username string `json:"username"`
}

// AuditLogRepository handles operations for the AuditLog model
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// FindAll retrieves a list of audit logs with pagination, joining users for username
func (r *AuditLogRepository) FindAll(offset int, limit int) ([]AuditLogWithUser, int64, error) {
	var logs []AuditLogWithUser
	var total int64

	// Get total count
	if err := r.db.Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with LEFT JOIN on users to resolve username
	if err := r.db.Table("audit_logs").
		Select("audit_logs.*, COALESCE(users.name, 'System') as username").
		Joins("LEFT JOIN users ON audit_logs.user_id = users.id::text").
		Order("audit_logs.created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

