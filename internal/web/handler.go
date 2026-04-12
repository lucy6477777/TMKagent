package web

import (
	"context"
	"encoding/json"
	"fmt"
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

func translatePair(ctx context.Context, client translate.Client, text, sourceLang, targetLang string) (string, string, error) {
	if targetLang == "" || targetLang == sourceLang {
		return text, text, nil
	}

	normalizedSrc, translated, err := client.Translate(ctx, text, sourceLang, targetLang)
	if err != nil {
		return text, "[翻译失败]", err
	}
	if normalizedSrc != "" {
		return normalizedSrc, translated, nil
	}

	return text, translated, nil
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}
	defer func() {
		s.stopPipeline(conn)
		conn.Close()
	}()

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
			if s.hasActivePipelineForConn(conn, "stream") {
				sendJSON(conn, statusMsg{Type: "status", State: "listening"})
				continue
			}
			sendJSON(conn, statusMsg{Type: "status", State: "listening"})
			s.startPipeline(conn, "stream", func(ctx context.Context, pipelineID uint64) {
				s.runStream(ctx, pipelineID, conn, cmd.SourceLang, cmd.TargetLang)
			})

		case "stop":
			s.stopPipeline(conn)
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
			s.startPipeline(conn, "transcript", func(ctx context.Context, pipelineID uint64) {
				s.runTranscript(ctx, pipelineID, conn, uploadPath, cmd.SourceLang, cmd.TargetLang)
			})

		case "rtc_speaker_start":
			if s.cfg.LiveKitURL == "" {
				sendJSON(conn, errorMsg{Type: "error", Msg: "LiveKit not configured (set LIVEKIT_URL, LIVEKIT_API_KEY, LIVEKIT_API_SECRET)"})
				continue
			}
			if s.cfg.DeepgramAPIKey == "" {
				sendJSON(conn, errorMsg{Type: "error", Msg: "Deepgram not configured (set DEEPGRAM_API_KEY)"})
				continue
			}
			room := cmd.Room
			sourceLang := cmd.SourceLang
			targetLang := cmd.TargetLang
			if s.hasActivePipelineForConn(conn, "rtc_speaker") {
				sendJSON(conn, statusMsg{Type: "status", State: "listening"})
				continue
			}
			sendJSON(conn, statusMsg{Type: "status", State: "listening"})
			s.startPipeline(conn, "rtc_speaker", func(ctx context.Context, pipelineID uint64) {
				s.runRTCSpeaker(ctx, pipelineID, conn, room, sourceLang, targetLang)
			})

		case "rtc_join":
			if s.cfg.LiveKitURL == "" {
				sendJSON(conn, errorMsg{Type: "error", Msg: "LiveKit not configured (set LIVEKIT_URL, LIVEKIT_API_KEY, LIVEKIT_API_SECRET)"})
				continue
			}
			room := cmd.Room
			if s.hasActivePipelineForConn(conn, "rtc_listener") {
				sendJSON(conn, statusMsg{Type: "status", State: "listening"})
				continue
			}
			sendJSON(conn, statusMsg{Type: "status", State: "listening"})
			s.startPipeline(conn, "rtc_listener", func(ctx context.Context, pipelineID uint64) {
				s.runRTCListener(ctx, pipelineID, conn, room)
			})

		case "rtc_stop":
			s.stopPipeline(conn)
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
func (s *Server) runStream(ctx context.Context, pipelineID uint64, conn *websocket.Conn, sourceLang, targetLang string) {
	if s.cfg.DeepgramAPIKey == "" {
		s.failPipeline(pipelineID, conn, "DEEPGRAM_API_KEY not set")
		return
	}

	streamClient := asr.NewDeepgramStreamClient(s.cfg.DeepgramAPIKey)
	translateClient := translate.NewOpenAIClient(s.cfg.APIKey, s.cfg.BaseURL)

	session, err := streamClient.Connect(ctx, sourceLang)
	if err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
		return
	}
	defer session.Close()

	capturer, frameCh, err := audio.NewCapturer()
	if err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
		return
	}
	if err := capturer.Start(); err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
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
			_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
			return
		default:
		}

		if !result.IsFinal {
			if err := s.sendPipelineJSON(pipelineID, conn, interimMsg{Type: "interim", Text: result.Text}); err != nil {
				return
			}
			continue
		}

		// Final result — translate then send pair
		src, target, err := translatePair(ctx, translateClient, result.Text, sourceLang, targetLang)
		if err != nil {
			log.Printf("[WARN] translation error: %v", err)
		}
		if err := s.sendPipelineJSON(pipelineID, conn, pairMsg{
			Type:   "pair",
			Source: src,
			Target: target,
			Ts:     time.Now().UnixMilli(),
		}); err != nil {
			return
		}
	}

	_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
}

