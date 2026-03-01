package policy

import (
	"errors"

	"github.com/gibran/go-gin-boilerplate/internal/model"
	repository "github.com/gibran/go-gin-boilerplate/internal/repository/policy"
	"github.com/google/uuid"
)

type PolicyService struct {
	repo *repository.PolicyRepository
}

func NewPolicyService(repo *repository.PolicyRepository) *PolicyService {
	return &PolicyService{repo: repo}
}

type CreatePolicyRequest struct {
	Type     string `json:"type" binding:"required,oneof=DENY_IP ALLOW_IP REQUIRE_ROLE"`
	Value    string `json:"value" binding:"required"`
	Resource string `json:"resource"`
	IsActive *bool  `json:"isActive"`
}

type UpdatePolicyRequest struct {
	Type     string `json:"type" binding:"omitempty,oneof=DENY_IP ALLOW_IP REQUIRE_ROLE"`
	Value    string `json:"value"`
	Resource string `json:"resource"`
	IsActive *bool  `json:"isActive"`
}

func (s *PolicyService) CreatePolicy(req CreatePolicyRequest) (*model.PolicyRule, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	policy := &model.PolicyRule{
		Type:     req.Type,
		Value:    req.Value,
		Resource: req.Resource,
		IsActive: isActive,
	}
	err := s.repo.Create(policy)
	return policy, err
}

func (s *PolicyService) GetAllPolicies() ([]model.PolicyRule, error) {
	return s.repo.FindAll()
}

func (s *PolicyService) UpdatePolicy(id uuid.UUID, req UpdatePolicyRequest) (*model.PolicyRule, error) {
	policy, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("policy not found")
	}

	if req.Type != "" {
		policy.Type = req.Type
	}
	if req.Value != "" {
		policy.Value = req.Value
	}
	if req.Resource != "" {
		policy.Resource = req.Resource
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}

	err = s.repo.Update(policy)
	return policy, err
}

func (s *PolicyService) DeletePolicy(id uuid.UUID) error {
	return s.repo.Delete(id)
}
