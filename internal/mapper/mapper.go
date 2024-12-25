package mapper

import (
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
)

type MappedUser struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RoleID    uint      `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

func UserMapper(u *models.User) *MappedUser {
	return &MappedUser{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		RoleID:    u.RoleID,
		CreatedAt: u.CreatedAt,
	}
}
