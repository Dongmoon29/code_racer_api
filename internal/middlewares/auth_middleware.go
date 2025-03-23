package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Dongmoon29/code_racer_api/internal/config"
	"github.com/Dongmoon29/code_racer_api/internal/env"
	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(app *config.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		app.Logger.Debugln("AuthMiddleware 시작")

		// 요청 정보 출력
		fmt.Printf("URL: %s\n", c.Request.URL.Path)
		fmt.Printf("Method: %s\n", c.Request.Method)
		fmt.Printf("Headers: %v\n", c.Request.Header)

		// 나머지 코드...

		token, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
			return
		}

		secretKey := env.GetString("JWT_SECRET", "secret")
		jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !jwtToken.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userID, err := strconv.Atoi(claims["user_id"].(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		user, err := getUser(app, c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to load user"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func getUser(app *config.Application, ctx context.Context, userID int) (*mapper.MappedUser, error) {
	if !app.Config.RedisConfig.Enabled {
		return getUserFromDB(app, ctx, userID)
	}

	user, err := app.CacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err := app.Repository.UserRepository.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		mappedUser := mapper.UserMapper(user)
		if err := app.CacheStorage.Users.Set(ctx, mappedUser); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func getUserFromDB(app *config.Application, ctx context.Context, userID int) (*mapper.MappedUser, error) {
	user, err := app.Repository.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapper.UserMapper(user), nil
}
