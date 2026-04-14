package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	Port          string
	Database      DatabaseConfig
	JWTSecret     string
	JWTExpiresIn  int64
	AdminUsername string
	AdminPassword string
	AllowOrigins  []string
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv: getEnv("APP_ENV", "dev"),
		Port:   getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Driver: strings.ToLower(getEnv("DB_DRIVER", "mysql")),
			DSN:    getEnv("DB_DSN", "root:password@tcp(127.0.0.1:3306)/rainbow?charset=utf8mb4&parseTime=True&loc=Local"),
		},
		JWTSecret:     getEnv("JWT_SECRET", "replace_with_a_strong_secret"),
		JWTExpiresIn:  getEnvInt64("JWT_EXPIRES_IN", 7200),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "change_me"),
		AllowOrigins:  splitAndTrim(getEnv("ALLOW_ORIGINS", "http://localhost:3000,http://localhost:5173")),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Port == "" {
		return errors.New("PORT is required")
	}
	if c.Database.Driver != "mysql" {
		return fmt.Errorf("unsupported DB_DRIVER %q, only mysql is supported", c.Database.Driver)
	}
	if c.Database.DSN == "" {
		return errors.New("DB_DSN is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.JWTExpiresIn <= 0 {
		return errors.New("JWT_EXPIRES_IN must be greater than 0")
	}
	if c.AdminUsername == "" {
		return errors.New("ADMIN_USERNAME is required")
	}
	if c.AdminPassword == "" {
		return errors.New("ADMIN_PASSWORD is required")
	}

	return nil
}

func (c Config) GinMode() string {
	switch strings.ToLower(c.AppEnv) {
	case "prod", "production":
		return "release"
	case "test":
		return "test"
	default:
		return "debug"
	}
}

func (c Config) Address() string {
	return ":" + c.Port
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}

	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return parsed
		}
	}

	return fallback
}

func splitAndTrim(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		result = append(result, item)
	}

	return result
}
