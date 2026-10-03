package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestHostPrefersHostHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/public/scene-page-config", nil)
	req.Host = "cluo7.dapinsport.cn"
	req.Header.Set("X-Forwarded-Host", "internal.example.com")

	if got := RequestHost(req); got != "cluo7.dapinsport.cn" {
		t.Fatalf("expected Host header to win, got %q", got)
	}
}

func TestRequestHostFallsBackToForwardedHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/public/scene-page-config", nil)
	req.Host = ""
	req.Header.Set("X-Forwarded-Host", "cluo7.dapinsport.cn, proxy.local")

	if got := RequestHost(req); got != "cluo7.dapinsport.cn" {
		t.Fatalf("expected first forwarded host, got %q", got)
	}
}
