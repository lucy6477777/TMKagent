package web

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lucyliuu/mini-tmk-agent/internal/asr"
	"github.com/lucyliuu/mini-tmk-agent/internal/audio"
	"github.com/lucyliuu/mini-tmk-agent/internal/rtc"
	"github.com/lucyliuu/mini-tmk-agent/internal/translate"
)

const maxChunkBytes = 20 * 1024 * 1024 // 20MB — below Whisper's 25MB limit

// Incoming message from frontend.
type cmdMsg struct {
	Type       string `json:"type"`
	Action     string `json:"action"`
	SourceLang string `json:"sourceLang"`
	TargetLang string `json:"targetLang"`
	Room       string `json:"room"`
	Role       string `json:"role"`
}

// Outgoing messages to frontend.
type pairMsg struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Target string `json:"target"`
	Ts     int64  `json:"ts"`
}

type statusMsg struct {
	Type  string `json:"type"`
	State string `json:"state"`
}

type progressMsg struct {
	Type    string `json:"type"`
	Current int    `json:"current"`
	Total   int    `json:"total"`
}

type errorMsg struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

type interimMsg struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}
	defer conn.Close()

	sendJSON(conn, statusMsg{Type: "status", State: "idle"})

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var cmd cmdMsg
		if err := json.Unmarshal(msg, &cmd); err != nil {
			continue
		}
		if cmd.Type != "cmd" {
			continue
		}

		switch cmd.Action {
		case "start_stream":
			sendJSON(conn, statusMsg{Type: "status", State: "listening"})
			s.startPipeline(func(ctx context.Context) {
				s.runStream(ctx, conn, cmd.SourceLang, cmd.TargetLang)
			})

		case "stop":
			s.mu.Lock()
			s.cancelPipeline()
			s.mu.Unlock()
			sendJSON(conn, statusMsg{Type: "status", State: "idle"})

		case "transcript":
			s.mu.Lock()
			uploadPath := s.lastUpload
			s.mu.Unlock()
			if uploadPath == "" {
				sendJSON(conn, errorMsg{Type: "error", Msg: "no file uploaded"})
				continue
			}
			sendJSON(conn, statusMsg{Type: "status", State: "processing"})
			s.startPipeline(func(ctx context.Context) {
				s.runTranscript(ctx, conn, uploadPath, cmd.SourceLang, cmd.TargetLang)
			})

		case "rtc_speaker_start":
			if s.cfg.LiveKitURL == "" {
				sendJSON(conn, errorMsg{Type: "error", Msg: "LiveKit not configured (set LIVEKIT_URL, LIVEKIT_API_KEY, LIVEKIT_API_SECRET)"})
				continue
			}
			room := cmd.Room
			sourceLang := cmd.SourceLang
			targetLang := cmd.TargetLang
			sendJSON(conn, statusMsg{Type: "status", State: "listening"})
			s.startPipeline(func(ctx context.Context) {
				s.runRTCSpeaker(ctx, conn, room, sourceLang, targetLang)
			})

		case "rtc_join":
			if s.cfg.LiveKitURL == "" {
				sendJSON(conn, errorMsg{Type: "error", Msg: "LiveKit not configured (set LIVEKIT_URL, LIVEKIT_API_KEY, LIVEKIT_API_SECRET)"})
				continue
			}
			room := cmd.Room
			sendJSON(conn, statusMsg{Type: "status", State: "listening"})
			s.startPipeline(func(ctx context.Context) {
				s.runRTCListener(ctx, conn, room)
			})

		case "rtc_stop":
			s.mu.Lock()
			s.cancelPipeline()
			s.mu.Unlock()
			sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		}
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "invalid multipart request", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	tmpFile, err := os.CreateTemp("", "tmk-upload-*"+ext)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.lastUpload = tmpFile.Name()
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// runStream runs a microphone→Deepgram streaming ASR→translate pipeline.
// Interim results are sent as interimMsg; final results are translated and sent as pairMsg.
func (s *Server) runStream(ctx context.Context, conn *websocket.Conn, sourceLang, targetLang string) {
	if s.cfg.DeepgramAPIKey == "" {
		sendJSON(conn, errorMsg{Type: "error", Msg: "DEEPGRAM_API_KEY not set"})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}

	streamClient := asr.NewDeepgramStreamClient(s.cfg.DeepgramAPIKey)
	translateClient := translate.NewOpenAIClient(s.cfg.APIKey, s.cfg.BaseURL)

	session, err := streamClient.Connect(ctx, sourceLang)
	if err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}
	defer session.Close()

	capturer, frameCh, err := audio.NewCapturer()
	if err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}
	if err := capturer.Start(); err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}
	defer capturer.Stop()

	// g1: stream audio frames to Deepgram
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case frame, ok := <-frameCh:
				if !ok {
					return
				}
				if err := session.Send(audio.Int16ToBytes(frame)); err != nil {
					log.Printf("[WARN] deepgram send: %v", err)
					return
				}
			}
		}
	}()

	// g2: receive results, forward interim immediately, translate finals
	for result := range session.Results() {
		select {
		case <-ctx.Done():
			sendJSON(conn, statusMsg{Type: "status", State: "idle"}) //nolint:errcheck
			return
		default:
		}

		if !result.IsFinal {
			sendJSON(conn, interimMsg{Type: "interim", Text: result.Text}) //nolint:errcheck
			continue
		}

		// Final result — translate then send pair
		src := result.Text
		target := ""
		if targetLang != "" && targetLang != sourceLang {
			normalizedSrc, translated, err := translateClient.Translate(ctx, result.Text, sourceLang, targetLang)
			if err != nil {
				log.Printf("[WARN] translation error: %v", err)
				target = "[翻译失败]"
			} else {
				target = translated
				if normalizedSrc != "" {
					src = normalizedSrc
				}
			}
		} else {
			target = result.Text
		}
		if err := sendJSON(conn, pairMsg{
			Type:   "pair",
			Source: src,
			Target: target,
			Ts:     time.Now().UnixMilli(),
		}); err != nil {
			return
		}
	}

	sendJSON(conn, statusMsg{Type: "status", State: "idle"}) //nolint:errcheck
}

