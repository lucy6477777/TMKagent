package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lucyliuu/mini-tmk-agent/internal/asr"
	"github.com/lucyliuu/mini-tmk-agent/internal/audio"
)

const maxChunkBytes = 20 * 1024 * 1024 // 20MB — below Whisper's 25MB limit

// TranscriptConfig holds parameters for the transcript command.
type TranscriptConfig struct {
	FilePath   string
	OutputPath string
	SourceLang string // BCP-47 code; empty = auto-detect
}

// RunTranscript reads an audio file, transcribes it via Whisper, and writes
// the plain-text result to OutputPath.
func RunTranscript(ctx context.Context, cfg TranscriptConfig, asrClient asr.Client) error {
	audioBytes, filename, err := audio.ReadAudioFile(cfg.FilePath)
	if err != nil {
		return err
	}

	transcription, err := transcribeChunked(ctx, audioBytes, filename, cfg.SourceLang, asrClient)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cfg.OutputPath, []byte(transcription+"\n"), 0644); err != nil {
		return fmt.Errorf("writing output to %s: %w", cfg.OutputPath, err)
	}
	fmt.Printf("Transcription written to %s\n", cfg.OutputPath)
	return nil
}

// transcribeChunked splits large files into ≤20MB chunks and concatenates results.
func transcribeChunked(ctx context.Context, data []byte, filename, lang string, c asr.Client) (string, error) {
	if len(data) <= maxChunkBytes {
		return c.Transcribe(ctx, data, filename, lang)
	}

	var parts []string
	for i := 0; i < len(data); i += maxChunkBytes {
		end := i + maxChunkBytes
		if end > len(data) {
			end = len(data)
		}
		text, err := c.Transcribe(ctx, data[i:end], filename, lang)
		if err != nil {
			return "", fmt.Errorf("chunk %d: %w", i/maxChunkBytes, err)
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, " "), nil
}
