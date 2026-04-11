package tts

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
)

const (
	playerSampleRate = 24000 // OpenAI TTS PCM output is 24kHz
	playerChannels   = 1
	playerFrameSize  = 2048 // larger frame = more resilient to scheduling jitter
)

// Player plays PCM audio through PortAudio speakers.
// It tracks playback state via an atomic flag for speaker-mode mic gating.
type Player struct {
	playing atomic.Bool
	// initDone tracks whether we initialised PortAudio inside this Player.
	// When the Capturer already called portaudio.Initialize(), we skip it.
	initDone bool
}

// NewPlayer creates a Player. skipInit should be true when PortAudio is
// already initialised by the audio Capturer (the typical stream-mode case).
func NewPlayer(skipInit bool) *Player {
	return &Player{initDone: !skipInit}
}

// IsPlaying returns true while audio is being played back.
func (p *Player) IsPlaying() bool {
	return p.playing.Load()
}

// PlayStream reads ALL PCM data from r into memory first, then plays through
// speakers in one shot. This avoids ALSA underruns caused by network I/O
// stalls when streaming from the TTS API.
func (p *Player) PlayStream(r io.Reader) error {
	// Buffer entire audio into memory to decouple network I/O from playback
	var audioBuf bytes.Buffer
	if _, err := io.Copy(&audioBuf, r); err != nil {
		return err
	}
	pcmData := audioBuf.Bytes()
	if len(pcmData) < 2 {
		return nil
	}

	if !p.initDone {
		if err := portaudio.Initialize(); err != nil {
			return err
		}
		p.initDone = true
	}

	buf := make([]int16, playerFrameSize)
	stream, err := portaudio.OpenDefaultStream(0, playerChannels, float64(playerSampleRate), playerFrameSize, &buf)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop() //nolint:errcheck

	p.playing.Store(true)
	defer p.playing.Store(false)

	totalSamples := len(pcmData) / 2
	for offset := 0; offset < totalSamples; offset += playerFrameSize {
		end := offset + playerFrameSize
		if end > totalSamples {
			end = totalSamples
		}
		// Decode int16 LE samples
		for i := offset; i < end; i++ {
			buf[i-offset] = int16(binary.LittleEndian.Uint16(pcmData[i*2 : i*2+2]))
		}
		// Zero-pad if last chunk is partial
		for i := end - offset; i < playerFrameSize; i++ {
			buf[i] = 0
		}
		if err := stream.Write(); err != nil {
			return err
		}
	}
	return nil
}