// runTranscript reads an audio file, transcribes it chunk-by-chunk with progress updates,
// translates each chunk, and sends pairMsg over the WebSocket.
func (s *Server) runTranscript(ctx context.Context, conn *websocket.Conn, filePath, sourceLang, targetLang string) {
	asrClient := asr.NewWhisperClient(s.cfg.APIKey, s.cfg.BaseURL)
	translateClient := translate.NewOpenAIClient(s.cfg.APIKey, s.cfg.BaseURL)

	audioBytes, filename, err := audio.ReadAudioFile(filePath)
	if err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}

	// Split into chunks
	var chunks [][]byte
	for i := 0; i < len(audioBytes); i += maxChunkBytes {
		end := i + maxChunkBytes
		if end > len(audioBytes) {
			end = len(audioBytes)
		}
		chunks = append(chunks, audioBytes[i:end])
	}
	total := len(chunks)

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			sendJSON(conn, statusMsg{Type: "status", State: "idle"})
			return
		default:
		}

		sendJSON(conn, progressMsg{Type: "progress", Current: i + 1, Total: total})

		text, err := asrClient.Transcribe(ctx, chunk, filename, sourceLang)
		if err != nil {
			log.Printf("[WARN] ASR error on chunk %d: %v", i, err)
			sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
			continue
		}
		if text == "" {
			continue
		}

		src := text
		target := ""
		if targetLang != "" && targetLang != sourceLang {
			normalizedSrc, translated, err := translateClient.Translate(ctx, text, sourceLang, targetLang)
			if err != nil {
				log.Printf("[WARN] translation error on chunk %d: %v", i, err)
				target = "[翻译失败]"
			} else {
				target = translated
				if normalizedSrc != "" {
					src = normalizedSrc
				}
			}
		} else {
			target = text
		}

		if err := sendJSON(conn, pairMsg{
			Type:   "pair",
			Source: src,
			Target: target,
			Ts:     time.Now().UnixMilli(),
		}); err != nil {
			return
		}
	}

	sendJSON(conn, statusMsg{Type: "status", State: "idle"}) //nolint:errcheck
}

