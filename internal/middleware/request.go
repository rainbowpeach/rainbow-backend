package middleware

import (
	"net/http"
	"strings"
)

func RequestHost(r *http.Request) string {
	host := ForwardedHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}

	return host
}

func RequestBaseURL(r *http.Request) string {
	scheme := ForwardedHeaderValue(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	return scheme + "://" + RequestHost(r)
}

func ForwardedHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if idx := strings.Index(value, ","); idx >= 0 {
		value = value[:idx]
	}

	return strings.TrimSpace(value)
}
