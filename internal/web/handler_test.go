package web

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lucyliuu/mini-tmk-agent/internal/translate"
)

type stubTranslator struct {
	normalized string
	translated string
	err        error
}

func (s *stubTranslator) Translate(_ context.Context, _ string, _, _ string) (string, string, error) {
	return s.normalized, s.translated, s.err
}

var _ translate.Client = (*stubTranslator)(nil)

func TestTranslatePair_PassthroughWhenTargetMatchesSource(t *testing.T) {
	src, target, err := translatePair(context.Background(), &stubTranslator{}, "hello", "en", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src != "hello" || target != "hello" {
		t.Fatalf("got (%q, %q), want passthrough", src, target)
	}
}

func TestTranslatePair_UsesNormalizedSource(t *testing.T) {
	src, target, err := translatePair(context.Background(), &stubTranslator{
		normalized: "简体中文",
		translated: "English",
	}, "繁體中文", "zh", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src != "简体中文" {
		t.Fatalf("got src %q, want %q", src, "简体中文")
	}
	if target != "English" {
		t.Fatalf("got target %q, want %q", target, "English")
	}
}

func TestTranslatePair_ReturnsFallbackOnError(t *testing.T) {
	src, target, err := translatePair(context.Background(), &stubTranslator{
		err: os.ErrInvalid,
	}, "hello", "en", "ja")
	if err == nil {
		t.Fatal("expected error")
	}
	if src != "hello" {
		t.Fatalf("got src %q, want original text", src)
	}
	if target != "[翻译失败]" {
		t.Fatalf("got target %q, want failure placeholder", target)
	}
}

func TestHandleUpload_MethodNotAllowed(t *testing.T) {
	server := NewServer(ServerConfig{})
	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	rec := httptest.NewRecorder()

	server.handleUpload(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleUpload_MissingFileField(t *testing.T) {
	server := NewServer(ServerConfig{})
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	server.handleUpload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleUpload_SavesUploadedFile(t *testing.T) {
	server := NewServer(ServerConfig{})
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "sample.wav")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("audio-data")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	server.handleUpload(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusNoContent)
	}
	if server.lastUpload == "" {
		t.Fatal("expected lastUpload to be set")
	}

	data, err := os.ReadFile(server.lastUpload)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(data) != "audio-data" {
		t.Fatalf("got uploaded contents %q, want %q", string(data), "audio-data")
	}
}

func TestStartPipeline_ReplacesPreviousPipelineAndCancelsIt(t *testing.T) {
	server := NewServer(ServerConfig{})
	conn := &websocket.Conn{}
	canceled := make(chan struct{})

	server.startPipeline(conn, "stream", func(ctx context.Context, _ uint64) {
		<-ctx.Done()
		close(canceled)
	})
	server.startPipeline(conn, "transcript", func(context.Context, uint64) {})

	select {
	case <-canceled:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected previous pipeline to be canceled")
	}

	if !server.hasActivePipelineForConn(conn, "transcript") {
		t.Fatal("expected transcript pipeline to be active")
	}

	server.stopPipeline(conn)
	if server.hasActivePipelineForConn(conn, "transcript") {
		t.Fatal("expected no active pipeline after stop")
	}
}
