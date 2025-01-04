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

func NewGameService(gameStore cache.GameRedisStoreInterface, logger *zap.SugaredLogger) GameService {
	once.Do(func() {
		gameManager := NewGameManager()
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
		gs.logger.Errorf("inside of ConnectGameSocketConnect(), gameManager is not created.")
		return fmt.Errorf("insdie of COnnectGmaeSocketConnect()")
	}
	player := &Player{
		ID:   userID,
		Conn: conn,
		send: make(chan []byte, 256),
	}

	gs.gameManager.Register <- player
	return nil
}

func (gs *GameService) GetGameRooms() []map[string]interface{} {
	gameRoom := gs.gameManager.getRoomsList()
	return gameRoom
}
