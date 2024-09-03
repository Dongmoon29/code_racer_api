package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Dongmoon29/codeRacer_api/db"
	controllers "github.com/Dongmoon29/codeRacer_api/internal/controllers/judge0"
	"github.com/Dongmoon29/codeRacer_api/internal/services"
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
	apiVersion := "v1"
	apiGroup := r.Group(fmt.Sprintf("/api/%s", apiVersion))

	app.setJudge0Routes(apiGroup)
	return r
}

func (app *App) setJudge0Routes(rg *gin.RouterGroup) {
	js := services.NewJudge0Service()
	jc := controllers.NewJudge0Controller(js)

	cg := rg.Group("/code")
	{
		cg.GET("/about", jc.GetAbout)

		cg.POST("/submit", jc.CreateCodeSubmission)
		cg.GET("/submit", jc.GetCodeSubmission)
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
