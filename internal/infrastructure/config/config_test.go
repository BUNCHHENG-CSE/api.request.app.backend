package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear relevant env vars to test defaults
	origDBURL := os.Getenv("DATABASE_URL")
	origPort := os.Getenv("SERVER_PORT")
	defer func() {
		os.Setenv("DATABASE_URL", origDBURL)
		os.Setenv("SERVER_PORT", origPort)
	}()

	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("DATABASE_URL", "host=127.0.0.1 user=test dbname=test")

	cfg := LoadConfig()
	if cfg.ServerPort != "9999" {
		t.Fatalf("expected port 9999, got %s", cfg.ServerPort)
	}
	if cfg.DatabaseURL != "host=127.0.0.1 user=test dbname=test" {
		t.Fatalf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
}

func TestLoadConfigExpandsEnv(t *testing.T) {
	os.Setenv("MY_TEST_DB_HOST", "db.internal")
	defer os.Unsetenv("MY_TEST_DB_HOST")

	os.Setenv("DATABASE_URL", "host=${MY_TEST_DB_HOST} user=app dbname=app")
	defer os.Unsetenv("DATABASE_URL")

	cfg := LoadConfig()
	if cfg.DatabaseURL != "host=db.internal user=app dbname=app" {
		t.Fatalf("expected expanded URL, got %s", cfg.DatabaseURL)
	}
}
