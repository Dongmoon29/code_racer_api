package game

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/Dongmoon29/code_racer_api/internal/services/game"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GameController struct {
	GameService game.GameService
	logger      *zap.SugaredLogger
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
		}
	})
	return instance
}

func (gc *GameController) HandleCreateGameRoom(c *gin.Context) {
	var createGameRoomDto dtos.CreateGameRoomDto
	gc.logger.Debugln("inside of HandleCreateGameRoom")
	user, exist := c.Get("user")
	if !exist {
		fmt.Println("Create game roome failed, user not exist")
	}

	if err := c.ShouldBindJSON(&createGameRoomDto); err != nil {
		c.JSON(http.StatusBadRequest, dtos.CreateGameRoomResponseDto{
			Message: "dto error",
			Ok:      false,
		})
		return
	}
	convertedUser, ok := user.(*mapper.MappedUser)
	if !ok {
		c.JSON(http.StatusInternalServerError, dtos.CreateGameRoomResponseDto{
			Message: "failed to create room convert error",
			Ok:      false,
		})
	}
	roomID, err := gc.GameService.CreateGameRoom(createGameRoomDto.RoomName, convertedUser)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.CreateGameRoomResponseDto{
			Message: "failed to create room",
			Ok:      false,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.CreateGameRoomResponseDto{
		Message: "created room successfully",
		Ok:      true,
		RoomID:  *roomID,
	})
}

func (gc *GameController) HandleJoinGameRoom(c *gin.Context) {
	gc.logger.Debugln("inside of HandleJoinGameRoom()")
	roomID := c.Param("id")

	gc.GameService.JoinGameRoom(c, roomID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, dtos.CreateGameRoomResponseDto{
	// 		Message: "failed to join room",
	// 		Ok:      false,
	// 	})
	// 	return
	// }

	// c.JSON(http.StatusOK, dtos.JoinGameRoomResponseDto{
	// 	Message: "join room successfully",
	// 	Ok:      true,
	// })
}

func (gc *GameController) HandleGetGameRooms(c *gin.Context) {
	keys, err := gc.GameService.GetAllGameRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get game rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": keys})
}
