package service

import (
	"context"
	"errors"
	"time"

	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
	repository "github.com/gibran/go-gin-boilerplate/internal/repository/user"
	"github.com/google/uuid"
)

// UserService handles user management logic
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type UpdateUserRequest struct {
	Name string `json:"name" binding:"omitempty,min=3"`
	Role string `json:"role" binding:"omitempty,oneof=admin user"`
}

// GetAllUsers retrieves users with pagination
func (s *UserService) GetAllUsers(page, perPage int) ([]model.User, int64, error) {
	return s.repo.FindAll(page, perPage)
}

// GetUserByID retrieves a single user
func (s *UserService) GetUserByID(id uuid.UUID) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// UpdateUser modifies user details
func (s *UserService) UpdateUser(id uuid.UUID, req UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	err = s.repo.Update(user)
	return user, err
}

// DeleteUser removes a user
func (s *UserService) DeleteUser(id uuid.UUID) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	return s.repo.Delete(id)
}

// RevokeSessions explicitly forces a user to log out by blacklisting them in Redis
func (s *UserService) RevokeSessions(id uuid.UUID) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	
	if database.RedisClient != nil {
		// Set a revocation flag spanning max JWT lifetime (e.g. 7 days)
		return database.RedisClient.Set(context.Background(), "revoke_user:"+id.String(), "true", 7*24*time.Hour).Err()
	}
	
	return errors.New("redis is not configured for session revocation")
}
