package audio

import (
	"math"
	"time"
)

// VADConfig holds tunable parameters for voice activity detection.
type VADConfig struct {
	SilenceThreshold  float64       // RMS energy below this level = silence
	SilenceDuration   time.Duration // consecutive silence before ending a chunk
	MinSpeechDuration time.Duration // minimum speech length; shorter chunks are discarded
	MaxChunkDuration  time.Duration // force-cut to avoid exceeding Whisper's 25MB limit
	SampleRate        int
}

// DefaultVADConfig returns production-ready defaults for a quiet microphone environment.
func DefaultVADConfig() VADConfig {
	return VADConfig{
		SilenceThreshold:  800.0,
		SilenceDuration:   400 * time.Millisecond,
		MinSpeechDuration: 200 * time.Millisecond,
		MaxChunkDuration:  8 * time.Second,
		SampleRate:        16000,
	}
}

type vadState int

const (
	vadSilence vadState = iota
	vadSpeaking
)

// VAD accumulates microphone frames and emits speech chunks.
type VAD struct {
	cfg          VADConfig
	state        vadState
	buf          []int16
	silenceFrames int // consecutive silent frames since last speech
	speechSamples int // samples accumulated while energy > threshold (excludes trailing silence)
}

// NewVAD creates a VAD with the given configuration.
func NewVAD(cfg VADConfig) *VAD {
	return &VAD{cfg: cfg}
}

// Feed processes one audio frame. Returns a complete speech chunk when a
// speech segment ends, nil otherwise. The caller owns the returned slice.
func (v *VAD) Feed(frame []int16) []int16 {
	energy := rms(frame)

	samplesPerFrame := len(frame)
	framesPerSecond := v.cfg.SampleRate / samplesPerFrame
	silenceFrameThreshold := int(v.cfg.SilenceDuration.Seconds() * float64(framesPerSecond))
	if silenceFrameThreshold < 1 {
		silenceFrameThreshold = 1
	}

	maxFrames := int(v.cfg.MaxChunkDuration.Seconds()) * framesPerSecond

	switch v.state {
	case vadSilence:
		if energy > v.cfg.SilenceThreshold {
			v.state = vadSpeaking
			v.buf = append(v.buf, frame...)
			v.silenceFrames = 0
			v.speechSamples += len(frame)
		}

	case vadSpeaking:
		v.buf = append(v.buf, frame...)
		if energy <= v.cfg.SilenceThreshold {
			v.silenceFrames++
			if v.silenceFrames >= silenceFrameThreshold {
				return v.flush()
			}
		} else {
			v.silenceFrames = 0
			v.speechSamples = len(v.buf) // all buffered samples are now considered speech
		}
		totalFrames := len(v.buf) / samplesPerFrame
		if totalFrames >= maxFrames {
			return v.flush()
		}
	}
	return nil
}

// flush finalises the current speech chunk and resets state.
// Returns nil if the speech portion is shorter than MinSpeechDuration.
func (v *VAD) flush() []int16 {
	chunk := v.buf
	speechSamples := v.speechSamples
	v.buf = nil
	v.state = vadSilence
	v.silenceFrames = 0
	v.speechSamples = 0

	minSamples := int(v.cfg.MinSpeechDuration.Seconds() * float64(v.cfg.SampleRate))
	if speechSamples < minSamples {
		return nil
	}
	result := make([]int16, len(chunk))
	copy(result, chunk)
	return result
}

// rms computes Root Mean Square energy of a frame of int16 samples.
func rms(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		f := float64(s)
		sum += f * f
	}
	return math.Sqrt(sum / float64(len(samples)))
}
