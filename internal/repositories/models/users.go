package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`                    // 기본 키
	Username  string    `gorm:"unique;not null" json:"username"`         // 고유한 사용자명
	Password  string    `gorm:"not null" json:"-"`                       // 비밀번호는 JSON 응답에서 제외
	Email     string    `gorm:"unique;not null" json:"email"`            // 이메일
	RoleID    uint      `gorm:"not null" json:"role_id"`                 // 역할 ID
	Role      *Role     `gorm:"foreignKey:RoleID" json:"role,omitempty"` // 역할 정보 (null일 경우 생략)
	IsActive  bool      `gorm:"default:true" json:"is_active"`           // 활성화 여부
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`        // 생성 시간
}
