package pipeline

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestMetricsLogger_LogASRAndTranslate(t *testing.T) {
	ml, logFile, err := newMetricsLogger()
	if err != nil {
		t.Fatalf("newMetricsLogger: %v", err)
	}
	defer os.Remove(logFile)

	ml.logASR(150, 32, 5000)
	ml.logASR(200, 48, 8000)
	ml.logTranslate(80, 50, 120)
	ml.close()

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 log lines (2 asr + 1 translate + 1 session), got %d", len(lines))
	}

	// Verify ASR events
	var asrEvent map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &asrEvent); err != nil {
		t.Fatalf("unmarshal asr event: %v", err)
	}
	if asrEvent["event"] != "asr" {
		t.Errorf("first event type = %v, want asr", asrEvent["event"])
	}

	// Verify translate event
	var transEvent map[string]any
	if err := json.Unmarshal([]byte(lines[2]), &transEvent); err != nil {
		t.Fatalf("unmarshal translate event: %v", err)
	}
	if transEvent["event"] != "translate" {
		t.Errorf("third event type = %v, want translate", transEvent["event"])
	}

	// Verify session summary
	var session map[string]any
	if err := json.Unmarshal([]byte(lines[3]), &session); err != nil {
		t.Fatalf("unmarshal session: %v", err)
	}
	if session["event"] != "session" {
		t.Errorf("session event type = %v, want session", session["event"])
	}
	if session["chunks"].(float64) != 2 {
		t.Errorf("chunks = %v, want 2", session["chunks"])
	}
	if session["est_cost_usd"].(float64) <= 0 {
		t.Errorf("est_cost_usd should be > 0, got %v", session["est_cost_usd"])
	}
}

func TestMetricsLogger_SessionAverages(t *testing.T) {
	ml, logFile, err := newMetricsLogger()
	if err != nil {
		t.Fatalf("newMetricsLogger: %v", err)
	}
	defer os.Remove(logFile)

	ml.logASR(100, 10, 1000)
	ml.logASR(300, 30, 3000)
	ml.logTranslate(200, 100, 200)
	ml.logTranslate(400, 200, 300)
	ml.close()

	data, _ := os.ReadFile(logFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	lastLine := lines[len(lines)-1]

	var session map[string]any
	json.Unmarshal([]byte(lastLine), &session)

	// asr_avg_ms = (100+300)/2 = 200
	if session["asr_avg_ms"].(float64) != 200 {
		t.Errorf("asr_avg_ms = %v, want 200", session["asr_avg_ms"])
	}
	// trans_avg_ms = (200+400)/2 = 300
	if session["trans_avg_ms"].(float64) != 300 {
		t.Errorf("trans_avg_ms = %v, want 300", session["trans_avg_ms"])
	}
}

func TestMetricsLogger_EmptySession(t *testing.T) {
	ml, logFile, err := newMetricsLogger()
	if err != nil {
		t.Fatalf("newMetricsLogger: %v", err)
	}
	defer os.Remove(logFile)

	ml.close()

	data, _ := os.ReadFile(logFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 session line, got %d", len(lines))
	}

	var session map[string]any
	json.Unmarshal([]byte(lines[0]), &session)
	if session["chunks"].(float64) != 0 {
		t.Errorf("chunks = %v, want 0", session["chunks"])
	}
	if session["asr_avg_ms"].(float64) != 0 {
		t.Errorf("asr_avg_ms should be 0 with no chunks")
	}
}
