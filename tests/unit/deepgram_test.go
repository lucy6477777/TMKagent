package unit_test

import (
	"context"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/asr"
)

// mockStreamSession implements asr.StreamSession for testing.
type mockStreamSession struct {
	results chan asr.TranscriptResult
	closed  bool
}

func newMockStreamSession() *mockStreamSession {
	return &mockStreamSession{results: make(chan asr.TranscriptResult, 16)}
}

func (m *mockStreamSession) Send(_ []byte) error   { return nil }
func (m *mockStreamSession) Results() <-chan asr.TranscriptResult { return m.results }
func (m *mockStreamSession) Close() error           { m.closed = true; close(m.results); return nil }

// mockStreamClient implements asr.StreamClient for testing.
type mockStreamClient struct {
	session *mockStreamSession
}

func (m *mockStreamClient) Connect(_ context.Context, _ string) (asr.StreamSession, error) {
	return m.session, nil
}

// Compile-time interface checks.
var _ asr.StreamClient = (*mockStreamClient)(nil)
var _ asr.StreamSession = (*mockStreamSession)(nil)

func TestStreamSession_InterimAndFinal(t *testing.T) {
	sess := newMockStreamSession()

	// Simulate Deepgram sending interim then final result
	sess.results <- asr.TranscriptResult{Text: "hello", IsFinal: false}
	sess.results <- asr.TranscriptResult{Text: "hello world", IsFinal: true}

	r1 := <-sess.Results()
	if r1.IsFinal {
		t.Error("first result should be interim")
	}
	if r1.Text != "hello" {
		t.Errorf("got %q, want %q", r1.Text, "hello")
	}

	r2 := <-sess.Results()
	if !r2.IsFinal {
		t.Error("second result should be final")
	}
	if r2.Text != "hello world" {
		t.Errorf("got %q, want %q", r2.Text, "hello world")
	}
}

func TestStreamClient_Connect(t *testing.T) {
	sess := newMockStreamSession()
	client := &mockStreamClient{session: sess}

	session, err := client.Connect(context.Background(), "zh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("session should not be nil")
	}

	if err := session.Send([]byte{0, 0}); err != nil {
		t.Fatalf("send error: %v", err)
	}

	if err := session.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	if !sess.closed {
		t.Error("session should be closed")
	}
}
