package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed static
var staticFiles embed.FS

// Server holds HTTP handler state.
type Server struct {
	upgrader   websocket.Upgrader
	mu         sync.Mutex
	nextID     uint64
	pipelines  map[*websocket.Conn]connPipeline
	lastUpload string
	cfg        ServerConfig
}

type connPipeline struct {
	id     uint64
	kind   string
	cancel context.CancelFunc
}

var errStalePipeline = errors.New("stale pipeline")

// ServerConfig holds the parameters needed to create a Server.
type ServerConfig struct {
	APIKey           string
	BaseURL          string
	DeepgramAPIKey   string
	Port             int
	PublicBaseURL    string
	LiveKitURL       string
	LiveKitAPIKey    string
	LiveKitAPISecret string
}

// NewServer creates a Server with the given config.
func NewServer(cfg ServerConfig) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		pipelines: make(map[*websocket.Conn]connPipeline),
		cfg:       cfg,
	}
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	stripped, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	fileServer := http.FileServer(http.FS(stripped))

	mux.HandleFunc("/assets/", fileServer.ServeHTTP)
	mux.HandleFunc("/ws", s.handleWS)
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the exact file
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		// Strip leading slash for fs.Stat
		fsPath := path[1:]
		if _, err := fs.Stat(stripped, fsPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback: serve index.html
		indexContent, err := fs.ReadFile(stripped, "index.html")
		if err != nil {
			// During development (no static files built yet), return 200 with placeholder
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<!DOCTYPE html><html><body>Build frontend with 'make web-build'</body></html>"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexContent)
	})

	return mux
}

// handleInfo returns the server's local IP and port as JSON.
func (s *Server) handleInfo(w http.ResponseWriter, _ *http.Request) {
	baseURL, ip, port := s.publicAccessInfo()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ip":       ip,
		"port":     port,
		"baseURL":  baseURL,
	})
}

func (s *Server) publicAccessInfo() (baseURL, ip string, port int) {
	port = s.cfg.Port
	baseURL = strings.TrimSpace(s.cfg.PublicBaseURL)
	if baseURL == "" {
		ip = localIP()
		baseURL = "http://" + net.JoinHostPort(ip, strconv.Itoa(port))
		return baseURL, ip, port
	}

	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" {
		ip = localIP()
		baseURL = "http://" + net.JoinHostPort(ip, strconv.Itoa(port))
		return baseURL, ip, port
	}

	if parsed.Scheme == "" {
		parsed.Scheme = "http"
	}
	baseURL = strings.TrimRight(parsed.String(), "/")

	host := parsed.Hostname()
	if host == "" {
		host = localIP()
	}
	ip = host
	if parsedPort := parsed.Port(); parsedPort != "" {
		if n, err := strconv.Atoi(parsedPort); err == nil {
			port = n
		}
	}

	return baseURL, ip, port
}

// localIP returns the first non-loopback IPv4 address found.
func localIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "localhost"
}

// startPipeline replaces the active pipeline for this WebSocket connection and starts fn.
func (s *Server) startPipeline(conn *websocket.Conn, kind string, fn func(ctx context.Context, pipelineID uint64)) uint64 {
	s.mu.Lock()
	prevPipeline, hadPrev := s.pipelines[conn]
	s.nextID++
	pipelineID := s.nextID
	ctx, cancel := context.WithCancel(context.Background())
	s.pipelines[conn] = connPipeline{
		id:     pipelineID,
		kind:   kind,
		cancel: cancel,
	}
	s.mu.Unlock()

	if hadPrev && prevPipeline.cancel != nil {
		prevPipeline.cancel()
	}
	go fn(ctx, pipelineID)
	return pipelineID
}

// stopPipeline cancels the active pipeline for conn and clears its identity.
func (s *Server) stopPipeline(conn *websocket.Conn) {
	s.mu.Lock()
	pipeline, ok := s.pipelines[conn]
	if ok {
		delete(s.pipelines, conn)
	}
	s.mu.Unlock()

	if ok && pipeline.cancel != nil {
		pipeline.cancel()
	}
}

// isActivePipeline reports whether pipelineID is still the current pipeline for conn.
func (s *Server) isActivePipeline(conn *websocket.Conn, pipelineID uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	pipeline, ok := s.pipelines[conn]
	return ok && pipeline.id == pipelineID && pipeline.cancel != nil
}

// hasActivePipelineForConn reports whether conn already owns an active pipeline of kind.
func (s *Server) hasActivePipelineForConn(conn *websocket.Conn, kind string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	pipeline, ok := s.pipelines[conn]
	return ok && pipeline.kind == kind && pipeline.cancel != nil
}

// sendPipelineJSON sends a message only if pipelineID is still current.
func (s *Server) sendPipelineJSON(pipelineID uint64, conn *websocket.Conn, v any) error {
	if !s.isActivePipeline(conn, pipelineID) {
		return errStalePipeline
	}
	return sendJSON(conn, v)
}
