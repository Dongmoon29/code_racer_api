package auth

import (
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/services/auth"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	AuthService auth.AuthService
}

var (
	instance *AuthController
	once     sync.Once
)

func NewAuthController(authService auth.AuthService) *AuthController {
	once.Do(func() {
		instance = &AuthController{
			AuthService: authService,
		}
	})
	return instance
}

func (uc *AuthController) HandleSignup(c *gin.Context) {
	var dto dtos.SignupRequestDto

	// Handle request
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := uc.AuthService.UserSignup(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error signup"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true, "user": user})

}

func (uc *AuthController) HandleSignin(c *gin.Context) {
	var signinRequestDto dtos.SigninRequestDto

	// 1. 요청 바디에서 SigninRequestDto에 데이터 바인딩
	if err := c.ShouldBindJSON(&signinRequestDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 2. AuthService를 이용해 로그인 처리
	user, err := uc.AuthService.UserSignin(signinRequestDto)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공적으로 로그인하면 JWT 토큰 반환
	c.JSON(http.StatusOK, gin.H{"ok": true, "user": user})
}
