//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/web"
)

func TestWebServerIntegration_InfoEndpoint(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port:          8080,
		PublicBaseURL: "http://127.0.0.1:8080",
	})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	resp, err := http.Get(httpServer.URL + "/api/info")
	if err != nil {
		t.Fatalf("GET /api/info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["baseURL"] != "http://127.0.0.1:8080" {
		t.Fatalf("got baseURL %v, want %q", payload["baseURL"], "http://127.0.0.1:8080")
	}
}

func TestWebServerIntegration_UploadEndpoint(t *testing.T) {
	server := web.NewServer(web.ServerConfig{})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "sample.wav")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("wav-data")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/upload", &body)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /upload: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}
