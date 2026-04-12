package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockASR struct {
	response string
	err      error
}

func (m *mockASR) Transcribe(_ context.Context, _ []byte, _ string, _ string) (string, error) {
	return m.response, m.err
}

func TestRunTranscript_WritesOutputFile(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "out.txt")

	err := RunTranscript(context.Background(), TranscriptConfig{
		FilePath:   "../../testdata/hello_zh.wav",
		OutputPath: outFile,
		SourceLang: "zh",
	}, &mockASR{response: "你好世界"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !strings.Contains(string(data), "你好世界") {
		t.Errorf("output file missing transcription; got: %q", string(data))
	}
}

func TestRunTranscript_MissingFile(t *testing.T) {
	err := RunTranscript(context.Background(), TranscriptConfig{
		FilePath:   "/nonexistent/file.wav",
		OutputPath: "/tmp/out.txt",
	}, &mockASR{response: "x"})
	if err == nil {
		t.Error("expected error for missing input file")
	}
}

func TestStreamConfig_Fields(t *testing.T) {
	cfg := StreamConfig{
		SourceLang: "zh",
		TargetLang: "en",
	}
	if cfg.SourceLang != "zh" {
		t.Error("source lang not set")
	}
	if cfg.TargetLang != "en" {
		t.Error("target lang not set")
	}
	if cfg.EnableTTS {
		t.Error("TTS should be disabled by default")
	}
}
