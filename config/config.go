package config

import (
	"fmt"
	"log"
	"os"

	"github.com/Dongmoon29/code_racer_api/db"
	gameController "github.com/Dongmoon29/code_racer_api/internal/controllers/game"
	judge0Controller "github.com/Dongmoon29/code_racer_api/internal/controllers/judge0"
	gameService "github.com/Dongmoon29/code_racer_api/internal/services/game"
	judge0Service "github.com/Dongmoon29/code_racer_api/internal/services/judge0"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
	db     *gorm.DB
}

func (app *App) initialize() {
	db, err := db.ConnectDB()

	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	app.db = db
	app.router = app.setRoutes()
}

func (app *App) setRoutes() *gin.Engine {
	webHost := os.Getenv("CODERACER_WEB")
	r := gin.Default()
	// setup CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{webHost},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	apiVersion := "v1"
	apiGroup := r.Group(fmt.Sprintf("/api/%s", apiVersion))

	app.setJudge0Routes(apiGroup)
	app.setGameRoutes(apiGroup)
	return r
}

func (app *App) setJudge0Routes(rg *gin.RouterGroup) {
	js := judge0Service.NewJudge0Service()
	jc := judge0Controller.NewJudge0Controller(js)

	cg := rg.Group("/code")
	{
		cg.GET("/about", jc.GetAbout)
		cg.POST("/submit", jc.HandleCreateCodeSubmission)
	}
}

func (app *App) setGameRoutes(rg *gin.RouterGroup) {
	gs := gameService.NewGameService()
	gc := gameController.NewGameController(gs)

	gg := rg.Group("/games")
	{
		gg.GET("", gc.HandleGetGameRooms)
		gg.POST("", gc.HandleCreateGameRoom)
		gg.GET("/:id", gc.HandleJoinGameRoom)
	}
}

func (app *App) Start() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("godotenv load error")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	app.initialize()

	err = app.router.Run(port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("Running service on port: %s", port)
}
