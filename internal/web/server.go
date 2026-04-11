package web

import (
	"context"
	"embed"
	"io/fs"
	"log"
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
	APIKey  string
	BaseURL string
	Port    int
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
