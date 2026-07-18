package config

import (
	"os"
)

// AppConfig holds environment variables
type AppConfig struct {
	DatabaseURL string
	ServerPort  string
	JWTSecret   string
}

// LoadConfig retrieves configuration from the environment
func LoadConfig() *AppConfig {
	return &AppConfig{
		DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-key"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
