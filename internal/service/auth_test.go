package service

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type stubAdminRepo struct {
	admin *model.Admin
	err   error
}

func (r *stubAdminRepo) GetByUsername(context.Context, string) (*model.Admin, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.admin, nil
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	admin := &model.Admin{ID: 1, Username: "admin"}
	if err := admin.SetPassword("123456"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	service := NewAuthService(&stubAdminRepo{admin: admin}, NewTokenManager("secret", 7200))
	result, err := service.Login(context.Background(), "admin", "123456")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.Token == "" {
		t.Fatal("expected token to be generated")
	}
	if result.ExpiresIn != 7200 {
		t.Fatalf("expected ExpiresIn 7200, got %d", result.ExpiresIn)
	}
}

func TestAuthServiceLoginUnauthorized(t *testing.T) {
	service := NewAuthService(&stubAdminRepo{err: gorm.ErrRecordNotFound}, NewTokenManager("secret", 7200))
	_, err := service.Login(context.Background(), "admin", "123456")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestTokenManagerGenerateAndParse(t *testing.T) {
	manager := NewTokenManager("secret", 7200)
	admin := &model.Admin{ID: 2, Username: "root"}

	token, err := manager.Generate(admin)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if claims.AdminID != admin.ID {
		t.Fatalf("expected admin id %d, got %d", admin.ID, claims.AdminID)
	}
	if claims.Username != admin.Username {
		t.Fatalf("expected username %q, got %q", admin.Username, claims.Username)
	}
}
