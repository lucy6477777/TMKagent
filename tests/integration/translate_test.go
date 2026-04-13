//go:build integration

package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/translate"
)

func TestTranslateIntegration_EnToZh(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set; skipping integration test")
	}

	client := translate.NewOpenAIClient(apiKey, "")
	_, result, err := client.Translate(context.Background(), "Hello, how are you?", "en", "zh")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if strings.TrimSpace(result) == "" {
		t.Fatal("expected non-empty translation")
	}
	t.Logf("en→zh: %q", result)
}

func TestTranslateIntegration_ZhToEn_ReturnsTwoFields(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set; skipping integration test")
	}

	client := translate.NewOpenAIClient(apiKey, "")
	// Traditional Chinese input — expect normalized simplified + English translation
	normalized, translation, err := client.Translate(context.Background(), "你好，今天天气怎么样？", "zh", "en")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if strings.TrimSpace(translation) == "" {
		t.Fatal("expected non-empty translation")
	}
	t.Logf("zh→en normalized: %q, translation: %q", normalized, translation)
}

func TestTranslateIntegration_SameLangPassthrough(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set; skipping integration test")
	}

	client := translate.NewOpenAIClient(apiKey, "")
	_, result, err := client.Translate(context.Background(), "Hello world", "en", "es")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if strings.TrimSpace(result) == "" {
		t.Fatal("expected non-empty translation")
	}
	t.Logf("en→es: %q", result)
}
