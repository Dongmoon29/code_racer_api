package utils

import (
	"fmt"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/env"
	"github.com/golang-jwt/jwt/v5"
)

// JWT 토큰 생성
func GenerateJWT(userID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": userID,
			"exp":     expirationTime.Unix(),
		})

	// 환경 변수에서 JWT_SECRET 키 가져오기
	secretKey := env.GetString("JWT_SECRET", "secret")
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// JWT 토큰 검증
func VerifyToken(tokenString string) error {
	secretKey := env.GetString("JWT_SECRET", "secret")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 서명 방법 확인
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
