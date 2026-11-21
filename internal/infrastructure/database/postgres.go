package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB establishes a connection to the PostgreSQL database using GORM with automatic retry.
func NewPostgresDB(dsn string) (*gorm.DB, error) {
	logLevel := logger.Info
	if os.Getenv("GIN_MODE") == "release" {
		logLevel = logger.Warn
	}
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	const maxRetries = 5
	const retryInterval = 1 * time.Second

	var (
		db  *gorm.DB
		err error
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			var sqlDBErr error
			sqlDB, err2 := db.DB()
			if err2 == nil {
				sqlDBErr = sqlDB.Ping()
			}
			if err2 == nil && sqlDBErr == nil {
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetMaxOpenConns(50)
				sqlDB.SetConnMaxIdleTime(10 * time.Minute)
				sqlDB.SetConnMaxLifetime(time.Hour)

				log.Println("Successfully connected to the PostgreSQL database")
				return db, nil
			}
			if err2 != nil {
				err = err2
			} else {
				err = sqlDBErr
			}
		}

		if attempt < maxRetries {
			log.Printf("Database connection attempt %d/%d failed: %v. Retrying in %v...", attempt, maxRetries, err, retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
