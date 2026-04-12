package tts

import (
	"context"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestNewOpenAIClient_Defaults(t *testing.T) {
	client := NewOpenAIClient("sk-test", "", "", "")

	if client.voice != openai.VoiceAlloy {
		t.Fatalf("got voice %q, want %q", client.voice, openai.VoiceAlloy)
	}
	if client.format != openai.SpeechResponseFormatPcm {
		t.Fatalf("got format %q, want %q", client.format, openai.SpeechResponseFormatPcm)
	}
}

func TestNewOpenAIClient_UsesCustomVoiceAndFormat(t *testing.T) {
	client := NewOpenAIClient("sk-test", "", "nova", "mp3")

	if client.voice != openai.SpeechVoice("nova") {
		t.Fatalf("got voice %q, want %q", client.voice, "nova")
	}
	if client.format != openai.SpeechResponseFormat("mp3") {
		t.Fatalf("got format %q, want %q", client.format, "mp3")
	}
}

func TestIsUnderrun(t *testing.T) {
	if !isUnderrun(errors.New("portaudio output underflow")) {
		t.Fatal("expected underflow error to be treated as underrun")
	}
	if isUnderrun(errors.New("some other error")) {
		t.Fatal("unexpected underrun match for unrelated error")
	}
	if isUnderrun(nil) {
		t.Fatal("nil error should not be treated as underrun")
	}
}

func TestOpenAIClient_Speak(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/speech" {
			t.Fatalf("got path %q, want %q", r.URL.Path, "/audio/speech")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pcm-data"))
	}))
	defer server.Close()

	client := NewOpenAIClient("sk-test", server.URL, "nova", "mp3")
	rc, err := client.Speak(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "pcm-data" {
		t.Fatalf("got data %q, want %q", string(data), "pcm-data")
	}
}

type mockTTSClient struct {
	audio []byte
	err   error
}

func (m *mockTTSClient) Speak(_ context.Context, _ string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(bytes.NewReader(m.audio)), nil
}

var _ Client = (*mockTTSClient)(nil)

func TestMockTTSClient_Speak(t *testing.T) {
	pcm := make([]byte, 4800)
	client := &mockTTSClient{audio: pcm}

	rc, err := client.Speak(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if len(data) != len(pcm) {
		t.Errorf("got %d bytes, want %d", len(data), len(pcm))
	}
}

func TestTTSPlayer_IsPlayingDefault(t *testing.T) {
	player := NewPlayer(true)
	if player.IsPlaying() {
		t.Error("player should not be playing when freshly created")
	}
}

func TestTTSPlayer_StopNoop(t *testing.T) {
	player := NewPlayer(true)
	player.Stop() // should not panic when stream is nil
}

func TestMockTTSClient_Error(t *testing.T) {
	client := &mockTTSClient{err: io.ErrUnexpectedEOF}
	_, err := client.Speak(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}
