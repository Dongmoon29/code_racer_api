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
	for {
		select {
		case player := <-gm.Register:
			gm.handlePlayerJoin(player)
		case player := <-gm.Unregister:
			gm.handlePlayerLeave(player)
		}
	}
}

// handlePlayerJoin handles a new player joining the game.
func (gm *GameManager) handlePlayerJoin(player *Player) {
	// Add the player to the register channel
	player.send = make(chan []byte, 256)

	// Send the current rooms list to the newly connected player
	roomsList := gm.getRoomsList()
	msg, _ := json.Marshal(Message{Type: "room_list", Payload: roomsList})
	player.send <- msg

	// Start the read and write pumps for the player
	go player.writePump()
	go player.readPump(gm)
}

// handlePlayerLeave handles a player leaving the game.
func (gm *GameManager) handlePlayerLeave(player *Player) {
	if player.Room != nil {
		player.Room.Mutex.Lock()
		delete(player.Room.Players, player.ID)

		// If the player is a host, update the room status and assign a new host
		if player.IsHost && len(player.Room.Players) > 0 {
			for _, p := range player.Room.Players {
				p.IsHost = true
				break
			}
		}

		// If the room is empty, delete it
		if len(player.Room.Players) == 0 {
			player.Room.Status = "closed"
			delete(gm.Rooms, player.Room.ID)
		} else {
			// Notify other players in the room about the player leaving
			player.Room.Broadcast <- []byte(fmt.Sprintf("Player %d left", player.ID))
		}
		player.Room.Mutex.Unlock()
		// Broadcast the updated rooms list to all players
		gm.broadcastRoomsList()
	}
}

// CreateRoom creates a new game room.
func (gm *GameManager) CreateRoom(player *Player) *Room {
	gm.Mutex.Lock()
	defer gm.Mutex.Unlock()

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
	gm.Rooms[roomID] = room
	go room.run()

	// Notify the player about the room creation
	player.send <- createRoomMessage(room, player)

	// Broadcast the updated rooms list to all players
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
	msg, _ := json.Marshal(Message{Type: "roomsList", Payload: roomsList})

	for _, room := range gm.Rooms {
		for _, player := range room.Players {
			player.send <- msg
		}
	}
}
