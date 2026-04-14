package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

type TokenManager struct {
	secret    []byte
	expiresIn int64
}

type AuthService struct {
	adminRepo repo.AdminRepository
	tokens    TokenManager
}

type AdminClaims struct {
	AdminID  uint   `json:"adminId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var ErrUnauthorized = errors.New("unauthorized")

func NewTokenManager(secret string, expiresIn int64) TokenManager {
	return TokenManager{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

func NewAuthService(adminRepo repo.AdminRepository, tokens TokenManager) *AuthService {
	return &AuthService{
		adminRepo: adminRepo,
		tokens:    tokens,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*model.LoginResponse, error) {
	admin, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("get admin by username: %w", err)
	}

	if !admin.CheckPassword(password) {
		return nil, ErrUnauthorized
	}

	token, err := s.tokens.Generate(admin)
	if err != nil {
		return nil, fmt.Errorf("generate jwt: %w", err)
	}

	return &model.LoginResponse{
		Token:     token,
		ExpiresIn: s.tokens.expiresIn,
	}, nil
}

func (m TokenManager) Generate(admin *model.Admin) (string, error) {
	now := time.Now()
	claims := AdminClaims{
		AdminID:  admin.ID,
		Username: admin.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", admin.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expiresIn) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m TokenManager) Parse(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, ErrUnauthorized
	}

	return claims, nil
}
