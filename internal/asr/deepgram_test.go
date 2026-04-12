package asr

import (
	"context"
	"testing"
)

type mockStreamSession struct {
	results chan TranscriptResult
	closed  bool
}

func newMockStreamSession() *mockStreamSession {
	return &mockStreamSession{results: make(chan TranscriptResult, 16)}
}

func (m *mockStreamSession) Send(_ []byte) error            { return nil }
func (m *mockStreamSession) Results() <-chan TranscriptResult { return m.results }
func (m *mockStreamSession) Close() error                    { m.closed = true; close(m.results); return nil }

type mockStreamClient struct {
	session *mockStreamSession
}

func (m *mockStreamClient) Connect(_ context.Context, _ string) (StreamSession, error) {
	return m.session, nil
}

var _ StreamClient = (*mockStreamClient)(nil)
var _ StreamSession = (*mockStreamSession)(nil)

func TestStreamSession_InterimAndFinal(t *testing.T) {
	sess := newMockStreamSession()

	sess.results <- TranscriptResult{Text: "hello", IsFinal: false}
	sess.results <- TranscriptResult{Text: "hello world", IsFinal: true}

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
