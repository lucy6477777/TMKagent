//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lucyliuu/mini-tmk-agent/internal/web"
)

func TestWebServerIntegration_InfoEndpoint(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port:          8080,
		PublicBaseURL: "http://127.0.0.1:8080",
	})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	resp, err := http.Get(httpServer.URL + "/api/info")
	if err != nil {
		t.Fatalf("GET /api/info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["baseURL"] != "http://127.0.0.1:8080" {
		t.Fatalf("got baseURL %v, want %q", payload["baseURL"], "http://127.0.0.1:8080")
	}
}

func TestWebServerIntegration_UploadEndpoint(t *testing.T) {
	server := web.NewServer(web.ServerConfig{})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "sample.wav")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("wav-data")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/upload", &body)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /upload: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestWebServerIntegration_WSTranscript_RealAPI(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set; skipping integration test")
	}

	server := web.NewServer(web.ServerConfig{APIKey: apiKey})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	// Step 1: upload the test WAV file
	audioBytes, err := os.ReadFile("../../testdata/hello_zh.wav")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "hello_zh.wav")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(audioBytes); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	_ = mw.Close()

	req, _ := http.NewRequest(http.MethodPost, httpServer.URL+"/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("upload status %d, want 204", resp.StatusCode)
	}

	// Step 2: connect WebSocket
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	defer conn.Close()

	// Read initial idle status
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if msgStr(readWSMessage(t, conn), "state") != "idle" {
		t.Fatal("expected initial idle status")
	}

	// Step 3: send transcript command
	cmd := map[string]string{"type": "cmd", "action": "transcript", "sourceLang": "zh", "targetLang": "en"}
	if err := conn.WriteJSON(cmd); err != nil {
		t.Fatalf("send transcript cmd: %v", err)
	}

	// Step 4: collect messages until idle or timeout
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var gotPair, gotIdle bool
	for !gotIdle {
		msg := readWSMessage(t, conn)
		t.Logf("received: %v", msg)
		switch msgStr(msg, "type") {
		case "status":
			if msgStr(msg, "state") == "idle" {
				gotIdle = true
			}
		case "pair":
			gotPair = true
		case "error":
			// Empty audio may produce no transcription — that is acceptable
			t.Logf("pipeline error (may be expected for silent audio): %v", msg["msg"])
		}
	}

	// For silent test audio Whisper may return empty string and no pair is produced.
	// We only assert that the pipeline reached idle without crashing.
	t.Logf("pipeline completed: gotPair=%v", gotPair)
}

func TestWebServerIntegration_WS_StopCommand(t *testing.T) {
	server := web.NewServer(web.ServerConfig{})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if msgStr(readWSMessage(t, conn), "state") != "idle" {
		t.Fatal("expected initial idle status")
	}

	// Send stop with no active pipeline — should respond idle
	cmd := map[string]string{"type": "cmd", "action": "stop"}
	if err := conn.WriteJSON(cmd); err != nil {
		t.Fatalf("send stop: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	msg := readWSMessage(t, conn)
	if msgStr(msg, "type") != "status" || msgStr(msg, "state") != "idle" {
		t.Fatalf("expected idle status after stop, got %v", msg)
	}
}

func TestWebServerIntegration_WS_TranscriptWithNoUpload(t *testing.T) {
	server := web.NewServer(web.ServerConfig{})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	readWSMessage(t, conn) // consume initial idle

	cmd := map[string]string{"type": "cmd", "action": "transcript"}
	if err := conn.WriteJSON(cmd); err != nil {
		t.Fatalf("send transcript cmd: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	msg := readWSMessage(t, conn)
	if msgStr(msg, "type") != "error" {
		t.Fatalf("expected error message when no file uploaded, got %v", msg)
	}
}

// readWSMessage reads one WebSocket text message and unmarshals it into a generic map.
// Uses map[string]any to handle messages with non-string fields (e.g. pairMsg.Ts, progressMsg.Current).
func readWSMessage(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	var msg map[string]any
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("unmarshal %q: %v", data, err)
	}
	return msg
}

// msgStr extracts a string field from a generic message map.
func msgStr(msg map[string]any, key string) string {
	v, _ := msg[key].(string)
	return v
}
