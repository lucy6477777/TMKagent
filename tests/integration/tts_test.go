//go:build integration

package integration_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/tts"
)

func TestTTSIntegration_Speak_ReturnsPCMAudio(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set; skipping integration test")
	}

	client := tts.NewOpenAIClient(apiKey, "", "", "pcm")
	rc, err := client.Speak(context.Background(), "Hello.")
	if err != nil {
		t.Fatalf("Speak failed: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty audio bytes")
	}
	// PCM 24kHz 16-bit mono: even byte count, reasonable size for a short utterance
	if len(data)%2 != 0 {
		t.Errorf("PCM data length %d is not even (expected 16-bit samples)", len(data))
	}
	t.Logf("Speak returned %d PCM bytes (~%.2f seconds at 24kHz)", len(data), float64(len(data))/2/24000)
}

func TestTTSIntegration_Speak_ChineseText(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set; skipping integration test")
	}

	client := tts.NewOpenAIClient(apiKey, "", "", "pcm")
	rc, err := client.Speak(context.Background(), "你好世界")
	if err != nil {
		t.Fatalf("Speak failed: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty audio bytes for Chinese text")
	}
	t.Logf("Chinese TTS returned %d PCM bytes", len(data))
}
