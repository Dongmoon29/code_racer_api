package game

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/cache"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	instance GameService
	once     sync.Once
)

var gameRooms sync.Map

type GameService struct {
	gameStore cache.GameRedisStoreInterface
	logger    *zap.SugaredLogger
}

func NewGameService(gameStore cache.GameRedisStoreInterface, logger *zap.SugaredLogger) GameService {
	once.Do(func() {
		instance = GameService{
			gameStore: gameStore,
			logger:    logger,
		}
	})
	return instance
}

var mu sync.Mutex

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketClient struct {
	conn *websocket.Conn
	send chan []byte
}

type GameRoom struct {
	id         string
	clients    map[*WebsocketClient]bool
	broadcast  chan []byte
	register   chan *WebsocketClient
	unregister chan *WebsocketClient
}

func newGameRoom() *GameRoom {
	id := uuid.NewString()
	return &GameRoom{
		id:         id,
		clients:    make(map[*WebsocketClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *WebsocketClient),
		unregister: make(chan *WebsocketClient),
	}
}

// TODO
func (room *GameRoom) run() {
	for {
		select {
		case client := <-room.register:
			if len(room.clients) >= 2 {
				client.conn.WriteMessage(websocket.TextMessage, []byte("Room is full"))
				client.conn.Close()
			} else {
				room.clients[client] = true
			}
		case client := <-room.unregister:
			if _, ok := room.clients[client]; ok {
				delete(room.clients, client)
				close(client.send)
			}
		case message := <-room.broadcast:
			for client := range room.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(room.clients, client)
				}
			}
		}
	}
}

func (gs *GameService) GetAllGameRooms() ([]models.GameState, error) {
	ctx := context.Background()
	gameRooms, err := gs.gameStore.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return gameRooms, nil
}

func (gs *GameService) CreateGameRoom(roomName string, user *models.User) (*string, error) {
	gameID := uuid.NewString()

	gameState := &models.GameState{
		ID:        gameID,
		RoomName:  roomName,
		Status:    "waiting",
		CreatedAt: time.Now(),
		CreatedBy: user.Username,
		UserID:    user.ID,
	}

	ctx := context.Background()
	err := gs.gameStore.Set(ctx, gameState)
	if err != nil {
		log.Printf("Failed to save game state: %v", err)
		return nil, fmt.Errorf("failed to save game state: %v", err)
	}

	room := newGameRoom()

	gameRooms.Store(gameID, room)

	go room.run()

	return &gameID, nil
}

func (gc *GameService) JoinGameRoom(c *gin.Context, roomID string) error {
	ctx := context.Background()

	roomName, err := gc.gameStore.Get(ctx, roomID)
	if err != nil || roomName == nil {
		return fmt.Errorf("cannot find a room with id in redis: %s", roomID)
	}

	// room, ok := gameRooms.Load(fmt.Sprintf("game-%s", roomID))
	room, ok := gameRooms.Load(roomID)
	if !ok {
		return fmt.Errorf("cannot find a room with id in in memory: %s", roomID)
	}

	gameRoom, ok := room.(*GameRoom)
	if !ok {
		return fmt.Errorf("invalid room type")
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade WebSocket with roomID: %s", roomID)
	}

	client := &WebsocketClient{
		conn: conn,
		send: make(chan []byte),
	}
	gameRoom.register <- client

	go client.readPump(gameRoom)
	go client.writePump()

	return nil
}

func (c *WebsocketClient) readPump(room *GameRoom) {
	defer func() {
		log.Println("close connection")
		room.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Printf("error happen %v\n", err.Error())
			break
		}
		room.broadcast <- message
	}
}

func (c *WebsocketClient) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if !ok {
			return
		}
		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}

// func cleanupRooms() {
// 	ticker := time.NewTicker(10 * time.Minute)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		mu.Lock()
// 		for id, room := range gameRooms {
// 			if room.isExpired() {
// 				delete(gameRooms, id)
// 			}
// 		}
// 		mu.Unlock()
// 	}
// }
