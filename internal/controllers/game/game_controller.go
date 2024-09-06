package game

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/services/game"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
			room.clients[client] = true
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

func (gc *GameController) HandleWebSocket(c *gin.Context) {
	roomID := c.Param("id")
	room, ok := gameRooms[roomID]
	if !ok {
		room = newGameRoom()
		gameRooms[roomID] = room
		go room.run()
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 업그레이드 실패:", err)
		return
	}

	client := &WebsocketClient{conn: conn, send: make(chan []byte)}
	room.register <- client

	go client.readPump(room)
	go client.writePump()
}

func (c *WebsocketClient) readPump(room *GameRoom) {
	defer func() {
		room.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		fmt.Printf("incomming message: %s\n", message)
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
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
