package translate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildSystemPrompt_ZhUsesTwoLineFormat(t *testing.T) {
	prompt := buildSystemPrompt("zh", "en")

	if !strings.Contains(prompt, "Output exactly two lines") {
		t.Fatalf("zh prompt should request two-line output, got: %q", prompt)
	}
	if !strings.Contains(prompt, "translation into en") {
		t.Fatalf("zh prompt should include target language, got: %q", prompt)
	}
}

func TestBuildSystemPrompt_NonZhUsesTranslationOnlyFormat(t *testing.T) {
	prompt := buildSystemPrompt("es", "ja")

	if !strings.Contains(prompt, "Translate the following es text into ja") {
		t.Fatalf("non-zh prompt should include source and target language, got: %q", prompt)
	}
	if !strings.Contains(prompt, "Output only the translation") {
		t.Fatalf("non-zh prompt should request translation only, got: %q", prompt)
	}
}

func TestParseTranslationContent_ZhTwoLines(t *testing.T) {
	source, translation := parseTranslationContent("繁体中文\nEnglish translation", "zh")
	if source != "繁体中文" {
		t.Fatalf("got source %q, want %q", source, "繁体中文")
	}
	if translation != "English translation" {
		t.Fatalf("got translation %q, want %q", translation, "English translation")
	}
}

func TestParseTranslationContent_ZhFallbacksToSingleLine(t *testing.T) {
	source, translation := parseTranslationContent("single line response", "zh")
	if source != "" {
		t.Fatalf("got source %q, want empty", source)
	}
	if translation != "single line response" {
		t.Fatalf("got translation %q, want %q", translation, "single line response")
	}
}

func TestParseTranslationContent_NonZhTrimsAndReturnsTranslation(t *testing.T) {
	source, translation := parseTranslationContent("  hola  ", "es")
	if source != "" {
		t.Fatalf("got source %q, want empty", source)
	}
	if translation != "hola" {
		t.Fatalf("got translation %q, want %q", translation, "hola")
	}
}

func TestOpenAIClient_TranslateZh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("got path %q, want %q", r.URL.Path, "/chat/completions")
		}

		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode: %v", err)
		}
		if req.Model != "gpt-4o-mini" {
			t.Fatalf("got model %q, want %q", req.Model, "gpt-4o-mini")
		}
		if len(req.Messages) != 2 {
			t.Fatalf("got %d messages, want 2", len(req.Messages))
		}
		if !strings.Contains(req.Messages[0].Content, "Output exactly two lines") {
			t.Fatalf("unexpected system prompt: %q", req.Messages[0].Content)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "简体中文\nEnglish translation",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewOpenAIClient("sk-test", server.URL)
	source, translation, err := client.Translate(context.Background(), "繁體中文", "zh", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source != "简体中文" {
		t.Fatalf("got source %q, want %q", source, "简体中文")
	}
	if translation != "English translation" {
		t.Fatalf("got translation %q, want %q", translation, "English translation")
	}
}

func TestOpenAIClient_TranslateReturnsErrorOnEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{},
		})
	}))
	defer server.Close()

	client := NewOpenAIClient("sk-test", server.URL)
	_, _, err := client.Translate(context.Background(), "hola", "es", "en")
	if err == nil {
		t.Fatal("expected error")
	}
}
