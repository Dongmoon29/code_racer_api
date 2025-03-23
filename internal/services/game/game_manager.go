package game

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// GameManager manages game rooms and game logic.
type GameManager struct {
	Rooms      map[string]*Room `json:"rooms"`
	Mutex      sync.Mutex       `json:"-"`
	Register   chan *Player     `json:"-"`
	Unregister chan *Player     `json:"-"`
	IsRunning  bool             `json:"-"`
}

// NewGameManager creates a new GameManager.
func NewGameManager() *GameManager {
	return &GameManager{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
	}
}

// Run starts the GameManager loop.
func (gm *GameManager) Run() {
	gm.IsRunning = true
	defer func() {
		gm.IsRunning = false
	}()
	fmt.Println("Game Manager started")
	for {
		select {
		case player := <-gm.Register:
			fmt.Println("registered player")
			gm.handlePlayerJoin(player)
		case player := <-gm.Unregister:
			gm.handlePlayerLeave(player)
		}
	}
}

// handlePlayerJoin handles a new player joining the game.
func (gm *GameManager) handlePlayerJoin(player *Player) {
	player.send = make(chan []byte, 256)

	msg, _ := json.Marshal(Message{Type: "init", Payload: map[string]string{"message": "hello"}})
	player.send <- msg

	go player.writePump()
	go player.readPump(gm)
}

func (gm *GameManager) handlePlayerLeave(player *Player) {
	if player.Room != nil {
		// 룸 내부 변경은 룸의 락으로 보호
		player.Room.Mutex.Lock()
		delete(player.Room.Players, player.ID)

		if player.IsHost && len(player.Room.Players) > 0 {
			for _, p := range player.Room.Players {
				p.IsHost = true
				break
			}
		}

		if len(player.Room.Players) == 0 {
			player.Room.Status = "closed"
			// gm.Rooms 수정 전 gm.Mutex를 잡음
			gm.Mutex.Lock()
			delete(gm.Rooms, player.Room.ID)
			gm.Mutex.Unlock()
		} else {
			player.Room.Broadcast <- []byte(fmt.Sprintf("Player %d left", player.ID))
		}
		player.Room.Mutex.Unlock()

		gm.broadcastRoomsList()
	}
}

// CreateRoom creates a new game room.
func (gm *GameManager) CreateRoom(player *Player) *Room {
	// 룸 생성에 필요한 데이터 준비
	roomID := uuid.NewString()
	room := &Room{
		ID:        roomID,
		Players:   make(map[uint]*Player),
		Status:    "waiting",
		Game:      &Game{},
		Broadcast: make(chan []byte),
	}
	room.Players[player.ID] = player
	player.Room = room
	player.IsHost = true
	player.IsReady = false

	// gm.Rooms에 추가 및 룸 런 실행은 gm.Mutex로 보호
	gm.Mutex.Lock()
	gm.Rooms[roomID] = room
	gm.Mutex.Unlock()

	go room.run()

	// 채널 송신은 락 밖에서 진행
	player.send <- createRoomMessage(room, player)

	// 방 목록 브로드캐스트 역시 락 밖에서 처리
	gm.broadcastRoomsList()

	return room
}

// JoinRoom allows a player to join a specific room.
func (gm *GameManager) JoinRoom(player *Player, roomID string) {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()

	room, ok := gm.Rooms[roomID]
	if !ok {
		// Handle room not found
		player.send <- createErrorMessage("Room not found")
		return
	}

	if room.Status != "waiting" {
		player.send <- createErrorMessage("Room is not available")
		return
	}

	room.Players[player.ID] = player
	player.Room = room
	player.IsHost = false
	player.IsReady = false

	// Notify player about joining the room
	player.send <- createRoomMessage(room, player)

	// Notify other players about the new player
	room.Broadcast <- []byte(fmt.Sprintf("Player %d joined", player.ID))

	// Broadcast the updated rooms list to all players
	gm.broadcastRoomsList()
}

func createErrorMessage(message string) []byte {
	msg := Message{
		Type:    "error",
		Payload: message,
	}
	msgBytes, _ := json.Marshal(msg)
	return msgBytes
}

// getRoomsList returns a list of all available rooms.
func (gm *GameManager) getRoomsList() []map[string]interface{} {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()

	roomsList := make([]map[string]interface{}, 0)
	for _, room := range gm.Rooms {
		if room.Status == "waiting" {
			roomData := map[string]interface{}{
				"id":      room.ID,
				"players": len(room.Players),
				"status":  room.Status,
			}
			roomsList = append(roomsList, roomData)
		}
	}
	return roomsList
}

// broadcastRoomsList sends the updated rooms list to all connected players.
func (gm *GameManager) broadcastRoomsList() {
	roomsList := gm.getRoomsList()
	msg, _ := json.Marshal(Message{Type: "rooms_list", Payload: roomsList})

	for _, room := range gm.Rooms {
		for _, player := range room.Players {
			player.send <- msg
		}
	}
}
