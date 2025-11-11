package server

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed assets/*
var assetsFS embed.FS

// Server represents the HTTP server
type Server struct {
	addr string
}

// New creates a new server instance
func New(addr string) *Server {
	return &Server{addr: addr}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Serve static assets
	http.Handle("/assets/", http.FileServer(http.FS(assetsFS)))
	
	// API endpoints
	http.HandleFunc("/api/computer-move", s.handleComputerMove)
	
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
	
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Render error: %v", err)
	}
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.addr
}
