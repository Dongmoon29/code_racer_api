package game

import (
	"encoding/json"
	"log"
	"sync"
)

type Room struct {
	ID        string           `json:"id"`
	Players   map[uint]*Player `json:"players"`
	Status    string           `json:"status"` // "waiting", "playing", "closed"
	Game      *Game            `json:"game"`
	Mutex     sync.Mutex       `json:"-"`
	Broadcast chan []byte      `json:"-"`
}

func (room *Room) run() {
	defer func() {
		room.Status = "closed"
	}()

	for {
		select {
		case msg, ok := <-room.Broadcast:
			if !ok {
				return // Channel closed, exit loop
			}

			// 메시지 타입에 따라 적절한 핸들러 호출
			var parsedMsg Message
			if err := json.Unmarshal(msg, &parsedMsg); err != nil {
				log.Println("Error unmarshalling message:", err)
				continue
			}
			switch parsedMsg.Type {
			case "codeUpdate":
				room.handleCodeUpdate(msg)
			default:
				for _, player := range room.Players {
					select {
					case player.send <- msg: //player.send <- msg:
					default:
						close(player.send)
						delete(room.Players, player.ID)
					}
				}
			}
		}
	}
}

func (room *Room) handleCodeUpdate(msg []byte) {
	// 메시지 파싱 및 유효성 검사
	var codeUpdateMsg Message
	err := json.Unmarshal(msg, &codeUpdateMsg)
	if err != nil {
		log.Println("Error unmarshalling codeUpdate message:", err)
		return
	}

	payload, ok := codeUpdateMsg.Payload.(map[string]interface{})
	if !ok {
		log.Println("Invalid payload for codeUpdate")
		return
	}

	userID, ok := payload["userID"].(uint)
	if !ok {
		log.Println("Invalid payload for codeUpdate: missing or invalid 'userID'")
		return
	}

	code, ok := payload["code"].(string)
	if !ok {
		log.Println("Invalid payload for codeUpdate: missing or invalid 'code'")
		return
	}

	// 코드 업데이트
	player, ok := room.Players[userID]
	if ok {
		player.Code = code
	}

	// 브로드캐스트
	room.broadcastCodeUpdate(userID, code)
}

func (room *Room) broadcastCodeUpdate(userID uint, code string) {
	updateMsg := Message{
		Type: "codeUpdate",
		Payload: map[string]interface{}{
			"userID": userID,
			"code":   code,
		},
	}

	updateMsgBytes, _ := json.Marshal(updateMsg)

	for _, player := range room.Players {
		if player.ID == userID {
			player.Code = code
		}
		select {
		case player.send <- updateMsgBytes:
		default:
			close(player.send)
			delete(room.Players, player.ID)
		}
	}
}

// StartGame starts the game in the room.
func (room *Room) getHost() *Player {
	for _, player := range room.Players {
		if player.IsHost {
			return player
		}
	}
	return nil
}

func (room *Room) StartGame() {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	// Check if all players are ready
	allReady := true
	for _, player := range room.Players {
		if !player.IsReady {
			allReady = false
			break
		}
	}

	if !allReady {
		room.Broadcast <- createErrorMessage("Not all players are ready")
		return
	}

	room.Status = "playing"
	// Notify all players that the game has started
	msg := Message{Type: "gameStart"}
	msgBytes, _ := json.Marshal(msg)
	for _, player := range room.Players {
		player.send <- msgBytes
	}
}
