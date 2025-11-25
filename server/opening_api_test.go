package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-chess/engine"
	"go-chess/opening"
)

func TestOpeningAPIEndpoint(t *testing.T) {
	// Create a server with opening book
	srv := &Server{
		addr:        ":8080",
		engines:     []engine.EngineInfo{},
		openingBook: opening.NewOpeningBook(),
	}

	// Load opening book
	if err := srv.openingBook.LoadFromDirectory("assets/openings"); err != nil {
		t.Fatalf("Failed to load opening book: %v", err)
	}

	tests := []struct {
		name        string
		moves       []string
		wantECO     string
		wantName    string
		expectEmpty bool
	}{
		{
			name:     "Italian Game",
			moves:    []string{"e4", "e5", "Nf3", "Nc6", "Bc4"},
			wantECO:  "C50",
			wantName: "Italian Game",
		},
		{
			name:     "Sicilian Defense",
			moves:    []string{"e4", "c5"},
			wantECO:  "B20",
			wantName: "Sicilian Defense",
		},
		{
			name:     "French Defense",
			moves:    []string{"e4", "e6"},
			wantECO:  "C00",
			wantName: "French Defense",
		},
		{
			name:     "Anderssen's Opening",
			moves:    []string{"a3"},
			wantECO:  "A00",
			wantName: "Anderssen's Opening",
		},
		{
			name:     "Evans Gambit - minimal",
			moves:    []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Bc5", "b4"},
			wantECO:  "C51",
			wantName: "Italian Game: Evans Gambit",
		},
		{
			name:     "Evans Gambit Accepted",
			moves:    []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Bc5", "b4", "Bxb4"},
			wantECO:  "C51",
			wantName: "Italian Game: Evans Gambit Accepted",
		},
		{
			name:     "Evans Gambit - MacKenzie.pgn full sequence (16 moves, finds deepest match)",
			moves:    []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Bc5", "b4", "Bxb4", "c3", "Bc5", "O-O", "d6", "d4", "exd4", "cxd4", "Bb6"},
			wantECO:  "C51",
			wantName: "Italian Game: Evans Gambit, McDonnell Defense",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBody := OpeningRequest{
				Moves: tt.moves,
			}
			body, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/opening", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler
			srv.handleGetOpening(w, req)

			// Check response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var resp OpeningResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if !tt.expectEmpty {
				if resp.ECO != tt.wantECO {
					t.Errorf("ECO = %s, want %s", resp.ECO, tt.wantECO)
				}
				if resp.Name != tt.wantName {
					t.Errorf("Name = %s, want %s", resp.Name, tt.wantName)
				}
			}

			t.Logf("Response: %s (%s) - %s", resp.Name, resp.ECO, resp.PGN)
		})
	}
}

func TestOpeningAPIInvalidRequest(t *testing.T) {
	srv := &Server{
		addr:        ":8080",
		engines:     []engine.EngineInfo{},
		openingBook: opening.NewOpeningBook(),
	}

	// Test with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/opening", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleGetOpening(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestOpeningAPIMethodNotAllowed(t *testing.T) {
	srv := &Server{
		addr:        ":8080",
		engines:     []engine.EngineInfo{},
		openingBook: opening.NewOpeningBook(),
	}

	// Test with GET method
	req := httptest.NewRequest(http.MethodGet, "/api/opening", nil)
	w := httptest.NewRecorder()

	srv.handleGetOpening(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestOpeningAPINoBook(t *testing.T) {
	srv := &Server{
		addr:        ":8080",
		engines:     []engine.EngineInfo{},
		openingBook: nil, // No opening book
	}

	reqBody := OpeningRequest{
		Moves: []string{"e4", "e5"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/opening", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleGetOpening(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}
