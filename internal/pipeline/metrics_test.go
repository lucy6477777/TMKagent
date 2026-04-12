package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMetricsLogger_WritesEventLinesAndSummary(t *testing.T) {
	tmpDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	logger, filename, err := newMetricsLogger()
	if err != nil {
		t.Fatalf("newMetricsLogger: %v", err)
	}

	logger.logASR(120, 64, 2400)
	logger.logTranslate(55, 12, 24)
	logger.close()

	data, err := os.ReadFile(filepath.Join(tmpDir, filename))
	if err != nil {
		t.Fatalf("read metrics log: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `"event":"asr"`) {
		t.Fatalf("metrics log missing asr event: %s", content)
	}
	if !strings.Contains(content, `"event":"translate"`) {
		t.Fatalf("metrics log missing translate event: %s", content)
	}
	if !strings.Contains(content, `"event":"session"`) {
		t.Fatalf("metrics log missing session event: %s", content)
	}
	if !strings.Contains(content, `"chunks":1`) {
		t.Fatalf("metrics log missing chunk summary: %s", content)
	}
}

func TestMetricsLogger_SessionAverages(t *testing.T) {
	tmpDir := t.TempDir()
	origWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	_ = os.Chdir(tmpDir)

	logger, filename, err := newMetricsLogger()
	if err != nil {
		t.Fatalf("newMetricsLogger: %v", err)
	}

	logger.logASR(100, 10, 1000)
	logger.logASR(300, 30, 3000)
	logger.logTranslate(200, 100, 200)
	logger.logTranslate(400, 200, 300)
	logger.close()

	data, _ := os.ReadFile(filepath.Join(tmpDir, filename))
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	lastLine := lines[len(lines)-1]

	// asr_avg_ms = (100+300)/2 = 200, trans_avg_ms = (200+400)/2 = 300
	if !strings.Contains(lastLine, `"asr_avg_ms":200`) {
		t.Errorf("expected asr_avg_ms:200 in session line: %s", lastLine)
	}
	if !strings.Contains(lastLine, `"trans_avg_ms":300`) {
		t.Errorf("expected trans_avg_ms:300 in session line: %s", lastLine)
	}
}

func TestMetricsLogger_EmptySession(t *testing.T) {
	tmpDir := t.TempDir()
	origWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	_ = os.Chdir(tmpDir)

	logger, filename, err := newMetricsLogger()
	if err != nil {
		t.Fatalf("newMetricsLogger: %v", err)
	}
	logger.close()

	data, _ := os.ReadFile(filepath.Join(tmpDir, filename))
	content := string(data)
	if !strings.Contains(content, `"chunks":0`) {
		t.Errorf("empty session should have chunks:0: %s", content)
	}
	if !strings.Contains(content, `"asr_avg_ms":0`) {
		t.Errorf("empty session should have asr_avg_ms:0: %s", content)
	}
}

func TestNow_ReturnsRFC3339Timestamp(t *testing.T) {
	ts := now()
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %q: %v", ts, err)
	}
}
