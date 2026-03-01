package policy

import (
	service "github.com/gibran/go-gin-boilerplate/internal/service/policy"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.PolicyService
}

func NewHandler(s *service.PolicyService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetMany(c *gin.Context) {
	policies, err := h.service.GetAllPolicies()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Get policies successfully", policies)
}

func (h *Handler) Create(c *gin.Context) {
	var req service.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	policy, err := h.service.CreatePolicy(req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Created(c, "Policy created successfully", policy)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid policy ID format")
		return
	}

	var req service.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	policy, err := h.service.UpdatePolicy(id, req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Policy updated successfully", policy)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid policy ID format")
		return
	}

	err = h.service.DeletePolicy(id)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Policy deleted successfully", nil)
}
