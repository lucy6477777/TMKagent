package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockASR struct {
	response string
	err      error
	calls    int
}

func (m *mockASR) Transcribe(_ context.Context, _ []byte, _ string, _ string) (string, error) {
	m.calls++
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

func TestTranscribeChunked_SingleChunk(t *testing.T) {
	mock := &mockASR{response: "hello world"}
	data := make([]byte, 1000)

	result, err := transcribeChunked(context.Background(), data, "audio.wav", "en", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("got %q, want %q", result, "hello world")
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 call, got %d", mock.calls)
	}
}

func TestTranscribeChunked_MultipleChunks(t *testing.T) {
	multiMock := &multiChunkASR{responses: []string{"chunk one", "chunk two", "chunk three"}}

	data := make([]byte, maxChunkBytes*2+100)

	result, err := transcribeChunked(context.Background(), data, "audio.wav", "en", multiMock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if multiMock.calls != 3 {
		t.Errorf("expected 3 ASR calls for %d bytes, got %d", len(data), multiMock.calls)
	}
	if result != "chunk one chunk two chunk three" {
		t.Errorf("got %q, want concatenated chunks", result)
	}
}

type multiChunkASR struct {
	responses []string
	calls     int
}

func (m *multiChunkASR) Transcribe(_ context.Context, _ []byte, _ string, _ string) (string, error) {
	idx := m.calls
	m.calls++
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return "", nil
}

func TestTranscribeChunked_ErrorOnChunk(t *testing.T) {
	errMock := &mockASR{err: fmt.Errorf("API failure")}
	data := make([]byte, maxChunkBytes*2+1)

	_, err := transcribeChunked(context.Background(), data, "audio.wav", "en", errMock)
	if err == nil {
		t.Fatal("expected error on chunk failure")
	}
	if !strings.Contains(err.Error(), "chunk 0") {
		t.Errorf("error should mention chunk number, got: %v", err)
	}
}
