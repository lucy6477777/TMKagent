package tts

import (
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
	playerFrameSize  = 4096 // ~170ms buffer at 24kHz
)

// Player plays PCM audio through PortAudio speakers.
// The output stream is opened once and reused across sentences.
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

// Stop closes the persistent output stream.
func (p *Player) Stop() {
	if p.stream != nil {
		p.stream.Stop()  //nolint:errcheck
		p.stream.Close() //nolint:errcheck
		p.stream = nil
	}
}

// PlayStream reads PCM 24kHz/16-bit/mono directly from r and plays in real-time.
// Starts playing as soon as the first chunk arrives from the network — no buffering.
// Underruns are non-fatal (brief glitch, playback continues).
func (p *Player) PlayStream(r io.Reader) error {
	if err := p.ensureStream(); err != nil {
		return err
	}

	p.playing.Store(true)
	defer p.playing.Store(false)

	readBuf := make([]byte, playerFrameSize*2) // 2 bytes per int16
	for {
		n, err := io.ReadFull(r, readBuf)
		if n > 0 {
			samples := n / 2
			for i := 0; i < samples; i++ {
				p.buf[i] = int16(binary.LittleEndian.Uint16(readBuf[i*2 : i*2+2]))
			}
			for i := samples; i < playerFrameSize; i++ {
				p.buf[i] = 0
			}
			if werr := p.stream.Write(); werr != nil {
				if isUnderrun(werr) {
					log.Printf("[WARN] audio underrun (recovered)")
					continue
				}
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

func isUnderrun(err error) bool {
	return err != nil && strings.Contains(err.Error(), "underflow")
}
