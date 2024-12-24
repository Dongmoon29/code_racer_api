package models

import (
	"time"
)

// User 모델 정의
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	Email     string    `gorm:"unique;not null"`
	RoleID    uint      `gorm:"not null"`
	Role      *Role     `gorm:"foreignKey:RoleID"`
	IsActive  bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
