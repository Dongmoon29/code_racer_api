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
	// 1. Redis에서 모든 게임 키 검색
	keys, err := s.rdb.Keys(ctx, "game-*").Result()
	if err != nil {
		return nil, err
	}

	// 2. 결과 저장할 슬라이스 초기화
	var games []models.GameState

	// 3. 각 키에 대해 데이터 조회 및 파싱
	for _, key := range keys {
		data, err := s.rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			continue // 데이터 없으면 건너뜀
		} else if err != nil {
			return nil, err // 기타 에러 처리
		}

		// JSON 역직렬화
		var game models.GameState
		if err := json.Unmarshal([]byte(data), &game); err != nil {
			return nil, err
		}

		// 게임 목록에 추가
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
