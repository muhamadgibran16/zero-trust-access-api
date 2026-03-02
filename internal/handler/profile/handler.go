package profile

import (
	"fmt"
	"net/http"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	repository "github.com/gibran/go-gin-boilerplate/internal/repository/user"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gibran/go-gin-boilerplate/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles profile-related requests
type Handler struct {
	repo *repository.UserRepository
}

// NewHandler creates a new Profile Handler
func NewHandler(repo *repository.UserRepository) *Handler {
	return &Handler{repo: repo}
}

// GetProfile returns the currently authenticated user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := uuid.Parse(fmt.Sprintf("%v", userID))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.repo.FindByID(id)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	// Also fetch their devices
	var devices []model.Device
	database.DB.Where("user_id = ?", id).Find(&devices)

	response.Success(c, "Profile loaded", gin.H{
		"user":    user,
		"devices": devices,
	})
}

// UpdateProfileRequest represents the update profile request body
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

// UpdateProfile updates the current user's name
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := uuid.Parse(fmt.Sprintf("%v", userID))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.repo.FindByID(id)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	user.Name = req.Name
	if err := h.repo.Update(user); err != nil {
		response.InternalServerError(c, "Failed to update profile")
		return
	}

	response.Success(c, "Profile updated successfully", user)
}

// ChangePasswordRequest represents the change password request body
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// ChangePassword changes the current user's password
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := uuid.Parse(fmt.Sprintf("%v", userID))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.repo.FindByID(id)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	// Verify old password
	if !security.ComparePassword(user.Password, req.OldPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "Current password is incorrect"})
		c.Abort()
		return
	}

	// Hash and set new password
	hashed, err := security.HashPassword(req.NewPassword)
	if err != nil {
		response.InternalServerError(c, "Failed to hash password")
		return
	}

	user.Password = hashed
	if err := h.repo.Update(user); err != nil {
		response.InternalServerError(c, "Failed to update password")
		return
	}

	response.Success(c, "Password changed successfully", nil)
}
