package repositories

import (
	"context"
	"errors"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	DB *gorm.DB
}

func (s *UserRepositoryImpl) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := s.DB.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, err
}

func (s *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, err
}

func (s *UserRepositoryImpl) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if err := s.DB.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("duplicate entry for user")
		}
		return nil, err
	}
	return user, nil
}

func (s *UserRepositoryImpl) Activate(ctx context.Context, email string) error {
	err := s.DB.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", email).
		Update("is_active", true).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *UserRepositoryImpl) Delete(ctx context.Context, id int64) error {
	err := s.DB.WithContext(ctx).Delete(&models.User{}, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user not found")
	}
	return err
}
