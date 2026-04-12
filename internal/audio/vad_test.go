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

	// SilenceDuration=400ms, SampleRate=16000, frameSize=512
	// silenceFrameThreshold = 0.4 * (16000/512) ≈ 12.5 → 12 frames
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

	// Only 1 loud frame (~32ms at 16kHz/512 samples) — below 300ms MinSpeechDuration
	vad.Feed(makeFrame(512, 1000))

	var chunk []int16
	for i := 0; i < 40; i++ {
		chunk = vad.Feed(makeFrame(512, 0))
		if chunk != nil {
			break
		}
	}
	if chunk != nil {
		t.Error("short noise burst should be discarded (below MinSpeechDuration)")
	}
}

func TestVAD_MaxChunkDuration_ForcesCut(t *testing.T) {
	cfg := DefaultVADConfig()
	cfg.SilenceThreshold = 100
	cfg.MinSpeechDuration = 0
	cfg.MaxChunkDuration = 100 * time.Millisecond
	cfg.SampleRate = 16000
	vad := NewVAD(cfg)

	// 100ms at 16kHz with 512-sample frames = 16000*0.1/512 ≈ 3.125 → 3 frames
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
	// Chunk should contain roughly 3 frames × 512 = 1536 samples
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
	result := rms(nil)
	if result != 0 {
		t.Errorf("rms(nil) = %f, want 0", result)
	}
}
