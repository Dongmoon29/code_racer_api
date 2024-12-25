package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/go-redis/redis/v8"
)

type UserRedisImpl struct {
	rdb *redis.Client
}

const UserExpTime = time.Hour

func (s *UserRedisImpl) Get(ctx context.Context, userID int) (*mapper.MappedUser, error) {
	cacheKey := fmt.Sprintf("user-%d", userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user mapper.MappedUser
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserRedisImpl) Set(ctx context.Context, user *mapper.MappedUser) error {
	cacheKey := fmt.Sprintf("user-%d", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEX(ctx, cacheKey, json, UserExpTime).Err()
}

func (s *UserRedisImpl) Delete(ctx context.Context, userID int) error {
	// Redis 키 생성
	cacheKey := fmt.Sprintf("user-%d", userID)

	// Redis 키 삭제
	err := s.rdb.Del(ctx, cacheKey).Err()
	if err != nil {
		// 에러 처리
		return fmt.Errorf("failed to delete cache for user %d: %w", userID, err)
	}

	// 성공 시 nil 반환
	return nil
}
