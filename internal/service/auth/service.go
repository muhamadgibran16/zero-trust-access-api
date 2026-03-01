package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gibran/go-gin-boilerplate/config"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	repository "github.com/gibran/go-gin-boilerplate/internal/repository/user"
	"github.com/gibran/go-gin-boilerplate/pkg/security"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthService handles authentication logic
type AuthService struct {
	repo         *repository.UserRepository
	config       *config.Config
	googleConfig *oauth2.Config
}

// NewAuthService creates a new AuthService
func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
	gConfig := &oauth2.Config{
		RedirectURL:  cfg.GoogleRedirectURL,
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	return &AuthService{repo: repo, config: cfg, googleConfig: gConfig}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type MfaVerifyRequest struct {
	Email    string `json:"email" binding:"required,email"`
	MfaCode  string `json:"mfaCode" binding:"required"`
}

type SsoLoginRequest struct {
	Provider string `json:"provider" binding:"required"` // e.g. "google", "microsoft"
	SSOToken string `json:"ssoToken" binding:"required"` // External ID Token
}

type AuthResponse struct {
	NeedsMFA     bool        `json:"needsMfa,omitempty"`
	AccessToken  string      `json:"accessToken,omitempty"`
	RefreshToken string      `json:"refreshToken,omitempty"`
	User         *model.User `json:"user,omitempty"`
}

// Register creates a new user and returns their data
func (s *AuthService) Register(req RegisterRequest) (*model.User, error) {
	// Check if user already exists
	existing, _ := s.repo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
	}

	err = s.repo.Create(user)
	return user, err
}

// Login validates credentials and requires MFA
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare password
	if !security.ComparePassword(user.Password, req.Password) {
		// Increment risk score for failed login
		user.RiskScore += 10
		_ = s.repo.Update(user)
		return nil, errors.New("invalid email or password")
	}

	// Reset risk score upon successful login (optional, but good practice to lower risk on success, or keep it static? Let's just track it up for now and let admins reset it manually or reset on success. For this demo we'll reduce it by 5 on success to prevent accidental lockouts)
	if user.RiskScore > 0 {
		user.RiskScore -= 5
		if user.RiskScore < 0 {
			user.RiskScore = 0
		}
		_ = s.repo.Update(user)
	}

	// Adaptive Auth: If MFA is not enabled, log them in immediately so they can set it up
	if !user.MFAEnabled {
		accessToken, _ := security.GenerateToken(user.ID, user.Role, s.config.JWTSecret, s.config.JWTAccessExpireHours)
		refreshToken, _ := security.GenerateToken(user.ID, user.Role, s.config.JWTSecret, s.config.JWTRefreshExpireDays*24)
		return &AuthResponse{
			NeedsMFA:     false,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
		}, nil
	}

	return &AuthResponse{
		NeedsMFA: true,
	}, nil
}

// VerifyMFA validates the MFA code and generates JWT tokens
func (s *AuthService) VerifyMFA(req MfaVerifyRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email")
	}

	// Actual MFA Validation
	if user.MFAEnabled && user.MFASecret != "" {
		valid := totp.Validate(req.MfaCode, user.MFASecret)
		if !valid {
			user.RiskScore += 20 // MFA failure is higher risk
			_ = s.repo.Update(user)
			return nil, errors.New("invalid MFA code")
		}
	} else {
		// Fallback for users who haven't set up MFA yet in phase 2 testing
		if req.MfaCode != "123456" {
			user.RiskScore += 10
			_ = s.repo.Update(user)
			return nil, errors.New("invalid MFA code")
		}
	}

	// Generate tokens
	accessToken, err := security.GenerateToken(user.ID, user.Role, s.config.JWTSecret, s.config.JWTAccessExpireHours)
	if err != nil {
		return nil, err
	}

	refreshToken, err := security.GenerateToken(user.ID, user.Role, s.config.JWTSecret, s.config.JWTRefreshExpireDays*24)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		NeedsMFA:     false,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// GetGoogleLoginURL returns the OAuth2 consent page URL
func (s *AuthService) GetGoogleLoginURL(state string) string {
	return s.googleConfig.AuthCodeURL(state)
}

// GoogleCallback processes the code and logs the user in
func (s *AuthService) GoogleCallback(code string) (*AuthResponse, error) {
	token, err := s.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API returned status: %d", response.StatusCode)
	}

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Id    string `json:"id"`
	}

	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed parsing user info: %v", err)
	}

	// Find user or create if they don't exist
	user, err := s.repo.FindByEmail(userInfo.Email)
	if err != nil {
		// Auto-register via SSO
		user = &model.User{
			Name:       userInfo.Name,
			Email:      userInfo.Email,
			Password:   "", // No password for SSO users
			Role:       "user",
			MFAEnabled: false,
		}
		if err := s.repo.Create(user); err != nil {
			return nil, err
		}
	}

	// SSO assumes identity validated by IdP, skip passwords/MFA for demonstration (or enforce MFA if required)
	accessToken, err := security.GenerateToken(user.ID, user.Role, s.config.JWTSecret, s.config.JWTAccessExpireHours)
	if err != nil {
		return nil, err
	}

	refreshToken, err := security.GenerateToken(user.ID, user.Role, s.config.JWTSecret, s.config.JWTRefreshExpireDays*24)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		NeedsMFA:     false,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshToken validates a refresh token and generates a new access token
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	claims, err := security.ValidateToken(tokenString, s.config.JWTSecret)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Generate new access token using the latest role from DB
	accessToken, err := security.GenerateToken(claims.UserID, user.Role, s.config.JWTSecret, s.config.JWTAccessExpireHours)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

type MfaSetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qrCode"` // data URI for image
}

// SetupMFA generates a new TOTP secret for the user
func (s *AuthService) SetupMFA(userIDStr string) (*MfaSetupResponse, error) {
	userId, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id format")
	}
	
	user, err := s.repo.FindByID(userId)
	if err != nil {
		return nil, errors.New("user not found")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ZeroTrustAccess",
		AccountName: user.Email,
	})
	if err != nil {
		return nil, err
	}

	// Update user with secret, but not fully enabled yet
	user.MFASecret = key.Secret()
	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	// Convert exact QR code to data URI or return URL to a generator. Let's return the plain secret for now.
	return &MfaSetupResponse{
		Secret: key.Secret(),
		QRCode: key.URL(), // the otpauth:// URL that can be generated into a QR code on frontend
	}, nil
}

// EnableMFA verifies the first code and locks it in
func (s *AuthService) EnableMFA(userIDStr, code string) error {
	userId, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user id format")
	}
	
	user, err := s.repo.FindByID(userId)
	if err != nil {
		return errors.New("user not found")
	}

	valid := totp.Validate(code, user.MFASecret)
	if !valid {
		return errors.New("invalid verification code")
	}

	user.MFAEnabled = true
	return s.repo.Update(user)
}
