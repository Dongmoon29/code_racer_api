package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(getUser func(context.Context, int64) (*models.User, error)) gin.HandlerFunc {
	fmt.Println("AuthMiddleware called")
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is missing",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is malformed",
			})
			return
		}

		token := parts[1]
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token is empty",
			})
			return
		}

		// JWT 파싱 및 검증
		tokenString := parts[1]
		jwtToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 예제에서는 HS256 시크릿 키 사용 (실제 환경에서는 안전한 저장 방법 필요)
			return []byte("secret"), nil
		})
		if err != nil || !jwtToken.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}
		// Claims 파싱
		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			return
		}
		// 사용자 ID 추출
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user ID in token",
			})
			return
		}
		// 사용자 정보 로딩
		user, err := getUser(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "failed to load user",
			})
			return
		}
		// 사용자 정보를 컨텍스트에 저장
		c.Set(string("user"), user)

		c.Next()
	}
}
