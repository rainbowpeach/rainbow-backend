package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
)

func TestArkChatClientReadsTextReply(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("authorization = %s", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/v3/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"收到"}}]}`))
	}))
	defer server.Close()

	client := NewArkChatClient(config.ArkConfig{
		BaseURL: server.URL + "/api/v3",
		APIKey:  "test-key",
		Model:   "ep-test",
	}, server.Client())

	reply, err := client.Complete(context.Background(), []model.ChatTurn{{Role: "user", Content: "你好"}}, 0.8, 400)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if reply != "收到" {
		t.Fatalf("reply = %q", reply)
	}
}

func TestArkChatClientRequiresConfig(t *testing.T) {
	client := NewArkChatClient(config.ArkConfig{}, nil)
	_, err := client.Complete(context.Background(), nil, 0.8, 400)
	if !errors.Is(err, ErrChatNotConfigured) {
		t.Fatalf("Complete() error = %v", err)
	}
}
