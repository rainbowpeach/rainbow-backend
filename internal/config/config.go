package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	Host          string
	Port          string
	Log           LogConfig
	Database      DatabaseConfig
	JWTSecret     string
	JWTExpiresIn  int64
	AdminUsername string
	AdminPassword string
	AllowOrigins  []string
	Upload        UploadConfig
	Scene         SceneConfig
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

type LogConfig struct {
	RootDir string
}

type UploadConfig struct {
	RootDir      string
	ImageMaxSize int64
	AudioMaxSize int64
}

type SceneConfig struct {
	EnablePublicOverride bool
}

func Load() (Config, error) {
	_ = godotenv.Load()

	appEnv := getEnv("APP_ENV", "dev")

	cfg := Config{
		AppEnv: appEnv,
		Host:   getEnv("HOST", "0.0.0.0"),
		Port:   getEnv("PORT", "8080"),
		Log: LogConfig{
			RootDir: filepath.Clean(getEnv("LOG_ROOT", defaultLogRoot(appEnv))),
		},
		Database: DatabaseConfig{
			Driver: strings.ToLower(getEnv("DB_DRIVER", "mysql")),
			DSN:    getEnv("DB_DSN", "root:password@tcp(127.0.0.1:3306)/rainbow?charset=utf8mb4&parseTime=True&loc=Local"),
		},
		JWTSecret:     getEnv("JWT_SECRET", "replace_with_a_strong_secret"),
		JWTExpiresIn:  getEnvInt64("JWT_EXPIRES_IN", 7200),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "change_me"),
		AllowOrigins:  splitAndTrim(getEnv("ALLOW_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		Upload: UploadConfig{
			RootDir:      filepath.Clean(getEnv("UPLOAD_ROOT", defaultUploadRoot(appEnv))),
			ImageMaxSize: getEnvInt64("UPLOAD_IMAGE_MAX_SIZE", 50*1024*1024),
			AudioMaxSize: getEnvInt64("UPLOAD_AUDIO_MAX_SIZE", 50*1024*1024),
		},
		Scene: SceneConfig{
			EnablePublicOverride: getEnvBool("ENABLE_PUBLIC_SCENE_OVERRIDE", false),
		},
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
	if c.Host == "" {
		return errors.New("HOST is required")
	}
	if c.Database.Driver != "mysql" {
		return fmt.Errorf("unsupported DB_DRIVER %q, only mysql is supported", c.Database.Driver)
	}
	if c.Database.DSN == "" {
		return errors.New("DB_DSN is required")
	}
	if strings.TrimSpace(c.Log.RootDir) == "" {
		return errors.New("LOG_ROOT is required")
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
	if strings.TrimSpace(c.Upload.RootDir) == "" {
		return errors.New("UPLOAD_ROOT is required")
	}
	if c.Upload.ImageMaxSize <= 0 {
		return errors.New("UPLOAD_IMAGE_MAX_SIZE must be greater than 0")
	}
	if c.Upload.AudioMaxSize <= 0 {
		return errors.New("UPLOAD_AUDIO_MAX_SIZE must be greater than 0")
	}
	if c.isProduction() {
		if isInsecureSecret(c.JWTSecret) {
			return errors.New("JWT_SECRET must be replaced with a strong secret in production")
		}
		if isInsecureAdminPassword(c.AdminPassword) {
			return errors.New("ADMIN_PASSWORD must be replaced with a strong password in production")
		}
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

func (c Config) isProduction() bool {
	switch strings.ToLower(c.AppEnv) {
	case "prod", "production":
		return true
	default:
		return false
	}
}

func (c Config) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
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

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
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

func isInsecureSecret(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == "replace_with_a_strong_secret"
}

func isInsecureAdminPassword(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == "change_me"
}

func defaultUploadRoot(appEnv string) string {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "prod", "production":
		return "/opt/rainbow-backend/uploads/prod"
	case "test":
		return "/opt/rainbow-backend/uploads/test"
	default:
		return "./uploads/dev"
	}
}

func defaultLogRoot(appEnv string) string {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "prod", "production":
		return "/opt/rainbow-backend/logs/prod"
	case "test":
		return "/opt/rainbow-backend/logs/test"
	default:
		return "./logs/dev"
	}
}
