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

func TestNow_ReturnsRFC3339Timestamp(t *testing.T) {
	ts := now()
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %q: %v", ts, err)
	}
}
