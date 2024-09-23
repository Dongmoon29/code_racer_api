package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func ConnectDB() (*gorm.DB, error) {
	var err error

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	timezone := os.Getenv("DB_TIMEZONE")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable timezone=%s",
		host, port, user, dbname, password, timezone,
	)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	return db, err
}
