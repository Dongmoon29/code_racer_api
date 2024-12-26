package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/config"
	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/Dongmoon29/code_racer_api/internal/repositories"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/cache"
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
	CacheStorage cache.RedisStorage
	Config       *config.Config
	Logger       *zap.SugaredLogger
}

const apiVersion = "v1"

func (app *Application) Mount() *gin.Engine {
	// webHost := env.GetString("CODERACER_WEB", "localhost:3000")
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"*"},
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"DELETE", "POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
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
	us := authService.NewAuthService(app.Repository.UserRepository, app.Repository.RoleRepository, app.CacheStorage.Users, app.Logger)
	uc := authController.NewAuthController(us, app.Logger)

	cg := rg.Group("/users")
	{
		cg.POST("/signin", uc.HandleSignin)
		cg.POST("/signup", uc.HandleSignup)
		cg.POST("/logout", app.AuthMiddleware(), uc.HandleLogout)
		cg.GET("/profile", app.AuthMiddleware(), uc.HandleUserProfile)
	}
}

func (app *Application) setGameRoutes(rg *gin.RouterGroup) {
	gs := gameService.NewGameService(app.CacheStorage.Games, app.Logger)
	gc := gameController.NewGameController(gs, app.Logger)

	gg := rg.Group("/games")
	gg.Use(app.AuthMiddleware())
	{
		gg.GET("", gc.HandleGetGameRooms)
		gg.POST("", gc.HandleCreateGameRoom)
		gg.GET("/:id", gc.HandleJoinGameRoom)
	}
}

func (app *Application) setJudge0Routes(rg *gin.RouterGroup) {
	js := judge0Service.NewJudge0Service(app.Logger)
	jc := judge0Controller.NewJudge0Controller(js, app.Logger)

	jg := rg.Group("/code")
	jg.Use(app.AuthMiddleware())
	{
		jg.GET("/about", jc.GetAbout)
		jg.POST("/submit", jc.HandleCreateCodeSubmission)
	}
}

func (app *Application) GetUser(ctx context.Context, userID int) (*mapper.MappedUser, error) {
	if !app.Config.RedisConfig.Enabled {
		return app.getUserFromDB(ctx, userID)
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

func (app *Application) AuthMiddleware() gin.HandlerFunc {
	app.Logger.Debugln("AuthMiddleware() called")
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
			return
		}

		jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
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
		}
		user, err := app.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to load user"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func (app *Application) getUserFromDB(ctx context.Context, userID int) (*mapper.MappedUser, error) {
	user, err := app.Repository.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapper.UserMapper(user), nil
}
