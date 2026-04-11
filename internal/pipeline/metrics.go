package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// metricsLogger writes per-event JSON lines to a log file.
// Thread-safe: g2 and g3 goroutines call it concurrently.
type metricsLogger struct {
	f        *os.File
	chunks   atomic.Int64
	asrMs    atomic.Int64
	transMs  atomic.Int64
	audioMs  atomic.Int64
	srcChars atomic.Int64
	tgtChars atomic.Int64
}

// newMetricsLogger creates a timestamped log file and returns the logger and filename.
func newMetricsLogger() (*metricsLogger, string, error) {
	name := fmt.Sprintf("mini-tmk-%s.log", time.Now().Format("20060102-150405"))
	f, err := os.Create(name)
	if err != nil {
		return nil, "", err
	}
	return &metricsLogger{f: f}, name, nil
}

// logASR records one Whisper API call.
// audioMs = audio duration derived from WAV bytes: len(wavBytes)*1000/32000 (16kHz mono 16-bit).
func (m *metricsLogger) logASR(latencyMs, chunkKB, audioMs int64) {
	m.chunks.Add(1)
	m.asrMs.Add(latencyMs)
	m.audioMs.Add(audioMs)
	m.write(map[string]any{
		"event": "asr", "ts": now(),
		"latency_ms": latencyMs, "chunk_kb": chunkKB, "audio_ms": audioMs,
	})
}

// logTranslate records one GPT translation call.
func (m *metricsLogger) logTranslate(latencyMs, srcChars, tgtChars int64) {
	m.transMs.Add(latencyMs)
	m.srcChars.Add(srcChars)
	m.tgtChars.Add(tgtChars)
	m.write(map[string]any{
		"event": "translate", "ts": now(),
		"latency_ms": latencyMs, "src_chars": srcChars, "tgt_chars": tgtChars,
	})
}

// close writes the session summary line and closes the file.
func (m *metricsLogger) close() {
	chunks := m.chunks.Load()
	audioSec := float64(m.audioMs.Load()) / 1000.0

	// Cost estimates (as of 2026-04):
	//   Whisper-1:    $0.006 / minute of audio
	//   GPT-4o-mini:  $0.150 / 1M input tokens, $0.600 / 1M output tokens
	//   Token proxy:  Chinese ≈ 1 char/token, English ≈ 4 chars/token → use /3 conservatively
	whisperCost := audioSec / 60.0 * 0.006
	inputTokens := float64(m.srcChars.Load()) / 3.0
	outputTokens := float64(m.tgtChars.Load()) / 4.0
	gptCost := inputTokens*0.150/1e6 + outputTokens*0.600/1e6

	var asrAvg, transAvg int64
	if chunks > 0 {
		asrAvg = m.asrMs.Load() / chunks
		transAvg = m.transMs.Load() / chunks
	}

	m.write(map[string]any{
		"event": "session", "ts": now(),
		"chunks":       chunks,
		"asr_avg_ms":   asrAvg,
		"trans_avg_ms": transAvg,
		"audio_sec":    audioSec,
		"est_cost_usd": whisperCost + gptCost,
	})
	m.f.Close()
}

func (m *metricsLogger) write(v any) {
	b, _ := json.Marshal(v)
	m.f.Write(append(b, '\n')) //nolint:errcheck
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
