package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "secret")
	dbName := getEnv("DB_NAME", "mydb")
	dbPort := getEnv("DB_PORT", "5432")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	defaultDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	rawDatabaseURL := getEnv("DATABASE_URL", defaultDSN)
	databaseURL := os.ExpandEnv(rawDatabaseURL)
	// If the database URL still contains unresolved ${...} variables, fallback to assembled DSN
	if strings.Contains(databaseURL, "${") {
		databaseURL = defaultDSN
	}

	serverPort := getEnv("PORT", getEnv("SERVER_PORT", "8080"))
	allowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", getEnv("ALLOWED_ORIGINS", "http://localhost:3000"))

	return &AppConfig{
		DatabaseURL:          databaseURL,
		ServerPort:           serverPort,
		JWTSecret:            getEnv("JWT_SECRET", "super-secret-key"),
		AllowedOrigins:       allowedOrigins,
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
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
