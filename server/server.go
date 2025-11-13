package server

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed assets/*
var assetsFS embed.FS

// Server represents the HTTP server
type Server struct {
	addr    string
	engines []EngineInfo
}

// InitDebugLogging sets up logging to a file only (no stdout to avoid breaking TUI)
func InitDebugLogging(filename string) error {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	// Log only to file to avoid breaking TUI layout
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	return nil
}

// New creates a new chess server
func New(addr string) *Server {
	engines := DiscoverEngines()
	log.Printf("Discovered %d chess engines", len(engines))
	for _, engine := range engines {
		log.Printf("  - %s (%s)", engine.Name, engine.Path)
	}
	
	return &Server{
		addr:    addr,
		engines: engines,
	}
}

// GetEngines returns the list of discovered engines
func (s *Server) GetEngines() []EngineInfo {
	return s.engines
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Serve static assets
	http.Handle("/assets/", http.FileServer(http.FS(assetsFS)))

	// API endpoints
	http.HandleFunc("/api/computer-move", s.handleComputerMove)
	http.HandleFunc("/api/analysis", s.handleAnalysisWebSocket)
	http.HandleFunc("/api/engines", s.handleGetEngines)

	// Serve main page
	http.HandleFunc("/", s.handleIndex)

	log.Printf("Server starting on %s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// handleIndex serves the main chess board page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}

	// Pass engines and cache buster to template
	data := struct {
		Engines     []EngineInfo
		CacheBuster int64
	}{
		Engines:     s.engines,
		CacheBuster: time.Now().UnixNano(),
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Render error: %v", err)
	}
}

// handleGetEngines returns the list of discovered engines with their capabilities
func (s *Server) handleGetEngines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	engines := s.engines
	if engines == nil {
		engines = []EngineInfo{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(engines)
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.addr
}
