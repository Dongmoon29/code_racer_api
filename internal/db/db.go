package db

import (
	"fmt"

	"github.com/Dongmoon29/code_racer_api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func New(cfg config.DbConfig) (*gorm.DB, error) {

	var err error
	dsn := cfg.GetPostgresDsn()
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	sqlDb, err := db.DB()

	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}
	sqlDb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDb.SetConnMaxIdleTime(cfg.MaxIdleTime)

	return db, nil
}
