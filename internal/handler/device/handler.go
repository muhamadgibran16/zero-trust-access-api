package device

import (
	service "github.com/gibran/go-gin-boilerplate/internal/service/device"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles MDM device endpoints
type Handler struct {
	service *service.DeviceService
}

// NewHandler creates a new Device Handler
func NewHandler(s *service.DeviceService) *Handler {
	return &Handler{service: s}
}

// RegisterDevice handles POST /users/devices
func (h *Handler) RegisterDevice(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "Missing user context")
		return
	}

	userID, err := uuid.Parse(userIDStr.(uuid.UUID).String())
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req service.RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	device, err := h.service.RegisterDevice(userID, req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, "Device registration submitted for IT approval", device)
}

// GetMyDevices handles GET /users/devices
func (h *Handler) GetMyDevices(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, _ := uuid.Parse(userIDStr.(uuid.UUID).String())

	devices, err := h.service.GetUserDevices(userID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, "Get devices successfully", devices)
}

// GetAllDevices handles GET /users/admin/devices
func (h *Handler) GetAllDevices(c *gin.Context) {
	devices, err := h.service.GetAllDevices()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, "Get all devices successfully", devices)
}

// ApproveDevice handles PUT /users/admin/devices/:mac/approve
func (h *Handler) ApproveDevice(c *gin.Context) {
	mac := c.Param("mac")
	err := h.service.ApproveDevice(mac)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Device approved successfully", nil)
}

// RejectDevice handles PUT /users/admin/devices/:mac/reject
func (h *Handler) RejectDevice(c *gin.Context) {
	mac := c.Param("mac")
	err := h.service.RejectDevice(mac)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Device rejected successfully", nil)
}
