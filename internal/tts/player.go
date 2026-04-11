package tts

import (
	"encoding/binary"
	"io"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
)

const (
	playerSampleRate = 24000 // OpenAI TTS PCM output is 24kHz
	playerChannels   = 1
	playerFrameSize  = 1024
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

// PlayStream reads PCM 24kHz/16-bit/mono from r and plays through speakers.
// Blocks until all audio has been played or an error occurs.
func (p *Player) PlayStream(r io.Reader) error {
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

	readBuf := make([]byte, playerFrameSize*2) // 2 bytes per int16
	for {
		n, err := io.ReadFull(r, readBuf)
		if n > 0 {
			samples := n / 2
			for i := 0; i < samples; i++ {
				buf[i] = int16(binary.LittleEndian.Uint16(readBuf[i*2 : i*2+2]))
			}
			// Zero-pad the remaining buffer if partial read
			for i := samples; i < playerFrameSize; i++ {
				buf[i] = 0
			}
			if werr := stream.Write(); werr != nil {
				return werr
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
