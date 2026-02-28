package service

import (
	repository "github.com/gibran/go-gin-boilerplate/internal/repository/audit"
	"github.com/gibran/go-gin-boilerplate/internal/model"
)

// AuditLogService handles business logic for audit logs
type AuditLogService struct {
	repo *repository.AuditLogRepository
}

// NewAuditLogService creates a new AuditLogService
func NewAuditLogService(repo *repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

// GetAuditLogs returns a paginated list of audit logs
func (s *AuditLogService) GetAuditLogs(page, perPage int) ([]*model.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50 // Default max limit
	}
	
	offset := (page - 1) * perPage

	return s.repo.FindAll(offset, perPage)
}
