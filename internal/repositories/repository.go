package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Repository struct {
	UserRepository UserRepositoryInterface
	RoleRepository RoleRepositoryInterface
}

type UserRepositoryInterface interface {
	GetByID(context.Context, int) (*models.User, error)
	GetByEmail(context.Context, string) (*models.User, error)
	Create(context.Context, *models.User) (*models.User, error)
	Activate(context.Context, string) error
	Delete(context.Context, int64) error
}

type RoleRepositoryInterface interface {
	GetByName(context.Context, string) (*models.Role, error)
}

func NewRepository(db *gorm.DB) Repository {
	return Repository{
		UserRepository: &UserRepositoryImpl{db},
		RoleRepository: &RoleRepositoryImpl{db},
	}
}

func withTx(db *gorm.DB, ctx context.Context, fn func(tx *gorm.DB) error) error {
	// GORM 트랜잭션 시작
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 비즈니스 로직 실행
		if err := fn(tx); err != nil {
			return err // 에러 발생 시 자동 롤백
		}
		// 성공 시 자동 커밋
		return nil
	})
}
