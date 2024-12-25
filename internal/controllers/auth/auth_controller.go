package auth

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/Dongmoon29/code_racer_api/internal/services/auth"
	utils "github.com/Dongmoon29/code_racer_api/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	AuthService auth.AuthService
	logger      *zap.SugaredLogger
}

var (
	instance *AuthController
	once     sync.Once
)

func NewAuthController(authService auth.AuthService, logger *zap.SugaredLogger) *AuthController {
	once.Do(func() {
		instance = &AuthController{
			AuthService: authService,
			logger:      logger,
		}
	})
	return instance
}

func (uc *AuthController) HandleSignup(c *gin.Context) {
	var dto dtos.SignupRequestDto

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := uc.AuthService.CreateUser(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error signup"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true, "user": user})

}

func (uc *AuthController) HandleLogout(c *gin.Context) {
	userData, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userMap, ok := userData.(*mapper.MappedUser)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user data"})
		return
	}

	userID := userMap.ID
	uc.AuthService.DeleteSession(c.Request.Context(), int(userID))

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (uc *AuthController) HandleUserProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uc *AuthController) HandleSignin(c *gin.Context) {
	var signinRequestDto dtos.SigninRequestDto

	// 1. 요청 바인딩 및 검증
	if err := c.ShouldBindJSON(&signinRequestDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 2. 사용자 검증
	user, err := uc.AuthService.FindAndVerifyUserByEmail(signinRequestDto)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 3. JWT 토큰 생성
	token, err := utils.GenerateJWT(fmt.Sprint(user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 4. Redis 세션 저장
	err = uc.AuthService.SaveSession(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	// 5. 쿠키 설정
	c.SetCookie(
		"token",
		token,
		3600,
		"/",
		"",
		false, // Secure = false (HTTPS가 아닌 HTTP에서 테스트용)
		true,  // HttpOnly
	)
	c.JSON(http.StatusOK, gin.H{"ok": true, "user": user})
}
