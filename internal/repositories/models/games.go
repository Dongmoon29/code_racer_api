package models

import "time"

type RedisGameRoom struct {
	ID        string    `json:"id"`
	RoomName  string    `json:"room_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	OwnedBy   uint      `json:"created_by"`
}
