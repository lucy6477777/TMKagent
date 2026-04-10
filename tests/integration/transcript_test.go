//go:build integration

package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/asr"
	"github.com/lucyliuu/mini-tmk-agent/internal/pipeline"
)

func TestTranscriptIntegration_RealWhisper(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set; skipping integration test")
	}

	outFile := filepath.Join(t.TempDir(), "result.txt")
	asrClient := asr.NewWhisperClient(apiKey, "")

	err := pipeline.RunTranscript(context.Background(), pipeline.TranscriptConfig{
		FilePath:   "../../testdata/hello_zh.wav",
		OutputPath: outFile,
		SourceLang: "",
	}, asrClient)

	// The testdata file is silent, so Whisper may return empty string — that's fine.
	// We just verify the pipeline completes without error and creates the output file.
	if err != nil {
		t.Fatalf("RunTranscript failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	t.Logf("Transcription result: %q", strings.TrimSpace(string(data)))
}
