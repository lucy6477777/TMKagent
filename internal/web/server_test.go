package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicAccessInfo_UsesConfiguredBaseURL(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:          8080,
		PublicBaseURL: "https://demo.example:9443/app/",
	})

	baseURL, ip, port := server.publicAccessInfo()

	if baseURL != "https://demo.example:9443/app" {
		t.Fatalf("got baseURL %q, want %q", baseURL, "https://demo.example:9443/app")
	}
	if ip != "demo.example" {
		t.Fatalf("got ip %q, want %q", ip, "demo.example")
	}
	if port != 9443 {
		t.Fatalf("got port %d, want %d", port, 9443)
	}
}

func TestPublicAccessInfo_InvalidConfiguredBaseURLFallsBack(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:          8080,
		PublicBaseURL: "://bad-url",
	})

	baseURL, ip, port := server.publicAccessInfo()

	if !strings.HasPrefix(baseURL, "http://") {
		t.Fatalf("expected fallback http URL, got %q", baseURL)
	}
	if ip == "" {
		t.Fatal("expected fallback IP to be non-empty")
	}
	if port != 8080 {
		t.Fatalf("got port %d, want %d", port, 8080)
	}
}

func TestHandleInfo_ReturnsJSON(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:          8080,
		PublicBaseURL: "https://demo.example:9443/app/",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	rec := httptest.NewRecorder()

	server.handleInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["baseURL"] != "https://demo.example:9443/app" {
		t.Fatalf("got baseURL %v, want %q", payload["baseURL"], "https://demo.example:9443/app")
	}
}
