package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:255"` // 문자열 필드, 길이는 255자로 제한
	Email    string `gorm:"unique"`   // 이메일 필드, 중복 불가(unique)
	Password string `gorm:"size:255"` // 비밀번호 필드
}
