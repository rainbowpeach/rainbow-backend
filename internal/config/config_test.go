package config

import "testing"

func TestValidateRejectsInsecureProductionSecrets(t *testing.T) {
	cfg := Config{
		AppEnv: "prod",
		Host:   "127.0.0.1",
		Port:   "8080",
		Log: LogConfig{
			RootDir: "./logs/prod",
		},
		Database: DatabaseConfig{
			Driver: "mysql",
			DSN:    "user:password@tcp(127.0.0.1:3306)/rainbow?charset=utf8mb4&parseTime=True&loc=Local",
		},
		JWTSecret:     "replace_with_a_strong_secret",
		JWTExpiresIn:  7200,
		AdminUsername: "admin",
		AdminPassword: "change_me",
		Upload: UploadConfig{
			RootDir:      "./uploads/dev",
			ImageMaxSize: 10 * 1024 * 1024,
			AudioMaxSize: 20 * 1024 * 1024,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected insecure production config to fail validation")
	}
}

func TestValidateAllowsLocalExampleSecrets(t *testing.T) {
	cfg := Config{
		AppEnv: "dev",
		Host:   "127.0.0.1",
		Port:   "8080",
		Log: LogConfig{
			RootDir: "./logs/dev",
		},
		Database: DatabaseConfig{
			Driver: "mysql",
			DSN:    "user:password@tcp(127.0.0.1:3306)/rainbow?charset=utf8mb4&parseTime=True&loc=Local",
		},
		JWTSecret:     "replace_with_a_strong_secret",
		JWTExpiresIn:  7200,
		AdminUsername: "admin",
		AdminPassword: "change_me",
		Upload: UploadConfig{
			RootDir:      "./uploads/dev",
			ImageMaxSize: 10 * 1024 * 1024,
			AudioMaxSize: 20 * 1024 * 1024,
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected local config to pass validation, got %v", err)
	}
}

func TestDefaultUploadRoot(t *testing.T) {
	tests := []struct {
		name   string
		appEnv string
		want   string
	}{
		{name: "dev", appEnv: "dev", want: "./uploads/dev"},
		{name: "test", appEnv: "test", want: "/opt/rainbow-backend/uploads/test"},
		{name: "prod", appEnv: "prod", want: "/opt/rainbow-backend/uploads/prod"},
		{name: "production", appEnv: "production", want: "/opt/rainbow-backend/uploads/prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultUploadRoot(tt.appEnv); got != tt.want {
				t.Fatalf("defaultUploadRoot(%q) = %q, want %q", tt.appEnv, got, tt.want)
			}
		})
	}
}

func TestDefaultLogRoot(t *testing.T) {
	tests := []struct {
		name   string
		appEnv string
		want   string
	}{
		{name: "dev", appEnv: "dev", want: "./logs/dev"},
		{name: "test", appEnv: "test", want: "/opt/rainbow-backend/logs/test"},
		{name: "prod", appEnv: "prod", want: "/opt/rainbow-backend/logs/prod"},
		{name: "production", appEnv: "production", want: "/opt/rainbow-backend/logs/prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultLogRoot(tt.appEnv); got != tt.want {
				t.Fatalf("defaultLogRoot(%q) = %q, want %q", tt.appEnv, got, tt.want)
			}
		})
	}
}

func TestValidateRejectsInvalidUploadConfig(t *testing.T) {
	cfg := Config{
		AppEnv: "dev",
		Host:   "127.0.0.1",
		Port:   "8080",
		Log: LogConfig{
			RootDir: "",
		},
		Database: DatabaseConfig{
			Driver: "mysql",
			DSN:    "user:password@tcp(127.0.0.1:3306)/rainbow?charset=utf8mb4&parseTime=True&loc=Local",
		},
		JWTSecret:     "replace_with_a_strong_secret",
		JWTExpiresIn:  7200,
		AdminUsername: "admin",
		AdminPassword: "change_me",
		Upload: UploadConfig{
			RootDir:      "",
			ImageMaxSize: 0,
			AudioMaxSize: 0,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid upload config to fail validation")
	}
}
