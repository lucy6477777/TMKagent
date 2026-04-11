package tts

import (
	"context"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

// Client generates speech audio from text.
type Client interface {
	// Speak returns a stream of audio bytes (PCM 24kHz 16-bit mono).
	// The caller must close the returned ReadCloser.
	Speak(ctx context.Context, text string) (io.ReadCloser, error)
}

// OpenAIClient implements Client using OpenAI TTS-1.
type OpenAIClient struct {
	client *openai.Client
	voice  openai.SpeechVoice
	format openai.SpeechResponseFormat
}

// NewOpenAIClient creates a TTS client.
// voice: one of alloy, echo, fable, onyx, nova, shimmer (default "alloy").
// format: "pcm" for CLI playback (24kHz 16-bit mono LE), "mp3" for web.
func NewOpenAIClient(apiKey, baseURL, voice, format string) *OpenAIClient {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" && baseURL != "https://api.openai.com/v1" {
		cfg.BaseURL = baseURL
	}
	v := openai.SpeechVoice(voice)
	if voice == "" {
		v = openai.VoiceAlloy
	}
	f := openai.SpeechResponseFormat(format)
	if format == "" {
		f = openai.SpeechResponseFormatPcm
	}
	return &OpenAIClient{
		client: openai.NewClientWithConfig(cfg),
		voice:  v,
		format: f,
	}
}

func (o *OpenAIClient) Speak(ctx context.Context, text string) (io.ReadCloser, error) {
	resp, err := o.client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          text,
		Voice:          o.voice,
		ResponseFormat: o.format,
	})
	if err != nil {
		return nil, fmt.Errorf("tts: %w", err)
	}
	return resp.ReadCloser, nil
}
