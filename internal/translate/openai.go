package translate

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Client is the interface for text translation.
// Returns (normalizedSource, translation, error).
// normalizedSource is non-empty only when the source was normalized (e.g. trad→simp for zh).
type Client interface {
	Translate(ctx context.Context, text, fromLang, toLang string) (string, string, error)
}

// OpenAIClient implements Client using GPT-4o-mini.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient creates an OpenAIClient.
func NewOpenAIClient(apiKey, baseURL string) *OpenAIClient {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" && baseURL != "https://api.openai.com/v1" {
		cfg.BaseURL = baseURL
	}
	return &OpenAIClient{client: openai.NewClientWithConfig(cfg)}
}

// Translate translates text from fromLang to toLang.
// For zh source: returns (simplifiedSource, translation, error) — GPT normalises trad→simp in one call.
// For other languages: returns ("", translation, error).
func (o *OpenAIClient) Translate(ctx context.Context, text, fromLang, toLang string) (string, string, error) {
	var systemPrompt string
	twoLines := fromLang == "zh"

	if twoLines {
		systemPrompt = fmt.Sprintf(
			"You are a professional simultaneous interpreter.\n"+
				"Output exactly two lines with no labels:\n"+
				"Line 1: the input converted to Simplified Chinese (if already simplified, copy as-is)\n"+
				"Line 2: the translation into %s",
			toLang,
		)
	} else {
		systemPrompt = fmt.Sprintf(
			"You are a professional simultaneous interpreter. "+
				"Translate the following %s text into %s. "+
				"Output only the translation, no explanation.",
			fromLang, toLang,
		)
	}

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: text},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("translation: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", "", fmt.Errorf("translation: empty response")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)

	if twoLines {
		parts := strings.SplitN(content, "\n", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
		}
		// GPT didn't follow two-line format — fall back: use original text as source
		return "", content, nil
	}

	return "", content, nil
}
