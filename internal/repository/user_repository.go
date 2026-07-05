package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/user/network-monitoring/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uuid.UUID, orgID uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("id = ? AND organization_id = ?", id, orgID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListByOrg(orgID uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := r.db.Preload("Role").Where("organization_id = ?", orgID).Find(&users).Error
	return users, err
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uuid.UUID, orgID uuid.UUID) error {
	return r.db.Where("id = ? AND organization_id = ?", id, orgID).Delete(&model.User{}).Error
}
