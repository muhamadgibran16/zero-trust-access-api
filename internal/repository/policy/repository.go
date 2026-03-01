package policy

import (
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PolicyRepository struct {
	db *gorm.DB
}

func NewPolicyRepository(db *gorm.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

func (r *PolicyRepository) Create(policy *model.PolicyRule) error {
	return r.db.Create(policy).Error
}

func (r *PolicyRepository) FindAll() ([]model.PolicyRule, error) {
	var policies []model.PolicyRule
	err := r.db.Find(&policies).Error
	return policies, err
}

func (r *PolicyRepository) FindByID(id uuid.UUID) (*model.PolicyRule, error) {
	var policy model.PolicyRule
	err := r.db.First(&policy, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *PolicyRepository) Update(policy *model.PolicyRule) error {
	return r.db.Save(policy).Error
}

func (r *PolicyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.PolicyRule{}, "id = ?", id).Error
}
