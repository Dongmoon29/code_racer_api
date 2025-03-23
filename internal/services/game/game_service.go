package game

import (
	"fmt"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/cache"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	instance GameService
	once     sync.Once
)

// var gameRooms sync.Map

type GameService struct {
	gameStore   cache.GameRedisStoreInterface
	logger      *zap.SugaredLogger
	gameManager *GameManager
}

func NewGameService(gameManager *GameManager, gameStore cache.GameRedisStoreInterface, logger *zap.SugaredLogger) GameService {
	once.Do(func() {
		instance = GameService{
			gameStore:   gameStore,
			logger:      logger,
			gameManager: gameManager,
		}
	})
	return instance
}

func (gs *GameService) ConnectGameSocketConnect(conn *websocket.Conn, userID uint) error {
	if gs.gameManager == nil {
		gs.logger.Errorf("ConnectGameSocketConnect(), gameManager is not created.")
		return fmt.Errorf("gameManager is not created")
	}
	if gs.gameManager.Register == nil {
		gs.logger.Error("ConnectGameSocketConnect(), Register channel is nil.")
		return fmt.Errorf("register channel is nil")
	}
	player := &Player{
		ID:   userID,
		Conn: conn,
		send: make(chan []byte, 256),
	}
	gs.logger.Debug("before Register")

	// gameManager.Register 채널로 플레이어 등록
	gs.gameManager.Register <- player

	gs.logger.Debug("after Register")

	return nil
}

func (gs *GameService) DebugGameManager() *GameManager {
	return gs.gameManager
}

func (gs *GameService) GetGameRooms() []map[string]interface{} {
	gameRoom := gs.gameManager.getRoomsList()
	return gameRoom
}
