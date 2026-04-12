package pipeline

import (
	"context"
	"errors"
	"testing"
)

type countingASRClient struct {
	responses []string
	err       error
	calls     int
	sizes     []int
}

func (c *countingASRClient) Transcribe(_ context.Context, audioBytes []byte, _ string, _ string) (string, error) {
	c.calls++
	c.sizes = append(c.sizes, len(audioBytes))
	if c.err != nil {
		return "", c.err
	}
	if c.calls <= len(c.responses) {
		return c.responses[c.calls-1], nil
	}
	return "", nil
}

func TestTranscribeChunked_UsesSingleRequestForSmallFiles(t *testing.T) {
	client := &countingASRClient{responses: []string{"hello"}}

	text, err := transcribeChunked(context.Background(), make([]byte, 128), "audio.wav", "zh", client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "hello" {
		t.Fatalf("got text %q, want %q", text, "hello")
	}
	if client.calls != 1 {
		t.Fatalf("got %d calls, want 1", client.calls)
	}
}

func TestTranscribeChunked_SplitsLargeFiles(t *testing.T) {
	client := &countingASRClient{responses: []string{"first", "second"}}
	audioBytes := make([]byte, maxChunkBytes+17)

	text, err := transcribeChunked(context.Background(), audioBytes, "audio.wav", "zh", client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "first second" {
		t.Fatalf("got text %q, want %q", text, "first second")
	}
	if client.calls != 2 {
		t.Fatalf("got %d calls, want 2", client.calls)
	}
	if client.sizes[0] != maxChunkBytes {
		t.Fatalf("got first chunk size %d, want %d", client.sizes[0], maxChunkBytes)
	}
	if client.sizes[1] != 17 {
		t.Fatalf("got second chunk size %d, want %d", client.sizes[1], 17)
	}
}

func TestTranscribeChunked_PropagatesChunkErrors(t *testing.T) {
	client := &countingASRClient{err: errors.New("boom")}
	audioBytes := make([]byte, maxChunkBytes+1)

	_, err := transcribeChunked(context.Background(), audioBytes, "audio.wav", "zh", client)
	if err == nil {
		t.Fatal("expected error")
	}
}
