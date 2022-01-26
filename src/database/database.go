package database

import (
	"fmt"
	"os"

	"app/database/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() (db *gorm.DB) {

	if os.Getenv("POSTGRES_HOST") == "" ||
		os.Getenv("POSTGRES_USER") == "" ||
		os.Getenv("POSTGRES_PASSWORD") == "" ||
		os.Getenv("POSTGRES_DB") == "" {
		return nil
	}

	dsn := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
	dsn = fmt.Sprintf(dsn,
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Database is not connected")
	}

	DB = db

	Migrate()

	return db
}

func Migrate() {
	DB.AutoMigrate(&models.Envelope{})
}
