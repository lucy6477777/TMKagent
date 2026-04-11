package asr

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TranscriptResult holds one ASR result from the streaming service.
type TranscriptResult struct {
	Text    string
	IsFinal bool // false = interim (still speaking), true = final (utterance complete)
}

// StreamClient opens streaming ASR sessions.
type StreamClient interface {
	Connect(ctx context.Context, lang string) (StreamSession, error)
}

// StreamSession is a live bidirectional ASR connection.
type StreamSession interface {
	Send(pcmBytes []byte) error
	Results() <-chan TranscriptResult
	Close() error
}

// --- Deepgram implementation using raw gorilla/websocket ---

const deepgramWSEndpoint = "wss://api.deepgram.com/v1/listen"

// DeepgramStreamClient implements StreamClient via Deepgram's WebSocket API.
type DeepgramStreamClient struct {
	apiKey string
}

func NewDeepgramStreamClient(apiKey string) *DeepgramStreamClient {
	return &DeepgramStreamClient{apiKey: apiKey}
}

func (d *DeepgramStreamClient) Connect(ctx context.Context, lang string) (StreamSession, error) {
	params := url.Values{}
	params.Set("model", "nova-2")
	params.Set("encoding", "linear16")
	params.Set("sample_rate", "16000")
	params.Set("channels", "1")
	params.Set("interim_results", "true")
	params.Set("endpointing", "300")
	params.Set("smart_format", "true")
	params.Set("vad_events", "true")
	if lang != "" {
		params.Set("language", lang)
	}

	wsURL := fmt.Sprintf("%s?%s", deepgramWSEndpoint, params.Encode())

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	header := map[string][]string{
		"Authorization": {fmt.Sprintf("Token %s", d.apiKey)},
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("deepgram connect: %w", err)
	}

	sess := &deepgramSession{
		conn:      conn,
		resultsCh: make(chan TranscriptResult, 32),
		done:      make(chan struct{}),
	}
	go sess.readLoop()
	return sess, nil
}

type deepgramSession struct {
	conn      *websocket.Conn
	resultsCh chan TranscriptResult
	done      chan struct{}
	closeOnce sync.Once
}

// Send writes raw PCM audio bytes to the Deepgram WebSocket.
func (s *deepgramSession) Send(pcmBytes []byte) error {
	return s.conn.WriteMessage(websocket.BinaryMessage, pcmBytes)
}

func (s *deepgramSession) Results() <-chan TranscriptResult {
	return s.resultsCh
}

func (s *deepgramSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		// Send CloseStream message per Deepgram protocol
		closeMsg, _ := json.Marshal(map[string]string{"type": "CloseStream"})
		_ = s.conn.WriteMessage(websocket.TextMessage, closeMsg)
		err = s.conn.Close()
		<-s.done // wait for readLoop to finish
	})
	return err
}

// deepgramResponse is the subset of Deepgram's streaming JSON we care about.
type deepgramResponse struct {
	Type    string `json:"type"`
	Channel struct {
		Alternatives []struct {
			Transcript string `json:"transcript"`
		} `json:"alternatives"`
	} `json:"channel"`
	IsFinal     bool `json:"is_final"`
	SpeechFinal bool `json:"speech_final"`
}

func (s *deepgramSession) readLoop() {
	defer close(s.done)
	defer close(s.resultsCh)
	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[WARN] deepgram read: %v", err)
			}
			return
		}

		var resp deepgramResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			continue
		}

		if resp.Type != "Results" {
			continue
		}
		if len(resp.Channel.Alternatives) == 0 {
			continue
		}
		text := resp.Channel.Alternatives[0].Transcript
		if text == "" {
			continue
		}

		select {
		case s.resultsCh <- TranscriptResult{
			Text:    text,
			IsFinal: resp.IsFinal || resp.SpeechFinal,
		}:
		default:
		}
	}
}
