package config

import (
	"fmt"
	"time"

	"github.com/Dongmoon29/code_racer_api/internal/repositories"
	"github.com/Dongmoon29/code_racer_api/internal/repositories/cache"
	"github.com/Dongmoon29/code_racer_api/internal/services/game"
	"go.uber.org/zap"
)

type Application struct {
	Repository   repositories.Repository
	CacheStorage cache.RedisStorage
	Config       *Config
	Logger       *zap.SugaredLogger
	GameManager  *game.GameManager
}

type Config struct {
	DbConfig    DbConfig
	RedisConfig RedisConfig
	Addr        string
	Env         string
	GameManager *game.GameManager
}

type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	Db       int
}

type DbConfig struct {
	Host         string
	User         string
	Password     string
	Dbname       string
	Port         int
	Timezone     string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

func (cfg *DbConfig) GetPostgresDsn() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable timezone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Dbname, cfg.Password, cfg.Timezone,
	)
	return dsn
}
