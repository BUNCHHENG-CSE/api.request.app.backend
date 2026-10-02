package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig holds environment variables
type AppConfig struct {
	DatabaseURL          string
	ServerPort           string
	JWTSecret            string
	AllowedOrigins       string
	AllowPrivateRequests bool
	LocalCollectionsRoot string
}

// LoadConfig retrieves configuration from the environment
func LoadConfig() *AppConfig {
	_ = godotenv.Load()
	return &AppConfig{
		DatabaseURL:          getEnv("DATABASE_URL", "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"),
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		JWTSecret:            getEnv("JWT_SECRET", "super-secret-key"),
		AllowedOrigins:       getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		AllowPrivateRequests: getBoolEnv("ALLOW_PRIVATE_REQUESTS", false),
		LocalCollectionsRoot: getEnv("LOCAL_COLLECTIONS_ROOT", "./data/collections"),
	}
}

func getBoolEnv(key string, fallback bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
