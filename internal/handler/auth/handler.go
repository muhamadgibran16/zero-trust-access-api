package auth

import (
	"fmt"

	service "github.com/gibran/go-gin-boilerplate/internal/service/auth"
	"github.com/gibran/go-gin-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler handles authentication requests
type Handler struct {
	service *service.AuthService
}

// NewHandler creates a new Auth Handler
func NewHandler(s *service.AuthService) *Handler {
	return &Handler{service: s}
}

// Register handles POST /auth/register
// @Summary Register a new user
// @Description Create a new user account with name, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Registration details"
// @Success 201 {object} response.Response{data=model.User}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, "User registered successfully", user)
}

// Login handles POST /auth/login
// @Summary Login user
// @Description Authenticate user and return access & refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=service.AuthResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := h.service.Login(req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, "Login MFA Required", res)
}

// VerifyMFA handles POST /auth/verify-mfa
// @Summary Verify MFA Code
// @Description Validate MFA code and issue tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.MfaVerifyRequest true "MFA Code"
// @Success 200 {object} response.Response{data=service.AuthResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/verify-mfa [post]
func (h *Handler) VerifyMFA(c *gin.Context) {
	var req service.MfaVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := h.service.VerifyMFA(req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, "MFA Verification successful", res)
}

// SsoLogin handles POST /auth/sso-login
// @Summary SSO Login
// @Description Authenticate via external Provider (Google Workspace, Azure AD, Okta)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.SsoLoginRequest true "SSO Token details"
// @Success 200 {object} response.Response{data=service.AuthResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/sso-login [post]
func (h *Handler) SsoLogin(c *gin.Context) {
	var req service.SsoLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := h.service.SsoLogin(req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, "SSO Login successful", res)
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Refresh handles POST /auth/refresh
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	accessToken, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, "Token refreshed successfully", gin.H{
		"accessToken": accessToken,
	})
}

// Logout handles POST /auth/logout
// @Summary Logout user
// @Description Log out the current user (placeholder for stateless JWT)
// @Tags auth
// @Produce json
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// In a stateless JWT setup, logout is usually handled by the client 
	// (deleting the token). For more security, one could blacklist tokens.
	response.Success(c, "Logged out successfully", nil)
}

// SetupMFA handles POST /auth/setup-mfa
// @Summary Setup MFA
// @Description Initiate MFA setup and return secret URI
// @Tags auth
// @Produce json
// @Success 200 {object} response.Response{data=service.MfaSetupResponse}
// @Security BearerAuth
// @Router /auth/setup-mfa [post]
func (h *Handler) SetupMFA(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	res, err := h.service.SetupMFA(fmt.Sprintf("%v", userID))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, "MFA setup initialized", res)
}

type EnableMfaRequest struct {
	Code string `json:"code" binding:"required"`
}

// EnableMFA handles POST /auth/enable-mfa
// @Summary Enable MFA
// @Description Confirm the first TOTP code to fully enable MFA
// @Tags auth
// @Produce json
// @Param request body EnableMfaRequest true "MFA Code"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /auth/enable-mfa [post]
func (h *Handler) EnableMFA(c *gin.Context) {
	var req EnableMfaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID, _ := c.Get("userID")
	err := h.service.EnableMFA(fmt.Sprintf("%v", userID), req.Code)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	
	response.Success(c, "MFA successfully enabled", nil)
}
