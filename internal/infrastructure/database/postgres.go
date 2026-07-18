package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB establishes a connection to the PostgreSQL database using GORM.
func NewPostgresDB(dsn string) (*gorm.DB, error) {
	// Configure GORM to log SQL queries to the console (useful for debugging)
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to the PostgreSQL database")

	return db, nil
}
