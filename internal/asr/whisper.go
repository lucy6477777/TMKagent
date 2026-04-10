package asr

import (
	"bytes"
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// Client is the interface for speech-to-text transcription.
type Client interface {
	Transcribe(ctx context.Context, audioBytes []byte, filename string, lang string) (string, error)
}

// WhisperClient implements Client using OpenAI Whisper-1.
type WhisperClient struct {
	client *openai.Client
}

// NewWhisperClient creates a WhisperClient.
// apiKey: OpenAI API key. baseURL: API endpoint (empty = default OpenAI URL).
func NewWhisperClient(apiKey, baseURL string) *WhisperClient {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" && baseURL != "https://api.openai.com/v1" {
		cfg.BaseURL = baseURL
	}
	return &WhisperClient{client: openai.NewClientWithConfig(cfg)}
}

// Transcribe sends audioBytes to Whisper-1 and returns the recognised text.
// filename tells the API the audio format (e.g. "audio.wav", "audio.mp3").
// lang is the BCP-47 source language code; empty string = auto-detect.
func (w *WhisperClient) Transcribe(ctx context.Context, audioBytes []byte, filename string, lang string) (string, error) {
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		Reader:   bytes.NewReader(audioBytes),
		FilePath: filename,
	}
	if lang != "" {
		req.Language = lang
	}
	resp, err := w.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("whisper: %w", err)
	}
	return resp.Text, nil
}
