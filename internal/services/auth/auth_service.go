package auth

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/Dongmoon29/code_racer_api/internal/repositories"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/cache"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	utils "github.com/Dongmoon29/code_racer_api/internal/utils/auth"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	instance AuthService
	once     sync.Once
)

type AuthService struct {
	userRepository repositories.UserRepositoryInterface
	roleRepository repositories.RoleRepositoryInterface
	userStore      cache.UsersRedisStoreInterface
	logger         *zap.SugaredLogger
}

func NewAuthService(ur repositories.UserRepositoryInterface, rr repositories.RoleRepositoryInterface, us cache.UsersRedisStoreInterface, logger *zap.SugaredLogger) AuthService {
	once.Do(func() {
		instance = AuthService{
			userRepository: ur,
			roleRepository: rr,
			userStore:      us,
			logger:         logger,
		}
	})
	return instance
}

type Claims struct {
	UserID         string
	ExpirationTime time.Time
}

func (us *AuthService) GetUserByID(ctx context.Context, userID int) (*mapper.MappedUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := us.userRepository.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with ID: %d", userID)
		}
		return nil, err
	}

	mappedUser := mapper.UserMapper(user)

	return mappedUser, nil
}

func (us *AuthService) FindAndVerifyUserByEmail(dto dtos.SigninRequestDto) (*mapper.MappedUser, error) {
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

	mappedUser := mapper.UserMapper(user)

	return mappedUser, nil
}

func (us *AuthService) CreateUser(dto dtos.SignupRequestDto) (*mapper.MappedUser, error) {
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

	mappedUser := mapper.UserMapper(createdUser)

	return mappedUser, nil
}

func (us *AuthService) DeleteSession(ctx context.Context, userID int) {
	us.userStore.Delete(ctx, userID)
}

func (us *AuthService) SaveSession(ctx context.Context, user *mapper.MappedUser) error {
	err := us.userStore.Set(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}
