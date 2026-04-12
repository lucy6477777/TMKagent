package audio

import (
	"testing"
	"time"
)

func makeFrame(size int, amplitude int16) []int16 {
	samples := make([]int16, size)
	for i := range samples {
		samples[i] = amplitude
	}
	return samples
}

func TestRMS_ZeroSamples(t *testing.T) {
	vad := NewVAD(DefaultVADConfig())
	result := vad.Feed(makeFrame(512, 0))
	if result != nil {
		t.Error("silence frame should return nil chunk")
	}
}

func TestVAD_SpeechThenSilenceProducesChunk(t *testing.T) {
	cfg := DefaultVADConfig()
	cfg.SilenceThreshold = 100
	cfg.MinSpeechDuration = 0
	vad := NewVAD(cfg)

	for i := 0; i < 5; i++ {
		chunk := vad.Feed(makeFrame(512, 1000))
		if chunk != nil {
			t.Error("should not return chunk mid-speech")
		}
	}

	var chunk []int16
	for i := 0; i < 30; i++ {
		chunk = vad.Feed(makeFrame(512, 0))
		if chunk != nil {
			break
		}
	}
	if chunk == nil {
		t.Error("expected a speech chunk after silence threshold exceeded")
	}
}

func TestVAD_ShortNoiseDiscarded(t *testing.T) {
	cfg := DefaultVADConfig()
	cfg.SilenceThreshold = 100
	cfg.MinSpeechDuration = 300 * time.Millisecond
	vad := NewVAD(cfg)

	vad.Feed(makeFrame(512, 1000))

	var chunk []int16
	for i := 0; i < 40; i++ {
		chunk = vad.Feed(makeFrame(512, 0))
		if chunk != nil {
			break
		}
	}
	if chunk != nil {
		t.Error("short noise burst should be discarded")
	}
}

func TestVAD_MaxChunkDuration_ForcesCut(t *testing.T) {
	cfg := DefaultVADConfig()
	cfg.SilenceThreshold = 100
	cfg.MinSpeechDuration = 0
	cfg.MaxChunkDuration = 100 * time.Millisecond
	cfg.SampleRate = 16000
	vad := NewVAD(cfg)

	// 100ms at 16kHz with 512-sample frames ≈ 3 frames before force-cut
	var chunk []int16
	for i := 0; i < 10; i++ {
		chunk = vad.Feed(makeFrame(512, 1000))
		if chunk != nil {
			break
		}
	}
	if chunk == nil {
		t.Fatal("expected force-cut at MaxChunkDuration")
	}
	if len(chunk) < 512 {
		t.Errorf("chunk too small: %d samples", len(chunk))
	}
}

func TestVAD_DefaultConfig(t *testing.T) {
	cfg := DefaultVADConfig()
	if cfg.SampleRate != 16000 {
		t.Errorf("default SampleRate = %d, want 16000", cfg.SampleRate)
	}
	if cfg.SilenceDuration != 400*time.Millisecond {
		t.Errorf("default SilenceDuration = %v, want 400ms", cfg.SilenceDuration)
	}
	if cfg.MaxChunkDuration != 8*time.Second {
		t.Errorf("default MaxChunkDuration = %v, want 8s", cfg.MaxChunkDuration)
	}
}

func TestRMS_EmptySlice(t *testing.T) {
	if rms(nil) != 0 {
		t.Error("rms(nil) should be 0")
	}
	if rms([]int16{}) != 0 {
		t.Error("rms(empty) should be 0")
	}
}
