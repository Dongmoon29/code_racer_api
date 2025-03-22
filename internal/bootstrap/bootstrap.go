package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/config"
	"github.com/Dongmoon29/code_racer_api/internal/mapper"
	"github.com/Dongmoon29/code_racer_api/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	authController "github.com/Dongmoon29/code_racer_api/internal/controllers/auth"
	gameController "github.com/Dongmoon29/code_racer_api/internal/controllers/game"
	judge0Controller "github.com/Dongmoon29/code_racer_api/internal/controllers/judge0"

	authService "github.com/Dongmoon29/code_racer_api/internal/services/auth"
	gameService "github.com/Dongmoon29/code_racer_api/internal/services/game"
	judge0Service "github.com/Dongmoon29/code_racer_api/internal/services/judge0"
)

const apiVersion = "v1"

func Mount(app *config.Application) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"DELETE", "POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
	}))

	apiGroup := r.Group(fmt.Sprintf("/api/%s", apiVersion))

	setJudge0Routes(app, apiGroup)
	setGameRoutes(app, apiGroup)
	setUserRoutes(app, apiGroup)
	return r
}

func Run(app *config.Application, router *gin.Engine) error {
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

func setUserRoutes(app *config.Application, rg *gin.RouterGroup) {
	us := authService.NewAuthService(app.Repository.UserRepository, app.Repository.RoleRepository, app.CacheStorage.Users, app.Logger)
	uc := authController.NewAuthController(us, app.Logger)

	cg := rg.Group("/users")
	{
		cg.POST("/signin", uc.HandleSignin)
		cg.POST("/signup", uc.HandleSignup)
		cg.POST("/logout", middlewares.AuthMiddleware(app), uc.HandleLogout)
		cg.GET("/profile", middlewares.AuthMiddleware(app), uc.HandleUserProfile)
	}
}

func setGameRoutes(app *config.Application, rg *gin.RouterGroup) {
	gs := gameService.NewGameService(app.CacheStorage.Games, app.Logger)
	gc := gameController.NewGameController(gs, app.Logger)

	gg := rg.Group("/games")
	gg.Use(middlewares.AuthMiddleware(app))
	{
		gg.GET("", gc.HandleGetGameRooms)
		gg.GET("/ws", gc.HandleGameWebSocket)
	}
}

func setJudge0Routes(app *config.Application, rg *gin.RouterGroup) {
	js := judge0Service.NewJudge0Service(app.Logger)
	jc := judge0Controller.NewJudge0Controller(js, app.Logger)

	jg := rg.Group("/code")
	jg.Use(middlewares.AuthMiddleware(app))
	{
		jg.GET("/about", jc.GetAbout)
		jg.POST("/submit", jc.HandleCreateCodeSubmission)
	}
}

func GetUser(app *config.Application, ctx context.Context, userID int) (*mapper.MappedUser, error) {
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
