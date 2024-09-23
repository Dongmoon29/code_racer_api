package game

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/util/client"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GameService struct {
}

func NewGameService() GameService {
	return GameService{}
}

var mu sync.Mutex

// WebSocket 업그레이더
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 도메인에서 접근 허용
	},
}

// 클라이언트 연결 구조체
type WebsocketClient struct {
	conn *websocket.Conn
	send chan []byte
}

// 게임방 구조체
type GameRoom struct {
	clients    map[*WebsocketClient]bool
	broadcast  chan []byte
	register   chan *WebsocketClient
	unregister chan *WebsocketClient
}

var gameRooms = make(map[string]*GameRoom)

func newGameRoom() *GameRoom {
	return &GameRoom{
		clients:    make(map[*WebsocketClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *WebsocketClient),
		unregister: make(chan *WebsocketClient),
	}
}

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

func (gc *GameService) CreateGameRoom(roomName string) (*string, error) {
	rdsClient := client.RdsClient
	ctx := context.Background()
	roomID := strings.ReplaceAll(uuid.New().String(), "-", "")

	err := rdsClient.Set(ctx, roomID, roomName)

	if err != nil {
		return nil, err
	}

	room := newGameRoom()

	mu.Lock()
	if _, exists := gameRooms[roomID]; exists {
		mu.Unlock()
		return nil, fmt.Errorf("room with ID %s already exists", roomID)
	}
	gameRooms[roomID] = room
	mu.Unlock()

	go room.run()

	return &roomID, nil
}

func (gc *GameService) JoinGameRoom(c *gin.Context, roomID string) error {
	rdsClient := client.RdsClient
	ctx := context.Background()

	// Redis에서 게임방 존재 여부 확인
	roomName, err := rdsClient.Get(ctx, roomID)
	if err != nil || roomName == nil {
		return fmt.Errorf("cannot find a room with id in redis: %s", roomID)
	}

	// In-memory에 있는 게임방 확인
	room, ok := gameRooms[roomID]
	if !ok {
		return fmt.Errorf("cannot find a room with id in in memory: %s", roomID)
	}

	// WebSocket 업그레이드
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade WebSocket with roomID: %s", roomID)
	}

	// 새로운 클라이언트 생성 및 방에 등록
	client := &WebsocketClient{
		conn: conn,
		send: make(chan []byte),
	}
	room.register <- client

	// 클라이언트의 메시지 읽기 및 쓰기 루프 시작
	go client.readPump(room)
	go client.writePump()

	return nil
}

// 메세지를 받아서 뿌려주는 함수
func (c *WebsocketClient) readPump(room *GameRoom) {
	defer func() {
		log.Println("close connection")
		room.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
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
