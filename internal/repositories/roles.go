package repositories

import (
	"context"
	"errors"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"gorm.io/gorm"
)

type RoleRepositoryImpl struct {
	DB *gorm.DB
}

func (s *RoleRepositoryImpl) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := s.DB.WithContext(ctx).Where("name=?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &role, err
}