// runRTCSpeaker connects to LiveKit as speaker, runs mic→ASR→translate pipeline,
// and broadcasts pairs to the room while also sending them to the WebSocket client.
func (s *Server) runRTCSpeaker(ctx context.Context, conn *websocket.Conn, roomName, sourceLang, targetLang string) {
	rtcClient, err := rtc.Connect(ctx, s.cfg.LiveKitURL, s.cfg.LiveKitAPIKey, s.cfg.LiveKitAPISecret, roomName, "speaker-web")
	if err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}
	defer rtcClient.Close()

	asrClient := asr.NewWhisperClient(s.cfg.APIKey, s.cfg.BaseURL)
	translateClient := translate.NewOpenAIClient(s.cfg.APIKey, s.cfg.BaseURL)

	capturer, frameCh, err := audio.NewCapturer()
	if err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}
	if err := capturer.Start(); err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}
	defer capturer.Stop()

	vad := audio.NewVAD(audio.DefaultVADConfig())
	audioCh := make(chan []int16, 8)
	asrCh := make(chan string, 8)

	go func() {
		defer close(audioCh)
		for {
			select {
			case <-ctx.Done():
				return
			case frame, ok := <-frameCh:
				if !ok {
					return
				}
				if chunk := vad.Feed(frame); chunk != nil {
					select {
					case audioCh <- chunk:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	go func() {
		defer close(asrCh)
		for chunk := range audioCh {
			pcmBytes := audio.Int16ToBytes(chunk)
			wavBytes := audio.PCMToWAV(pcmBytes, 16000, 1, 16)
			text, err := asrClient.Transcribe(ctx, wavBytes, "audio.wav", sourceLang)
			if err != nil {
				log.Printf("[WARN] ASR error: %v", err)
				continue
			}
			if text == "" {
				continue
			}
			select {
			case asrCh <- text:
			case <-ctx.Done():
				return
			}
		}
	}()

	for text := range asrCh {
		src := text
		target := ""
		if targetLang != "" && targetLang != sourceLang {
			normalizedSrc, translated, err := translateClient.Translate(ctx, text, sourceLang, targetLang)
			if err != nil {
				log.Printf("[WARN] translation error: %v", err)
				target = "[翻译失败]"
			} else {
				target = translated
				if normalizedSrc != "" {
					src = normalizedSrc
				}
			}
		} else {
			target = text
		}

		// Broadcast to room
		rtcClient.Send(rtc.RelayMsg{Type: "pair", Source: src, Target: target}) //nolint:errcheck

		// Also send to this WebSocket client
		if err := sendJSON(conn, pairMsg{Type: "pair", Source: src, Target: target, Ts: time.Now().UnixMilli()}); err != nil {
			return
		}
	}

	sendJSON(conn, statusMsg{Type: "status", State: "idle"}) //nolint:errcheck
}

// runRTCListener connects to LiveKit as listener and forwards received pairs to the WebSocket client.
func (s *Server) runRTCListener(ctx context.Context, conn *websocket.Conn, roomName string) {
	rtcClient, err := rtc.Connect(ctx, s.cfg.LiveKitURL, s.cfg.LiveKitAPIKey, s.cfg.LiveKitAPISecret, roomName, "listener-web")
	if err != nil {
		sendJSON(conn, errorMsg{Type: "error", Msg: err.Error()})
		sendJSON(conn, statusMsg{Type: "status", State: "idle"})
		return
	}
	defer rtcClient.Close()

	msgs := rtcClient.Messages()
	for {
		select {
		case <-ctx.Done():
			sendJSON(conn, statusMsg{Type: "status", State: "idle"}) //nolint:errcheck
			return
		case msg, ok := <-msgs:
			if !ok {
				sendJSON(conn, statusMsg{Type: "status", State: "idle"}) //nolint:errcheck
				return
			}
			if msg.Type == "pair" {
				if err := sendJSON(conn, pairMsg{
					Type:   "pair",
					Source: msg.Source,
					Target: msg.Target,
					Ts:     time.Now().UnixMilli(),
				}); err != nil {
					return
				}
			}
		}
	}
}

func sendJSON(conn *websocket.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}
