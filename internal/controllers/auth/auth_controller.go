package auth

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/services/auth"
	utils "github.com/Dongmoon29/code_racer_api/internal/utils/auth"
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

func (uc *AuthController) HandleSignin(c *gin.Context) {
	var signinRequestDto dtos.SigninRequestDto

	if err := c.ShouldBindJSON(&signinRequestDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := uc.AuthService.FindAndVerifyUserByEmail(signinRequestDto)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(fmt.Sprint(user.ID))
	if err != nil {
		fmt.Printf("error ===> %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	userResponse := dtos.UserResponseDto{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		RoleID:    user.RoleID,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "user": userResponse, "token": token})
}
