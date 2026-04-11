package tts

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"strings"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
)

const (
	playerSampleRate = 24000 // OpenAI TTS PCM output is 24kHz
	playerChannels   = 1
	playerFrameSize  = 4096 // ~170ms buffer at 24kHz; large buffer reduces underrun risk
)

// Player plays PCM audio through PortAudio speakers.
// The output stream is opened once and reused across sentences to avoid
// repeated ALSA initialisation overhead that causes underruns.
type Player struct {
	playing  atomic.Bool
	stream   *portaudio.Stream
	buf      []int16
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

// ensureStream lazily opens a persistent PortAudio output stream.
func (p *Player) ensureStream() error {
	if p.stream != nil {
		return nil
	}
	if !p.initDone {
		if err := portaudio.Initialize(); err != nil {
			return err
		}
		p.initDone = true
	}
	p.buf = make([]int16, playerFrameSize)
	stream, err := portaudio.OpenDefaultStream(0, playerChannels, float64(playerSampleRate), playerFrameSize, &p.buf)
	if err != nil {
		return err
	}
	if err := stream.Start(); err != nil {
		stream.Close()
		return err
	}
	p.stream = stream
	return nil
}

// Stop closes the persistent output stream. Call once when done with TTS.
func (p *Player) Stop() {
	if p.stream != nil {
		p.stream.Stop()  //nolint:errcheck
		p.stream.Close() //nolint:errcheck
		p.stream = nil
	}
}

// PlayStream reads ALL PCM data from r into memory, decodes to int16 samples,
// then feeds them to the persistent PortAudio output stream.
func (p *Player) PlayStream(r io.Reader) error {
	// 1. Buffer entire audio to decouple network I/O from playback
	var raw bytes.Buffer
	if _, err := io.Copy(&raw, r); err != nil {
		return err
	}
	pcmData := raw.Bytes()
	if len(pcmData) < 2 {
		return nil
	}

	// 2. Pre-decode all int16 samples (no CPU work during playback loop)
	totalSamples := len(pcmData) / 2
	samples := make([]int16, totalSamples)
	for i := 0; i < totalSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2 : i*2+2]))
	}

	// 3. Ensure persistent output stream is open
	if err := p.ensureStream(); err != nil {
		return err
	}

	p.playing.Store(true)
	defer p.playing.Store(false)

	// 4. Feed pre-decoded frames — underrun is non-fatal, just continue
	for offset := 0; offset < totalSamples; offset += playerFrameSize {
		end := offset + playerFrameSize
		if end > totalSamples {
			end = totalSamples
		}
		n := end - offset
		copy(p.buf[:n], samples[offset:end])
		for i := n; i < playerFrameSize; i++ {
			p.buf[i] = 0
		}
		if err := p.stream.Write(); err != nil {
			if isUnderrun(err) {
				log.Printf("[WARN] audio underrun (recovered, continuing)")
				continue
			}
			return err
		}
	}
	return nil
}

func isUnderrun(err error) bool {
	return err != nil && strings.Contains(err.Error(), "underflow")
}
