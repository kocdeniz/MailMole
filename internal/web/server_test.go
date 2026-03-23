package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeWebAddr(t *testing.T) {
	addr, host := normalizeWebAddr(":8080")
	if addr != "127.0.0.1:8080" {
		t.Fatalf("expected normalized addr, got %q", addr)
	}
	if host != "127.0.0.1" {
		t.Fatalf("expected host 127.0.0.1, got %q", host)
	}

	addr, host = normalizeWebAddr("localhost:9090")
	if addr != "localhost:9090" {
		t.Fatalf("expected addr localhost:9090, got %q", addr)
	}
	if host != "localhost" {
		t.Fatalf("expected host localhost, got %q", host)
	}

	addr, host = normalizeWebAddr("0.0.0.0:8080")
	if addr != "0.0.0.0:8080" {
		t.Fatalf("expected addr 0.0.0.0:8080, got %q", addr)
	}
	if host != "0.0.0.0" {
		t.Fatalf("expected host 0.0.0.0, got %q", host)
	}
}

func TestIsLocalhost(t *testing.T) {
	cases := map[string]bool{
		"":          true,
		"localhost": true,
		"127.0.0.1": true,
		"::1":       true,
		"0.0.0.0":   false,
		"example":   false,
	}

	for host, want := range cases {
		if got := isLocalhost(host); got != want {
			t.Fatalf("isLocalhost(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestAuthorized(t *testing.T) {
	s := &Server{requireAuth: true, authToken: "secret"}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/status", nil)
	if s.authorized(req) {
		t.Fatalf("expected unauthorized without token")
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com/api/status", nil)
	req.Header.Set("Authorization", "Bearer secret")
	if !s.authorized(req) {
		t.Fatalf("expected authorized with bearer token")
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com/api/status", nil)
	req.Header.Set("X-MailMole-Token", "secret")
	if !s.authorized(req) {
		t.Fatalf("expected authorized with X-MailMole-Token")
	}

	req = httptest.NewRequest(http.MethodGet, "http://example.com/api/status?token=secret", nil)
	if !s.authorized(req) {
		t.Fatalf("expected authorized with query token")
	}
}

func TestAuthRequired(t *testing.T) {
	s := &Server{requireAuth: true, authToken: "secret"}
	handler := s.authRequired(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/status", nil)
	handler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "http://example.com/api/status", nil)
	req.Header.Set("Authorization", "Bearer secret")
	handler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestDecodeJSONLimit(t *testing.T) {
	payload := `{"a":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/test", bytes.NewBufferString(payload))
	rr := httptest.NewRecorder()

	var out map[string]interface{}
	if err := decodeJSON(rr, req, &out); err == nil {
		t.Fatalf("expected error for oversized payload")
	}
}
