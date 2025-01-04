package game

// 델타 전송: 이전 텍스트와 비교하여 변경된 부분(diff)만 전송하는 것이 더 효율적입니다.
// 예를 들어, diff-match-pattern 같은 라이브러리를 사용하여 텍스트 차이를 계산하고,
// 삽입/삭제된 텍스트와 위치 정보만 전송할 수 있습니다.
// 이렇게 하면 네트워크 트래픽을 줄이고 성능을 향상시킬 수 있습니ek.

import (
	"encoding/json"
	"log"
)



// Game represents the game state.
type Game struct {
	// TODO: implement game structure here
	editors []Editor
}

type Editor struct {
	id   string `json:"id"`
	code string `json:"code"`
}

type GameMessageType string

// Message represents a message sent between server and client.
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func createRoomMessage(room *Room, player *Player) []byte {
	// Create a map to hold the necessary information
	payload := map[string]interface{}{
		"roomID": room.ID,
		"userID": player.ID,
		"isHost": player.IsHost,
	}

	// Create the message
	msg := Message{
		Type:    "roomInfo",
		Payload: payload,
	}

	// Marshal the message to JSON
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling roomInfo message:", err)
		return nil // Or handle the error as appropriate for your application
	}

	return msgBytes
}

