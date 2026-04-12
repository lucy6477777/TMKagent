package unit_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/tts"
)

// mockTTSClient implements tts.Client for testing.
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

// Verify mockTTSClient satisfies the interface at compile time.
var _ tts.Client = (*mockTTSClient)(nil)

func TestTTSClient_Interface(t *testing.T) {
	pcm := make([]byte, 4800) // 100ms of silence at 24kHz 16-bit mono
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
	p := tts.NewPlayer(true)
	if p.IsPlaying() {
		t.Error("player should not be playing when freshly created")
	}
}
