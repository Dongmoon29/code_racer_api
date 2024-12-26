package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"github.com/go-redis/redis/v8"
)

const GameExpTime = time.Hour * 24

type GameRedisImpl struct {
	rdb *redis.Client
}

func (s *GameRedisImpl) Get(ctx context.Context, gameID string) (*models.GameState, error) {
	cacheKey := fmt.Sprintf("game-%s", gameID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var game models.GameState
	if data != "" {
		err := json.Unmarshal([]byte(data), &game)
		if err != nil {
			return nil, err
		}
	}

	return &game, nil
}

func (s *GameRedisImpl) GetAll(ctx context.Context) ([]models.GameState, error) {
	// TODO implement pagination
	keys, err := s.rdb.Keys(ctx, "game-*").Result()
	if err != nil {
		return nil, err
	}

	var games []models.GameState

	for _, key := range keys {
		data, err := s.rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			return nil, err
		}

		var game models.GameState
		if err := json.Unmarshal([]byte(data), &game); err != nil {
			return nil, err
		}

		games = append(games, game)
	}

	return games, nil
}

func (s *GameRedisImpl) Set(ctx context.Context, game *models.GameState) error {
	cacheKey := fmt.Sprintf("game-%s", game.ID)

	json, err := json.Marshal(game)
	if err != nil {
		return err
	}

	return s.rdb.SetEX(ctx, cacheKey, json, GameExpTime).Err()
}

func (s *GameRedisImpl) Delete(ctx context.Context, userID int) error {
	cacheKey := fmt.Sprintf("user-%d", userID)

	err := s.rdb.Del(ctx, cacheKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache for user %d: %w", userID, err)
	}

	return nil
}
