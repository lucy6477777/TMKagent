package translate

import (
	"context"
	"testing"
)

type mockClient struct {
	normalizedSrc string
	translation   string
	err           error
}

func (m *mockClient) Translate(_ context.Context, _, _, _ string) (string, string, error) {
	return m.normalizedSrc, m.translation, m.err
}

var _ Client = (*mockClient)(nil)

func TestClientInterface(t *testing.T) {
	mock := &mockClient{normalizedSrc: "你好", translation: "Hello"}
	src, tgt, err := mock.Translate(context.Background(), "你好", "zh", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src != "你好" {
		t.Errorf("normalizedSrc = %q, want %q", src, "你好")
	}
	if tgt != "Hello" {
		t.Errorf("translation = %q, want %q", tgt, "Hello")
	}
}

func TestNewOpenAIClient(t *testing.T) {
	c := NewOpenAIClient("sk-test", "")
	if c == nil {
		t.Fatal("NewOpenAIClient returned nil")
	}
}

func TestNewOpenAIClient_CustomBaseURL(t *testing.T) {
	c := NewOpenAIClient("sk-test", "https://custom.api/v1")
	if c == nil {
		t.Fatal("NewOpenAIClient returned nil")
	}
}
