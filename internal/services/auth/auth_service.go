package auth

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Dongmoon29/code_racer_api/db/models"
	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/users"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

var (
	instance AuthService
	once     sync.Once
)

type AuthService struct {
	UserRepository users.UserRepositoryImpl
}

func NewAuthService(db *gorm.DB) AuthService {
	once.Do(func() {
		instance = AuthService{
			UserRepository: users.NewUserRepository(db),
		}
	})
	return instance
}

type Claims struct {
	UserID         string
	ExpirationTime time.Time
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// 비밀번호 비교
func checkPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// JWT 토큰 생성
func generateJWT(userID string) (string, error) {
	// 토큰 만료 시간 설정 (예: 1시간 후 만료)
	expirationTime := time.Now().Add(1 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": userID,
			"exp":     expirationTime.Unix(), // 만료 시간은 Unix 타임스탬프로 설정
		})

	// 비밀 키를 []byte로 변환
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (us *AuthService) UserSignin(dto dtos.SigninRequestDto) (*models.User, error) {
	// 1. 이메일을 기준으로 사용자 찾기
	user, err := us.UserRepository.FindByEmail(dto.Email)
	if err != nil {
		// 사용자를 찾을 수 없으면 에러 반환
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// 2. 입력된 비밀번호와 해시된 비밀번호 비교
	if !checkPasswordHash(dto.Password, user.Password) {
		// 비밀번호가 일치하지 않으면 에러 반환
		return nil, fmt.Errorf("invalid password")
	}

	// 3. 비밀번호가 일치하면 JWT 토큰 생성
	// token, err := generateJWT(fmt.Sprint(user.ID))
	// if err != nil {
	// 	return "", nil, fmt.Errorf("failed to generate token: %v", err)
	// }

	// 4. JWT 토큰과 사용자 객체 반환
	return user, nil
}

func (us *AuthService) UserSignup(dto dtos.SignupRequestDto) (*models.User, error) {
	hashedPassword, err := hashPassword(dto.Password)

	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:     dto.Name,
		Password: hashedPassword,
		Email:    dto.Email,
	}

	createdUser, err := us.UserRepository.CreateOne(user)
	if err != nil {
		log.Println("failed create user")
		return nil, err
	}

	return createdUser, nil
}
