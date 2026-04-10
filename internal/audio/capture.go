package audio

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
)

const (
	captureChannels   = 1
	captureSampleRate = 16000
	captureFrameSize  = 512 // ~32ms at 16kHz
)

// Capturer streams microphone audio via a channel.
type Capturer struct {
	stream *portaudio.Stream
	ch     chan []int16
}

// NewCapturer initialises portaudio and returns a Capturer and its output channel.
// The channel receives one frame (~32ms) of int16 PCM samples at a time.
// The caller must call Start() then Stop() when done.
func NewCapturer() (*Capturer, <-chan []int16, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, nil, fmt.Errorf(
			"portaudio initialise: %w\n  macOS: brew install portaudio\n  Linux: sudo apt install libportaudio2 libportaudio-dev",
			err,
		)
	}

	ch := make(chan []int16, 64)
	c := &Capturer{ch: ch}

	stream, err := portaudio.OpenDefaultStream(
		captureChannels, 0, captureSampleRate, captureFrameSize,
		func(in []int16) {
			frame := make([]int16, len(in))
			copy(frame, in)
			select {
			case ch <- frame:
			default:
				// Drop frame if pipeline is lagging — prevents blocking portaudio's real-time thread
			}
		},
	)
	if err != nil {
		portaudio.Terminate()
		return nil, nil, fmt.Errorf("opening audio stream: %w", err)
	}
	c.stream = stream
	return c, ch, nil
}

// Start begins audio capture.
func (c *Capturer) Start() error {
	return c.stream.Start()
}

// Stop ends capture and releases portaudio resources.
func (c *Capturer) Stop() {
	c.stream.Stop()  //nolint:errcheck
	c.stream.Close() //nolint:errcheck
	portaudio.Terminate()
	close(c.ch)
}