// runTranscript reads an audio file, transcribes it chunk-by-chunk with progress updates,
// translates each chunk, and sends pairMsg over the WebSocket.
func (s *Server) runTranscript(ctx context.Context, pipelineID uint64, conn *websocket.Conn, filePath, sourceLang, targetLang string) {
	asrClient := asr.NewWhisperClient(s.cfg.APIKey, s.cfg.BaseURL)
	translateClient := translate.NewOpenAIClient(s.cfg.APIKey, s.cfg.BaseURL)

	audioBytes, filename, err := audio.ReadAudioFile(filePath)
	if err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
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
			_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
			return
		default:
		}

		if err := s.sendPipelineJSON(pipelineID, conn, progressMsg{Type: "progress", Current: i + 1, Total: total}); err != nil {
			return
		}

		text, err := asrClient.Transcribe(ctx, chunk, filename, sourceLang)
		if err != nil {
			log.Printf("[WARN] ASR error on chunk %d: %v", i, err)
			if err := s.sendPipelineJSON(pipelineID, conn, errorMsg{Type: "error", Msg: err.Error()}); err != nil {
				return
			}
			continue
		}
		if text == "" {
			continue
		}

		src, target, err := translatePair(ctx, translateClient, text, sourceLang, targetLang)
		if err != nil {
			log.Printf("[WARN] translation error on chunk %d: %v", i, err)
		}

		if err := s.sendPipelineJSON(pipelineID, conn, pairMsg{
			Type:   "pair",
			Source: src,
			Target: target,
			Ts:     time.Now().UnixMilli(),
		}); err != nil {
			return
		}
	}

	_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
}

// runRTCSpeaker connects to LiveKit as speaker, runs mic→ASR→translate pipeline,
// and broadcasts pairs to the room while also sending them to the WebSocket client.
func (s *Server) runRTCSpeaker(ctx context.Context, pipelineID uint64, conn *websocket.Conn, roomName, sourceLang, targetLang string) {
	identity := fmt.Sprintf("speaker-web-%d", pipelineID)
	rtcClient, err := rtc.Connect(ctx, s.cfg.LiveKitURL, s.cfg.LiveKitAPIKey, s.cfg.LiveKitAPISecret, roomName, identity)
	if err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
		return
	}
	defer rtcClient.Close()

	streamClient := asr.NewDeepgramStreamClient(s.cfg.DeepgramAPIKey)
	translateClient := translate.NewOpenAIClient(s.cfg.APIKey, s.cfg.BaseURL)

	session, err := streamClient.Connect(ctx, sourceLang)
	if err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
		return
	}
	defer session.Close()

	capturer, frameCh, err := audio.NewCapturer()
	if err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
		return
	}
	if err := capturer.Start(); err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
		return
	}
	defer capturer.Stop()

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

	for result := range session.Results() {
		select {
		case <-ctx.Done():
			_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
			return
		default:
		}

		if !result.IsFinal {
			rtcClient.Send(rtc.RelayMsg{Type: "interim", Text: result.Text}) //nolint:errcheck
			if err := s.sendPipelineJSON(pipelineID, conn, interimMsg{Type: "interim", Text: result.Text}); err != nil {
				return
			}
			continue
		}

		src, target, err := translatePair(ctx, translateClient, result.Text, sourceLang, targetLang)
		if err != nil {
			log.Printf("[WARN] translation error: %v", err)
		}

		// Broadcast to room
		rtcClient.Send(rtc.RelayMsg{Type: "pair", Source: src, Target: target}) //nolint:errcheck

		// Also send to this WebSocket client
		if err := s.sendPipelineJSON(pipelineID, conn, pairMsg{Type: "pair", Source: src, Target: target, Ts: time.Now().UnixMilli()}); err != nil {
			return
		}
	}

	_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
}

// runRTCListener connects to LiveKit as listener and forwards received pairs to the WebSocket client.
func (s *Server) runRTCListener(ctx context.Context, pipelineID uint64, conn *websocket.Conn, roomName string) {
	identity := fmt.Sprintf("listener-web-%d", pipelineID)
	rtcClient, err := rtc.Connect(ctx, s.cfg.LiveKitURL, s.cfg.LiveKitAPIKey, s.cfg.LiveKitAPISecret, roomName, identity)
	if err != nil {
		s.failPipeline(pipelineID, conn, err.Error())
		return
	}
	defer rtcClient.Close()

	msgs := rtcClient.Messages()
	for {
		select {
		case <-ctx.Done():
			_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
			return
		case msg, ok := <-msgs:
			if !ok {
				_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
				return
			}
			switch msg.Type {
			case "interim":
				if err := s.sendPipelineJSON(pipelineID, conn, interimMsg{Type: "interim", Text: msg.Text}); err != nil {
					return
				}
			case "pair":
				if err := s.sendPipelineJSON(pipelineID, conn, pairMsg{
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

// failPipeline sends an error message and transitions the pipeline to idle.
func (s *Server) failPipeline(pipelineID uint64, conn *websocket.Conn, msg string) {
	_ = s.sendPipelineJSON(pipelineID, conn, errorMsg{Type: "error", Msg: msg})
	_ = s.sendPipelineJSON(pipelineID, conn, statusMsg{Type: "status", State: "idle"})
}
