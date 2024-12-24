package config

import (
	"fmt"
	"time"
)

type Config struct {
	DbConfig    DbConfig
	RedisConfig RedisConfig
	Addr        string
	Env         string
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
