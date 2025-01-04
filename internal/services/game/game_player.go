package game

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Player represents a connected player.
type Player struct {
	ID      uint            `json:"id"`
	Conn    *websocket.Conn `json:"-"`
	Room    *Room           `json:"-"`
	IsHost  bool            `json:"isHost"`
	IsReady bool            `json:"isReady"`
	send    chan []byte     `json:"-"`
	Code    string          `json:"code"`
}

// readPump handles messages from the client.
func (p *Player) readPump(manager *GameManager) {
	defer func() {
		manager.Unregister <- p
		p.Conn.Close()
	}()

	for {
		_, message, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		p.handleMessage(message, manager)
	}
}

// writePump sends messages to the client.
func (p *Player) writePump() {
	ticker := time.NewTicker(time.Second) // Keep-alive ticker
	defer func() {
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-p.send:
			if !ok {
				// Channel closed, exit loop
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := p.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return // Error writing, exit loop
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return // Error closing writer, exit loop
			}
		}
	}
}

// handleMessage handles a message received from the client.
func (p *Player) handleMessage(message []byte, manager *GameManager) {
	var msg Message
	err := json.Unmarshal(message, &msg)
	if err != nil {
		log.Printf("error unmarshalling message: %v", err)
		return
	}

	switch msg.Type {
	case "createRoom":
		manager.CreateRoom(p)
	case "joinRoom":
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			log.Println("Invalid payload for joinRoom")
			return
		}
		roomID, ok := payload["roomId"].(string)
		if !ok {
			log.Println("Invalid payload for joinRoom: missing or invalid 'roomId'")
			return
		}
		manager.JoinRoom(p, roomID)
	case "playerReady":
		p.IsReady = true
		// Check if all players in the room are ready and start the game
		if p.Room != nil {
			allReady := true
			for _, player := range p.Room.Players {
				if !player.IsReady {
					allReady = false
					break
				}
			}
			if allReady {
				p.Room.StartGame()
			}
		}
	case "start":
		if p.Room != nil && p.IsHost {
			p.Room.StartGame()
		}

	}
}
