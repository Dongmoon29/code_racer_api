package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	r := gin.Default()
	// CORS 설정
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},  // 클라이언트 도메인
		AllowMethods:     []string{"POST", "GET", "OPTIONS"}, // 허용할 메서드
		AllowHeaders:     []string{"Origin", "Content-Type"}, // 허용할 헤더
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	apiVersion := "v1"
	apiGroup := r.Group(fmt.Sprintf("/api/%s", apiVersion))

	app.setJudge0Routes(apiGroup)
	app.setGameRoutes(apiGroup)
	return r
}

func (app *App) setGameRoutes(rg *gin.RouterGroup) {
	gs := gameService.NewGameService()
	gc := gameController.NewGameController(gs)

	gg := rg.Group("/games")
	{
		// WebSocket을 통한 게임 로비 기능 라우트 추가
		gg.GET("/:id", gc.HandleWebSocket)
	}
}

func (app *App) setJudge0Routes(rg *gin.RouterGroup) {
	js := judge0Service.NewJudge0Service()
	jc := judge0Controller.NewJudge0Controller(js)

	cg := rg.Group("/code")
	{
		cg.GET("/about", jc.GetAbout)
		cg.POST("/submit", jc.HandleCreateCodeSubmission)
		cg.GET("/submit", jc.HandleGetCodeSubmission)
	}
}

func (app *App) Start() {
	envPath := filepath.Join(os.Getenv("PWD"), ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		fmt.Println("godotenv load error")
		log.Fatalln(err)
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
