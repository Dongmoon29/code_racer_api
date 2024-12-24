package auth

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/repositories"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	utils "github.com/Dongmoon29/code_racer_api/internal/utils/auth"
	"gorm.io/gorm"
)

var (
	instance AuthService
	once     sync.Once
)

type AuthService struct {
	userRepository repositories.UserRepository
	roleRepository repositories.RoleRepository
}

func NewAuthService(ur repositories.UserRepository, rr repositories.RoleRepository) AuthService {
	once.Do(func() {
		instance = AuthService{
			userRepository: ur,
			roleRepository: rr,
		}
	})
	return instance
}

type Claims struct {
	UserID         string
	ExpirationTime time.Time
}

func (us *AuthService) FindAndVerifyUserByEmail(dto dtos.SigninRequestDto) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	user, err := us.userRepository.GetByEmail(ctx, dto.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	if !utils.CheckPasswordHash(dto.Password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (us *AuthService) CreateUser(dto dtos.SignupRequestDto) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hashedPassword, err := utils.HashPassword(dto.Password)

	if err != nil {
		return nil, err
	}

	role, err := us.roleRepository.GetByName(ctx, "user")
	if err != nil {
		log.Println("cannot find role with name user")
		return nil, err
	}

	user := &models.User{
		Username: dto.Username,
		Password: hashedPassword,
		Email:    dto.Email,
		IsActive: true,
		RoleID:   role.ID,
		Role:     role,
	}

	createdUser, err := us.userRepository.Create(ctx, user)
	if err != nil {
		log.Println("failed create user")
		return nil, err
	}

	return createdUser, nil
}
