package asr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhisperClient_Transcribe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/transcriptions" {
			t.Fatalf("got path %q, want %q", r.URL.Path, "/audio/transcriptions")
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if got := r.FormValue("model"); got != "whisper-1" {
			t.Fatalf("got model %q, want %q", got, "whisper-1")
		}
		if got := r.FormValue("language"); got != "zh" {
			t.Fatalf("got language %q, want %q", got, "zh")
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile: %v", err)
		}
		defer file.Close()

		if header.Filename != "audio.wav" {
			t.Fatalf("got filename %q, want %q", header.Filename, "audio.wav")
		}
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		if string(data) != "audio-bytes" {
			t.Fatalf("got file contents %q, want %q", string(data), "audio-bytes")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "你好"})
	}))
	defer server.Close()

	client := NewWhisperClient("sk-test", server.URL)
	text, err := client.Transcribe(context.Background(), []byte("audio-bytes"), "audio.wav", "zh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "你好" {
		t.Fatalf("got %q, want %q", text, "你好")
	}
}

func TestWhisperClient_TranscribeOmitsLanguageWhenEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if got := r.FormValue("language"); got != "" {
			t.Fatalf("expected empty language field, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "hello"})
	}))
	defer server.Close()

	client := NewWhisperClient("sk-test", server.URL)
	text, err := client.Transcribe(context.Background(), []byte("audio-bytes"), "audio.mp3", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "hello" {
		t.Fatalf("got %q, want %q", text, "hello")
	}
}
