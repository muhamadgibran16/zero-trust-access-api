package analytics

import (
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler handles analytics requests
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new Analytics Handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// GetDashboardMetrics returns aggregated dashboard metrics
func (h *Handler) GetDashboardMetrics(c *gin.Context) {
	var totalUsers int64
	h.db.Model(&model.User{}).Count(&totalUsers)

	var totalDevices int64
	h.db.Model(&model.Device{}).Count(&totalDevices)

	var approvedDevices int64
	h.db.Model(&model.Device{}).Where("is_approved = true").Count(&approvedDevices)

	var totalAuditLogs int64
	h.db.Model(&model.AuditLog{}).Count(&totalAuditLogs)

	var blockedRequests int64
	h.db.Model(&model.AuditLog{}).Where("status >= 400").Count(&blockedRequests)

	var activePolicies int64
	h.db.Model(&model.PolicyRule{}).Where("is_active = true").Count(&activePolicies)

	response.Success(c, "Dashboard metrics", gin.H{
		"totalUsers":       totalUsers,
		"totalDevices":     totalDevices,
		"approvedDevices":  approvedDevices,
		"totalAuditLogs":   totalAuditLogs,
		"blockedRequests":  blockedRequests,
		"activePolicies":   activePolicies,
	})
}
