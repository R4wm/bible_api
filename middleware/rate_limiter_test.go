package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIPPrefersForwardedFor(t *testing.T) {
	rl := &RateLimiter{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.Header.Set("X-Real-IP", "198.51.100.5")
	req.RemoteAddr = "192.0.2.1:1234"

	ip := rl.getClientIP(req)
	if ip != "203.0.113.10" {
		t.Fatalf("expected forwarded IP, got %s", ip)
	}
}

func TestGetClientIPFallsBackToRealIP(t *testing.T) {
	rl := &RateLimiter{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("X-Real-IP", "198.51.100.5")
	req.RemoteAddr = "192.0.2.1:1234"

	ip := rl.getClientIP(req)
	if ip != "198.51.100.5" {
		t.Fatalf("expected real IP, got %s", ip)
	}
}

func TestGetClientIPFallsBackToRemoteAddr(t *testing.T) {
	rl := &RateLimiter{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	ip := rl.getClientIP(req)
	if ip != "192.0.2.1" {
		t.Fatalf("expected remote IP, got %s", ip)
	}
}
