package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
	"rainbow-backend/internal/repo"
)

var (
	ErrInvalidMiniProgramCode = errors.New("invalid mini program code")
	ErrWeChatLoginFailed      = errors.New("wechat login failed")
)

type WeChatCode2SessionClient interface {
	CodeToOpenID(ctx context.Context, code string) (string, error)
}

type WeChatCode2SessionAPIClient struct {
	appID     string
	appSecret string
	endpoint  string
	client    *http.Client
}

type MiniProgramAuthService struct {
	users          repo.UserRepository
	weChat         WeChatCode2SessionClient
	tokenExpiresIn int64
	now            func() time.Time
}

func NewWeChatCode2SessionAPIClient(cfg config.WeChatMiniProgramConfig, client *http.Client) *WeChatCode2SessionAPIClient {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return &WeChatCode2SessionAPIClient{
		appID:     cfg.AppID,
		appSecret: cfg.AppSecret,
		endpoint:  cfg.JSCode2SessionURL,
		client:    client,
	}
}

func NewMiniProgramAuthService(users repo.UserRepository, weChat WeChatCode2SessionClient, tokenExpiresIn int64) *MiniProgramAuthService {
	return &MiniProgramAuthService{
		users:          users,
		weChat:         weChat,
		tokenExpiresIn: tokenExpiresIn,
		now:            time.Now,
	}
}

func (s *MiniProgramAuthService) Login(ctx context.Context, code string) (*model.MiniProgramLoginResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrInvalidMiniProgramCode
	}

	openID, err := s.weChat.CodeToOpenID(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrWeChatLoginFailed, err)
	}
	if strings.TrimSpace(openID) == "" {
		return nil, ErrWeChatLoginFailed
	}

	user, err := s.users.GetOrCreateByOpenID(ctx, openID)
	if err != nil {
		return nil, fmt.Errorf("get or create user by openid: %w", err)
	}

	token, err := generateUserToken()
	if err != nil {
		return nil, fmt.Errorf("generate user token: %w", err)
	}

	now := s.now()
	if err := s.users.UpdateLogin(ctx, user.ID, token, now.Add(time.Duration(s.tokenExpiresIn)*time.Second), now); err != nil {
		return nil, fmt.Errorf("update user login: %w", err)
	}

	return &model.MiniProgramLoginResponse{
		Token:     token,
		ExpiresIn: s.tokenExpiresIn,
	}, nil
}

func (c *WeChatCode2SessionAPIClient) CodeToOpenID(ctx context.Context, code string) (string, error) {
	endpoint, err := url.Parse(c.endpoint)
	if err != nil {
		return "", fmt.Errorf("parse jscode2session endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("appid", c.appID)
	query.Set("secret", c.appSecret)
	query.Set("js_code", code)
	query.Set("grant_type", "authorization_code")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create jscode2session request: %w", err)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("call jscode2session: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("jscode2session returned status %d", response.StatusCode)
	}

	var body struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode jscode2session response: %w", err)
	}
	if body.ErrCode != 0 {
		return "", fmt.Errorf("jscode2session returned errcode %d", body.ErrCode)
	}
	if strings.TrimSpace(body.OpenID) == "" {
		return "", errors.New("jscode2session response has no openid")
	}

	return body.OpenID, nil
}

func generateUserToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}
