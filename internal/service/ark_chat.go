package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
)

var ErrChatUpstream = errors.New("chat upstream failed")

type ChatCompletionClient interface {
	Complete(ctx context.Context, messages []model.ChatTurn, temperature float64, maxTokens int) (string, error)
}

type ArkChatClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewArkChatClient(cfg config.ArkConfig, client *http.Client) *ArkChatClient {
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &ArkChatClient{
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		apiKey:  strings.TrimSpace(cfg.APIKey),
		model:   strings.TrimSpace(cfg.Model),
		client:  client,
	}
}

type arkChatRequest struct {
	Model       string           `json:"model"`
	Messages    []model.ChatTurn `json:"messages"`
	Temperature float64          `json:"temperature"`
	MaxTokens   int              `json:"max_tokens"`
}

type arkChatResponse struct {
	Choices []struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (c *ArkChatClient) Complete(ctx context.Context, messages []model.ChatTurn, temperature float64, maxTokens int) (string, error) {
	if c == nil || c.apiKey == "" || c.model == "" || c.baseURL == "" {
		return "", ErrChatNotConfigured
	}

	payload, err := json.Marshal(arkChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("marshal ark request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build ark request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrChatUpstream, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("%w: read body: %s", ErrChatUpstream, err.Error())
	}

	var parsed arkChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("%w: invalid json", ErrChatUpstream)
	}
	if resp.StatusCode != http.StatusOK {
		message := ""
		if parsed.Error != nil {
			message = strings.TrimSpace(parsed.Error.Message)
			if message == "" {
				message = strings.TrimSpace(parsed.Error.Code)
			}
		}
		if message == "" {
			message = resp.Status
		}
		return "", fmt.Errorf("%w: %s", ErrChatUpstream, message)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("%w: empty choices", ErrChatUpstream)
	}
	text := pickArkText(parsed.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("%w: empty content", ErrChatUpstream)
	}

	return text, nil
}

func pickArkText(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var parts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var builder strings.Builder
		for _, part := range parts {
			builder.WriteString(part.Text)
		}
		return strings.TrimSpace(builder.String())
	}

	return ""
}
