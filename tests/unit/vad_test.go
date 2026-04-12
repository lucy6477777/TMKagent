package unit_test

import (
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/audio"
)

// makeFrame generates a frame of int16 samples with given amplitude.
func makeFrame(size int, amplitude int16) []int16 {
	samples := make([]int16, size)
	for i := range samples {
		samples[i] = amplitude
	}
	return samples
}

func TestRMS_ZeroSamples(t *testing.T) {
	vad := audio.NewVAD(audio.DefaultVADConfig())
	result := vad.Feed(makeFrame(512, 0))
	if result != nil {
		t.Error("silence frame should return nil chunk")
	}
}

func TestVAD_SpeechThenSilenceProducesChunk(t *testing.T) {
	cfg := audio.DefaultVADConfig()
	cfg.SilenceThreshold = 100
	cfg.MinSpeechDuration = 0 // disable minimum for this test
	vad := audio.NewVAD(cfg)

	// Feed 5 loud frames (speech)
	for i := 0; i < 5; i++ {
		chunk := vad.Feed(makeFrame(512, 1000))
		if chunk != nil {
			t.Error("should not return chunk mid-speech")
		}
	}

	// Feed enough silent frames to trigger end-of-speech
	// SilenceDuration=800ms, SampleRate=16000, frameSize=512
	// silenceFrameThreshold = 800 * 16000/1000 / 512 ≈ 25 frames
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
	cfg := audio.DefaultVADConfig()
	cfg.SilenceThreshold = 100
	cfg.MinSpeechDuration = 300 * 1e6 // 300ms in nanoseconds
	vad := audio.NewVAD(cfg)

	// Only 1 loud frame (very short, less than MinSpeechDuration)
	vad.Feed(makeFrame(512, 1000))

	// Then lots of silence
	var chunk []int16
	for i := 0; i < 40; i++ {
		chunk = vad.Feed(makeFrame(512, 0))
		if chunk != nil {
			break
		}
	}
	// 1 frame × 512 samples / 16000 sampleRate ≈ 32ms < 300ms MinSpeechDuration → discarded
	if chunk != nil {
		t.Error("short noise burst should be discarded (below MinSpeechDuration)")
	}
}
