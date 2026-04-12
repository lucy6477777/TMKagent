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
