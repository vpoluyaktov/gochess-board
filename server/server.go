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
	addr         string
	engines      []EngineInfo
	openingBook  *OpeningBook  // Opening name database
	polyglotBook *PolyglotBook // Polyglot opening book for move suggestions
}

// InitDebugLogging sets up logging to a file only (no stdout to avoid breaking TUI)
func InitDebugLogging(filename string) error {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	// Log only to file to avoid breaking TUI layout
	// Use custom format without file:line, we'll use module prefixes instead
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	return nil
}

// New creates a new chess server
func New(addr string, bookFile string) *Server {
	engines := DiscoverEngines(bookFile)
	Info("SERVER", "Discovered %d chess engines", len(engines))
	for _, engine := range engines {
		Info("SERVER", "  - %s (%s)", engine.Name, engine.Path)
	}

	// Initialize opening name database from embedded filesystem
	Info("SERVER", "Loading opening name database from embedded assets/openings")
	openingBook := NewOpeningBook()
	if err := openingBook.LoadFromEmbedded(assetsFS, "assets/openings"); err != nil {
		Warn("SERVER", "Failed to load opening name database: %v", err)
	} else {
		stats := openingBook.Stats()
		Info("SERVER", "Opening name database loaded: %d openings, %d nodes, max depth %d",
			stats["total_openings"], stats["total_nodes"], stats["max_depth"])
	}

	// Initialize Polyglot opening book if specified
	var polyglotBook *PolyglotBook
	if bookFile != "" {
		Info("SERVER", "Loading Polyglot opening book from %s", bookFile)
		polyglotBook = NewPolyglotBook()
		if err := polyglotBook.LoadFromFile(bookFile); err != nil {
			Warn("SERVER", "Failed to load Polyglot book: %v", err)
			polyglotBook = nil
		} else {
			Info("SERVER", "Polyglot book loaded successfully")
		}
	}

	return &Server{
		addr:         addr,
		engines:      engines,
		openingBook:  openingBook,
		polyglotBook: polyglotBook,
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
	http.HandleFunc("/api/opening", s.handleGetOpening)

	// Serve main page
	http.HandleFunc("/", s.handleIndex)

	Info("SERVER", "Server starting on %s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// handleIndex serves the main chess board page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		Error("SERVER", "Template error: %v", err)
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
		log.Printf("[SERVER] Render error: %v", err)
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

// GetOpeningStats returns statistics about the opening book
func (s *Server) GetOpeningStats() map[string]int {
	if s.openingBook == nil {
		return map[string]int{
			"total_openings": 0,
			"total_nodes":    0,
			"max_depth":      0,
		}
	}
	return s.openingBook.Stats()
}

// OpeningRequest represents a request to get the opening name
type OpeningRequest struct {
	Moves []string `json:"moves"` // Move history in SAN notation
}

// OpeningResponse represents the opening information response
type OpeningResponse struct {
	ECO  string `json:"eco"`
	Name string `json:"name"`
	PGN  string `json:"pgn"`
}

// handleGetOpening returns the opening name for a given position
func (s *Server) handleGetOpening(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req OpeningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request"})
		return
	}

	// Check if opening book is loaded
	if s.openingBook == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Opening book not loaded"})
		return
	}

	// Lookup the opening
	opening := s.openingBook.Lookup(req.Moves)
	if opening == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OpeningResponse{
			ECO:  "",
			Name: "",
			PGN:  "",
		})
		return
	}

	// Return the opening info
	response := OpeningResponse{
		ECO:  opening.ECO,
		Name: opening.Name,
		PGN:  opening.PGN,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
