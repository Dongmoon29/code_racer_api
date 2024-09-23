package game

import (
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/services/game"
	"github.com/gin-gonic/gin"
)

type GameController struct {
	GameService game.GameService
}

var (
	instance *GameController
	once     sync.Once
)

func NewGameController(gameService game.GameService) *GameController {
	once.Do(func() {
		instance = &GameController{
			GameService: gameService,
		}
	})
	return instance
}

func (gc *GameController) HandleCreateGameRoom(c *gin.Context) {
	var createGameRoomDto dtos.CreateGameRoomDto
	if err := c.ShouldBindJSON(&createGameRoomDto); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.CreateGameRoomResponseDto{
			Message: "internal server error",
			Ok:      false,
		})
		return
	}
	roomID, err := gc.GameService.CreateGameRoom(createGameRoomDto.RoomName)

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
	roomID := c.Param("id")
	err := gc.GameService.JoinGameRoom(c, roomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.CreateGameRoomResponseDto{
			Message: "failed to join room",
			Ok:      false,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.JoinGameRoomResponseDto{
		Message: "join room successfully",
		Ok:      true,
	})
}

func (gc *GameController) HandleGetGameRooms(c *gin.Context) {
	// TODO: implement fetching all the game rooms from redis
}
