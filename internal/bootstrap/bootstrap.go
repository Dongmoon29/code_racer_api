package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/config"
	"github.com/Dongmoon29/code_racer_api/internal/repositories"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/cache"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	authController "github.com/Dongmoon29/code_racer_api/internal/controllers/auth"
	gameController "github.com/Dongmoon29/code_racer_api/internal/controllers/game"
	judge0Controller "github.com/Dongmoon29/code_racer_api/internal/controllers/judge0"

	authService "github.com/Dongmoon29/code_racer_api/internal/services/auth"
	gameService "github.com/Dongmoon29/code_racer_api/internal/services/game"
	judge0Service "github.com/Dongmoon29/code_racer_api/internal/services/judge0"
)

type Application struct {
	Repository   repositories.Repository
	CacheStorage cache.Storage
	Config       *config.Config
	Logger       *zap.SugaredLogger
}

const apiVersion = "v1"

func (app *Application) Mount() *gin.Engine {
	// webHost := env.GetString("CODERACER_WEB", "localhost:3000")
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{webHost},
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"DELETE", "POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	apiGroup := r.Group(fmt.Sprintf("/api/%s", apiVersion))

	app.setJudge0Routes(apiGroup)
	app.setGameRoutes(apiGroup)
	app.setUserRoutes(apiGroup)
	return r
}

func (app *Application) Run(router *gin.Engine) error {
	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      router,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.Logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.Logger.Infow("server has started", "addr", app.Config.Addr, "env", app.Config.Env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) setUserRoutes(rg *gin.RouterGroup) {
	us := authService.NewAuthService(app.Repository.UserRepository, app.Repository.RoleRepository)
	uc := authController.NewAuthController(us)

	cg := rg.Group("/users")
	{
		cg.POST("/signin", uc.HandleSignin)
		cg.POST("/signup", uc.HandleSignup)
	}
}

func (app *Application) setJudge0Routes(rg *gin.RouterGroup) {
	js := judge0Service.NewJudge0Service()
	jc := judge0Controller.NewJudge0Controller(js)

	jg := rg.Group("/code")
	jg.Use(app.AuthMiddleware())
	{
		jg.GET("/about", jc.GetAbout)
		jg.POST("/submit", jc.HandleCreateCodeSubmission)
	}
}

func (app *Application) setGameRoutes(rg *gin.RouterGroup) {
	gs := gameService.NewGameService()
	gc := gameController.NewGameController(gs)

	gg := rg.Group("/games")
	gg.Use(app.AuthMiddleware())
	{
		gg.GET("", gc.HandleGetGameRooms)
		gg.POST("", gc.HandleCreateGameRoom)
		gg.GET("/:id", gc.HandleJoinGameRoom)
	}
}

func (app *Application) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	if !app.Config.RedisConfig.Enabled {
		return app.Repository.UserRepository.GetByID(ctx, userID)
	}

	user, err := app.CacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = app.Repository.UserRepository.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if err := app.CacheStorage.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *Application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization 헤더 확인
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is missing",
			})
			return
		}

		// 헤더 포맷 검증
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is malformed",
			})
			return
		}

		// 토큰 추출 및 검증 (예제에서는 단순 토큰 처리)
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
		user, err := app.GetUser(c.Request.Context(), userID)
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