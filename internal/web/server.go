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
		if r.URL.Path != "/" {
			if _, err := fs.Stat(stripped, r.URL.Path[1:]); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
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
