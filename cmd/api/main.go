package main

import (
	"log"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/bootstrap"
	"github.com/Dongmoon29/code_racer_api/internal/config"
	"github.com/Dongmoon29/code_racer_api/internal/db"
	"github.com/Dongmoon29/code_racer_api/internal/env"
	"github.com/Dongmoon29/code_racer_api/internal/repositories"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/cache"
	"github.com/Dongmoon29/code_racer_api/internal/services/game"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	duration, _ := time.ParseDuration("15m")
	gameManager := game.NewGameManager()
	// run gameManager
	go gameManager.Run()

	cfg := &config.Config{
		DbConfig: config.DbConfig{
			Host:         env.GetString("DB_HOST", "localhost"),
			User:         env.GetString("DB_USER", "postgres"),
			Password:     env.GetString("DB_PASSWORD", "password1234"),
			Dbname:       env.GetString("DB_NAME", "code_racer_db"),
			Port:         env.GetInt("DB_PORT", 5432),
			Timezone:     env.GetString("DB_TIMEZONE", "Asia/seoul"),
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			MaxIdleTime:  duration,
		},
		RedisConfig: config.RedisConfig{
			Addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			Enabled:  env.GetBool("REDIS_ENABLED", true),
			Password: env.GetString("REDIS_PASSWORD", ""),
			Db:       env.GetInt("REDIS_DB", 0),
		},

		Addr: env.GetString("ADDR", ":8080"),
		Env:  env.GetString("ENV", "dev"),
	}
	db, err := db.New(cfg.DbConfig)
	if err != nil {
		log.Fatalln(err.Error())
	}

	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}()

	repository := repositories.NewRepository(db)

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	defer sugar.Sync()

	var rdb *redis.Client
	if cfg.RedisConfig.Enabled {
		rdb = cache.NewRedisClient(cfg.RedisConfig.Addr, cfg.RedisConfig.Password, cfg.RedisConfig.Db)
		logger.Info("redis connection established")

		defer rdb.Close()
	}
	cacheStorage := cache.NewRedisStorage(rdb)

	app := &config.Application{
		Logger:       sugar,
		Config:       cfg,
		Repository:   repository,
		CacheStorage: cacheStorage,
		GameManager:  gameManager,
	}

	router := bootstrap.Mount(app)
	if err := bootstrap.Run(app, router); err != nil {
		log.Fatal(err)
	}
}
