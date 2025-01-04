package game

import (
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/Dongmoon29/code_racer_api/internal/services/game"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type GameController struct {
	GameService game.GameService
	logger      *zap.SugaredLogger
	upgrader    websocket.Upgrader
}

var (
	instance *GameController
	once     sync.Once
)

func NewGameController(gameService game.GameService, logger *zap.SugaredLogger) *GameController {
	once.Do(func() {
		instance = &GameController{
			GameService: gameService,
			logger:      logger,
			upgrader: websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		}
	})
	return instance
}

func (gc *GameController) HandleGetGameRooms(c *gin.Context) {
	rooms := gc.GameService.GetGameRooms()
	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}


func (gc *GameController) HandleGameWebSocket(c *gin.Context) {
	conn, err := gc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		gc.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}

	user, ok := c.Get("user")
	if !ok {
		gc.logger.Errorf("inside of HandleGameWebSocket(), cannot find user.")
	}

	convertedUser, ok := user.(*mapper.MappedUser)
	if !ok {
		gc.logger.Errorf("inside of HandleGameWebSocket(), cannot convert user.")
	}

	userID := convertedUser.ID

	err = gc.GameService.ConnectGameSocketConnect(conn, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "inside of HandleGameWebSocket(), failed to connect websocket"})
	}
}
