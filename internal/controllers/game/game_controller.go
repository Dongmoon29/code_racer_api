package game

import (
	"fmt"
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

// TODO
func (gc *GameController) HandleGameWebSocket(c *gin.Context) {
	gc.logger.Debug("웹소켓 연결 시도 감지됨")
	gc.logger.Debug("요청 헤더:")
	for name, values := range c.Request.Header {
		gc.logger.Debug(fmt.Sprintf("  %s: %v", name, values))
	}

	// 웹소켓 연결 업그레이드
	conn, err := gc.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		gc.logger.Error("웹소켓 업그레이드 실패", zap.Error(err))
		return // 이미 헤더가 전송되었으므로 JSON 응답을 반환하지 않음
	}
	gc.logger.Debug("웹소켓 연결 업그레이드 성공")

	// 사용자 정보 검증
	user, exists := c.Get("user")
	if !exists {
		gc.logger.Error("인증된 사용자 정보를 찾을 수 없음")
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized: Please login"))
		conn.Close()
		return
	}

	convertedUser, ok := user.(*mapper.MappedUser)
	if !ok {
		gc.logger.Error("사용자 객체 변환 실패")
		conn.WriteMessage(websocket.TextMessage, []byte("Server error: Invalid user data"))
		conn.Close()
		return
	}
	if convertedUser == nil {
		gc.logger.Error("convertedUser가 nil입니다")
		conn.WriteMessage(websocket.TextMessage, []byte("Server error: User data is nil"))
		conn.Close()
		return
	}

	// 게임 서비스로 연결 위임
	userID := convertedUser.ID
	gc.logger.Debug("게임 웹소켓 연결 처리 중", zap.Uint("userID", userID))

	err = gc.GameService.ConnectGameSocketConnect(conn, userID)
	if err != nil {
		gc.logger.Error("게임 웹소켓 연결 실패", zap.Error(err))
		conn.WriteMessage(websocket.TextMessage, []byte("Failed to connect to game service"))
		conn.Close()
		return
	}
}

func (gc *GameController) HandleGetGameManagerStatus(c *gin.Context) {
	manager := gc.GameService.DebugGameManager()
	c.JSON(http.StatusOK, gin.H{"manager": manager})
}
