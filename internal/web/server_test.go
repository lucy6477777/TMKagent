package web

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test", Port: 8080})
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestHandler_ServesRoot(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test"})
	handler := s.Handler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct == "" {
		t.Error("expected Content-Type header")
	}
}

func TestHandleUpload_Success(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test"})
	handler := s.Handler()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.wav")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("fake-wav-data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("POST /upload status = %d, want 204", rec.Code)
	}

	s.mu.Lock()
	uploadPath := s.lastUpload
	s.mu.Unlock()

	if uploadPath == "" {
		t.Fatal("lastUpload should be set after upload")
	}
	data, err := os.ReadFile(uploadPath)
	if err != nil {
		t.Fatalf("reading uploaded file: %v", err)
	}
	if string(data) != "fake-wav-data" {
		t.Errorf("uploaded content = %q, want %q", string(data), "fake-wav-data")
	}
	os.Remove(uploadPath)
}

func TestHandleUpload_MethodNotAllowed(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test"})
	handler := s.Handler()

	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /upload status = %d, want 405", rec.Code)
	}
}

func TestHandleUpload_MissingFile(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test"})
	handler := s.Handler()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /upload without file status = %d, want 400", rec.Code)
	}
}

func TestStartAndCancelPipeline(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test"})

	started := make(chan struct{})
	s.startPipeline(func(ctx context.Context) {
		close(started)
		<-ctx.Done()
	})
	<-started

	s.mu.Lock()
	s.cancelPipeline()
	s.mu.Unlock()

	if s.cancel != nil {
		t.Error("cancel should be nil after cancelPipeline")
	}
}

func TestCancelPipeline_Noop(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test"})
	s.mu.Lock()
	s.cancelPipeline() // should not panic when cancel is nil
	s.mu.Unlock()
}

func TestHandleUpload_PreservesExtension(t *testing.T) {
	s := NewServer(ServerConfig{APIKey: "sk-test"})
	handler := s.Handler()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "recording.m4a")
	io.WriteString(part, "fake-m4a")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}

	s.mu.Lock()
	path := s.lastUpload
	s.mu.Unlock()

	if path == "" {
		t.Fatal("no upload recorded")
	}
	defer os.Remove(path)

	if len(path) < 4 || path[len(path)-4:] != ".m4a" {
		t.Errorf("uploaded file path %q should end with .m4a", path)
	}
}
