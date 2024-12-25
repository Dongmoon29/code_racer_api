package cache

import (
	"context"

	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"github.com/go-redis/redis/v8"
)

type UsersRedisStoreInterface interface {
	Get(context.Context, int) (*mapper.MappedUser, error)
	Set(context.Context, *mapper.MappedUser) error
	Delete(context.Context, int) error
}

type GameRedisStoreInterface interface {
	Get(context.Context, string) (*models.GameState, error)
	GetAll(context.Context) ([]models.GameState, error)
	Set(context.Context, *models.GameState) error
	Delete(context.Context, int) error
}

type RedisStorage struct {
	Users UsersRedisStoreInterface
	Games GameRedisStoreInterface
}

func NewRedisStorage(rbd *redis.Client) RedisStorage {
	return RedisStorage{
		Users: &UserRedisImpl{rdb: rbd},
		Games: &GameRedisImpl{rdb: rbd},
	}
}
