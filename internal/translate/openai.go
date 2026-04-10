package translate

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// Client is the interface for text translation.
type Client interface {
	Translate(ctx context.Context, text, fromLang, toLang string) (string, error)
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

// Translate translates text from fromLang to toLang using GPT-4o-mini.
// fromLang and toLang are language codes (e.g. "zh", "en", "es", "ja").
func (o *OpenAIClient) Translate(ctx context.Context, text, fromLang, toLang string) (string, error) {
	systemPrompt := fmt.Sprintf(
		"You are a professional simultaneous interpreter. "+
			"Translate the following %s text into %s. "+
			"Output only the translation, no explanation.",
		fromLang, toLang,
	)
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: text},
		},
	})
	if err != nil {
		return "", fmt.Errorf("translation: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("translation: empty response")
	}
	return resp.Choices[0].Message.Content, nil
}
