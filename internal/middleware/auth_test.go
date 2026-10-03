package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rainbow-backend/internal/model"
)

type stubUserTokenRepo struct {
	user *model.User
	err  error
}

func (r stubUserTokenRepo) GetOrCreateByOpenID(context.Context, string) (*model.User, error) {
	return r.user, r.err
}

func (r stubUserTokenRepo) UpdateLogin(context.Context, uint, string, time.Time, time.Time) error {
	return nil
}

func (r stubUserTokenRepo) GetByToken(context.Context, string) (*model.User, error) {
	return r.user, r.err
}

func TestUserTokenAuthSetsUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expiresAt := time.Now().Add(time.Hour)
	engine := gin.New()
	engine.Use(UserTokenAuth(stubUserTokenRepo{
		user: &model.User{
			ID:            42,
			TokenExpireAt: &expiresAt,
		},
	}))
	engine.GET("/protected", func(c *gin.Context) {
		userID, ok := UserID(c)
		if !ok || userID != 42 {
			t.Fatalf("UserID() = %d, %v", userID, ok)
		}
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer user-token")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestUserTokenAuthRejectsMissingExpiredAndUnknownTokens(t *testing.T) {
	expiredAt := time.Now().Add(-time.Second)
	tests := []struct {
		name string
		repo stubUserTokenRepo
	}{
		{name: "missing", repo: stubUserTokenRepo{}},
		{name: "unknown", repo: stubUserTokenRepo{err: gorm.ErrRecordNotFound}},
		{name: "expired", repo: stubUserTokenRepo{user: &model.User{ID: 1, TokenExpireAt: &expiredAt}}},
		{name: "repository error", repo: stubUserTokenRepo{err: errors.New("db down")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.Use(UserTokenAuth(tt.repo))
			engine.GET("/protected", func(c *gin.Context) {
				t.Fatal("protected handler should not be called")
			})

			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.name != "missing" {
				request.Header.Set("Authorization", "Bearer user-token")
			}
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}
		})
	}
}
