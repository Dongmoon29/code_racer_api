package models

import "time"

type GameState struct {
	ID        string
	RoomName  string
	Status    string
	CreatedAt time.Time
	CreatedBy string
	UserID    uint
}
