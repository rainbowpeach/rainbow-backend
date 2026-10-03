package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
)

type stubMiniProgramUserRepo struct {
	user          *model.User
	openID        string
	updateToken   string
	tokenExpireAt time.Time
	lastLoginAt   time.Time
}

func (r *stubMiniProgramUserRepo) GetOrCreateByOpenID(_ context.Context, openID string) (*model.User, error) {
	r.openID = openID
	if r.user == nil {
		r.user = &model.User{ID: 7, OpenID: openID}
	}
	return r.user, nil
}

func (r *stubMiniProgramUserRepo) UpdateLogin(_ context.Context, userID uint, token string, tokenExpireAt, lastLoginAt time.Time) error {
	if r.user == nil || r.user.ID != userID {
		return errors.New("unexpected user")
	}
	r.updateToken = token
	r.tokenExpireAt = tokenExpireAt
	r.lastLoginAt = lastLoginAt
	return nil
}

func (r *stubMiniProgramUserRepo) GetByToken(context.Context, string) (*model.User, error) {
	return r.user, nil
}

type stubCode2SessionClient struct {
	openID string
	err    error
}

func (c stubCode2SessionClient) CodeToOpenID(context.Context, string) (string, error) {
	return c.openID, c.err
}

func TestMiniProgramAuthServiceLogin(t *testing.T) {
	users := &stubMiniProgramUserRepo{}
	auth := NewMiniProgramAuthService(users, stubCode2SessionClient{openID: "openid-1"}, 30*24*60*60)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	auth.now = func() time.Time { return now }

	result, err := auth.Login(context.Background(), "wx-code")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token")
	}
	if len(result.Token) < 40 {
		t.Fatalf("expected secure random token, got %q", result.Token)
	}
	if result.ExpiresIn != 30*24*60*60 {
		t.Fatalf("ExpiresIn = %d", result.ExpiresIn)
	}
	if users.openID != "openid-1" {
		t.Fatalf("openID = %q", users.openID)
	}
	if users.updateToken != result.Token {
		t.Fatal("saved token does not match response token")
	}
	if !users.tokenExpireAt.Equal(now.Add(30 * 24 * time.Hour)) {
		t.Fatalf("token expiry = %v", users.tokenExpireAt)
	}
	if !users.lastLoginAt.Equal(now) {
		t.Fatalf("last login = %v", users.lastLoginAt)
	}
}

func TestMiniProgramAuthServiceRejectsEmptyCode(t *testing.T) {
	auth := NewMiniProgramAuthService(&stubMiniProgramUserRepo{}, stubCode2SessionClient{}, 3600)

	_, err := auth.Login(context.Background(), " ")
	if !errors.Is(err, ErrInvalidMiniProgramCode) {
		t.Fatalf("expected ErrInvalidMiniProgramCode, got %v", err)
	}
}

func TestWeChatCode2SessionAPIClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("appid") != "app-id" ||
			r.URL.Query().Get("secret") != "app-secret" ||
			r.URL.Query().Get("js_code") != "code" ||
			r.URL.Query().Get("grant_type") != "authorization_code" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"openid":"openid-1","session_key":"not-returned"}`))
	}))
	defer server.Close()

	client := NewWeChatCode2SessionAPIClient(
		config.WeChatMiniProgramConfig{
			AppID:             "app-id",
			AppSecret:         "app-secret",
			JSCode2SessionURL: server.URL,
		},
		server.Client(),
	)

	openID, err := client.CodeToOpenID(context.Background(), "code")
	if err != nil {
		t.Fatalf("CodeToOpenID() error = %v", err)
	}
	if openID != "openid-1" {
		t.Fatalf("openID = %q", openID)
	}
	if strings.Contains(openID, "session_key") {
		t.Fatal("session_key must not be returned")
	}
}
