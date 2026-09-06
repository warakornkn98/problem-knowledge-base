// Package config loads runtime configuration from environment variables
// (optionally sourced from a local .env file).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config is the fully resolved application configuration.
type Config struct {
	AppEnv  string
	AppPort string

	DB DBConfig

	CORSOrigins []string

	JWTSecret string
	JWTExpiry time.Duration

	SeedAdminUsername string
	SeedAdminEmail    string
	SeedAdminPassword string

	AutoMigrate bool
	AutoSeed    bool
}

// DBConfig holds PostgreSQL connection settings.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN renders a lib/pq style connection string.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// Load reads configuration. It first tries to load a .env file from the current
// directory and its parents (handy when running `go run ./cmd/...`), then reads
// the process environment which always wins.
func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		AppEnv:  getenv("APP_ENV", "development"),
		AppPort: getenv("APP_PORT", "8080"),
		DB: DBConfig{
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenv("DB_PORT", "5432"),
			User:     firstNonEmpty(os.Getenv("DB_USER"), os.Getenv("POSTGRES_USER"), "pkb"),
			Password: firstNonEmpty(os.Getenv("DB_PASSWORD"), os.Getenv("POSTGRES_PASSWORD"), "pkb_secret"),
			Name:     firstNonEmpty(os.Getenv("DB_NAME"), os.Getenv("POSTGRES_DB"), "pkb"),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
		},
		CORSOrigins:       splitAndTrim(getenv("CORS_ORIGINS", "*")),
		JWTSecret:         getenv("JWT_SECRET", "insecure-dev-secret-change-me"),
		JWTExpiry:         time.Duration(getenvInt("JWT_EXPIRY_HOURS", 72)) * time.Hour,
		SeedAdminUsername: getenv("SEED_ADMIN_USERNAME", "admin"),
		SeedAdminEmail:    getenv("SEED_ADMIN_EMAIL", "admin@example.com"),
		SeedAdminPassword: getenv("SEED_ADMIN_PASSWORD", "admin1234"),
		AutoMigrate:       getenvBool("AUTO_MIGRATE", false),
		AutoSeed:          getenvBool("AUTO_SEED", false),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}
	return cfg, nil
}

// IsProduction reports whether the app runs in a production-like environment.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

func loadDotEnv() {
	candidates := []string{".env", "../.env", "../../.env"}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			return
		}
	}
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
