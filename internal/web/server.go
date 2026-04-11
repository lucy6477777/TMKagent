package web

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed static
var staticFiles embed.FS

// Server holds HTTP handler state.
type Server struct {
	upgrader   websocket.Upgrader
	mu         sync.Mutex
	cancel     context.CancelFunc
	lastUpload string
	cfg        ServerConfig
}

// ServerConfig holds the parameters needed to create a Server.
type ServerConfig struct {
	APIKey           string
	BaseURL          string
	DeepgramAPIKey   string
	Port             int
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
		cfg: cfg,
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
	ip := localIP()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ip": ip, "port": s.cfg.Port})
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

// cancelPipeline stops the current pipeline if one is running.
// Must be called with s.mu held.
func (s *Server) cancelPipeline() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

// startPipeline cancels any running pipeline and starts fn with a fresh context.
func (s *Server) startPipeline(fn func(ctx context.Context)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelPipeline()
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go fn(ctx)
}
