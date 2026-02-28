package admin

import (
	"strconv"

	service "github.com/gibran/go-gin-boilerplate/internal/service/audit"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler handles admin-related requests
type Handler struct {
	auditService *service.AuditLogService
}

// NewHandler creates a new Admin Handler
func NewHandler(as *service.AuditLogService) *Handler {
	return &Handler{auditService: as}
}

// GetAuditLogs handles GET /admin/audit-logs
// @Summary Get audit logs
// @Description Retrieve paginated audit logs
// @Tags admin
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /admin/audit-logs [get]
func (h *Handler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "50"))

	logs, total, err := h.auditService.GetAuditLogs(page, perPage)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve audit logs")
		return
	}

	response.Success(c, "Audit logs retrieved successfully", gin.H{
		"data":  logs,
		"meta": gin.H{
			"page":    page,
			"perPage": perPage,
			"total":   total,
		},
	})
}
