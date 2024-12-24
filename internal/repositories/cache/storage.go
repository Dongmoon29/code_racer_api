package cache

import (
	"context"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"github.com/go-redis/redis/v8"
)

type Users interface {
	Get(context.Context, int64) (*models.User, error)
	Set(context.Context, *models.User) error
	Delete(context.Context, int64)
}

type Storage struct {
	Users Users
}

func NewRedisStorage(rbd *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rbd},
	}
}
