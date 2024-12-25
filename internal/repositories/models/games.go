package models

import "time"

type GameState struct {
	ID        string    `json:"id"`
	RoomName  string    `json:"room_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	UserID    uint      `json:"user_id"`
}
