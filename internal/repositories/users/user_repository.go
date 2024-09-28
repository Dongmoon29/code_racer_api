package users

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Dongmoon29/code_racer_api/db/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindById(userID string) (*models.User, error)
	InsertOne(user models.User) (*models.User, error)
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

var (
	once     sync.Once
	instance UserRepositoryImpl
)

func NewUserRepository(db *gorm.DB) UserRepositoryImpl {
	once.Do(func() {
		instance = UserRepositoryImpl{
			db: db,
		}
	})
	return instance
}

func (ur *UserRepositoryImpl) FindById(userID string) (*models.User, error) {
	var user models.User

	if err := ur.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	var user models.User
	// 이메일을 기반으로 사용자 검색
	if err := ur.db.Where("email = ?", email).First(&user).Error; err != nil {
		// 사용자를 찾을 수 없으면 에러 반환
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) CreateOne(user models.User) (*models.User, error) {
	result := ur.db.Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
