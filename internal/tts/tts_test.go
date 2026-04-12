package tts

import (
	"bytes"
	"context"
	"io"
	"testing"
)

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

func TestTTSClient_Error(t *testing.T) {
	client := &mockTTSClient{err: io.ErrUnexpectedEOF}
	_, err := client.Speak(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPlayer_IsPlayingDefault(t *testing.T) {
	p := NewPlayer(true)
	if p.IsPlaying() {
		t.Error("player should not be playing when freshly created")
	}
}

func TestPlayer_StopNoop(t *testing.T) {
	p := NewPlayer(true)
	p.Stop() // should not panic when stream is nil
}

func TestIsUnderrun(t *testing.T) {
	if isUnderrun(nil) {
		t.Error("nil error should not be underrun")
	}
	if !isUnderrun(&underflowErr{}) {
		t.Error("underflow error should be detected")
	}
	if isUnderrun(io.EOF) {
		t.Error("EOF should not be underrun")
	}
}

type underflowErr struct{}

func (e *underflowErr) Error() string { return "Output underflow" }
