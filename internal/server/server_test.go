package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gochess-board/internal/engines"
	"gochess-board/internal/opening"
)

func TestServer_GetEngines(t *testing.T) {
	engineList := []engines.EngineInfo{
		{Name: "Stockfish", Path: "/usr/bin/stockfish", Type: "uci"},
		{Name: "Fruit", Path: "/usr/bin/fruit", Type: "uci"},
	}

	server := &Server{
		addr:    ":8080",
		engines: engineList,
	}

	result := server.GetEngines()

	if len(result) != 2 {
		t.Errorf("Expected 2 engines, got %d", len(result))
	}
	if result[0].Name != "Stockfish" {
		t.Errorf("Expected first engine to be Stockfish, got %s", result[0].Name)
	}
}

func TestServer_GetAddr(t *testing.T) {
	server := &Server{
		addr: ":8080",
	}

	addr := server.GetAddr()
	if addr != ":8080" {
		t.Errorf("Expected addr ':8080', got %q", addr)
	}
}

func TestServer_GetOpeningStats(t *testing.T) {
	t.Run("with opening book", func(t *testing.T) {
		book := opening.NewOpeningBook()
		server := &Server{
			addr:        ":8080",
			openingBook: book,
		}

		stats := server.GetOpeningStats()
		if stats == nil {
			t.Error("Expected stats map, got nil")
		}
		if _, ok := stats["total_openings"]; !ok {
			t.Error("Expected 'total_openings' in stats")
		}
	})

	t.Run("without opening book", func(t *testing.T) {
		server := &Server{
			addr:        ":8080",
			openingBook: nil,
		}

		stats := server.GetOpeningStats()
		if stats == nil {
			t.Error("Expected stats map, got nil")
		}
		if stats["total_openings"] != 0 {
			t.Errorf("Expected 0 openings, got %d", stats["total_openings"])
		}
	})
}

func TestHandleGetEngines(t *testing.T) {
	engineList := []engines.EngineInfo{
		{Name: "Stockfish", Path: "/usr/bin/stockfish", Type: "uci"},
	}

	server := &Server{
		addr:    ":8080",
		engines: engineList,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/engines", nil)
	w := httptest.NewRecorder()

	server.handleGetEngines(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %q", contentType)
	}

	var result []engines.EngineInfo
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 engine, got %d", len(result))
	}
	if result[0].Name != "Stockfish" {
		t.Errorf("Expected engine name 'Stockfish', got %q", result[0].Name)
	}
}

func TestHandleGetEngines_EmptyList(t *testing.T) {
	server := &Server{
		addr:    ":8080",
		engines: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/engines", nil)
	w := httptest.NewRecorder()

	server.handleGetEngines(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []engines.EngineInfo
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 engines, got %d", len(result))
	}
}

func TestHandleGetOpening_MethodNotAllowed(t *testing.T) {
	server := &Server{
		addr:        ":8080",
		openingBook: opening.NewOpeningBook(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/opening", nil)
	w := httptest.NewRecorder()

	server.handleGetOpening(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleGetOpening_InvalidRequest(t *testing.T) {
	server := &Server{
		addr:        ":8080",
		openingBook: opening.NewOpeningBook(),
	}

	req := httptest.NewRequest(http.MethodPost, "/api/opening", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	server.handleGetOpening(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleGetOpening_NoBook(t *testing.T) {
	server := &Server{
		addr:        ":8080",
		openingBook: nil,
	}

	reqBody := OpeningRequest{Moves: []string{"e4"}}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/opening", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	server.handleGetOpening(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestHandleGetOpening_NotFound(t *testing.T) {
	server := &Server{
		addr:        ":8080",
		openingBook: opening.NewOpeningBook(),
	}

	reqBody := OpeningRequest{Moves: []string{"a4", "b5", "c6"}}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/opening", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	server.handleGetOpening(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response OpeningResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ECO != "" || response.Name != "" {
		t.Error("Expected empty response for unknown opening")
	}
}

func TestMoveRequest_Structure(t *testing.T) {
	req := MoveRequest{
		FEN:            "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Moves:          []string{"e2e4", "e7e5"},
		EnginePath:     "/usr/bin/stockfish",
		MoveTime:       1000,
		WhiteTime:      300000,
		BlackTime:      300000,
		WhiteIncrement: 5000,
		BlackIncrement: 5000,
		EngineOptions:  map[string]string{"UCI_Elo": "2000"},
	}

	if req.FEN == "" {
		t.Error("FEN should not be empty")
	}
	if len(req.Moves) != 2 {
		t.Errorf("Expected 2 moves, got %d", len(req.Moves))
	}
	if req.EngineOptions["UCI_Elo"] != "2000" {
		t.Error("Expected UCI_Elo option")
	}
}

func TestMoveResponse_Structure(t *testing.T) {
	resp := MoveResponse{
		Move:      "e2e4",
		FEN:       "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		ThinkTime: 150,
	}

	if resp.Move != "e2e4" {
		t.Errorf("Expected move 'e2e4', got %q", resp.Move)
	}
	if resp.ThinkTime != 150 {
		t.Errorf("Expected think time 150, got %d", resp.ThinkTime)
	}
}

func TestErrorResponse_Structure(t *testing.T) {
	err := ErrorResponse{Error: "Test error"}
	if err.Error != "Test error" {
		t.Errorf("Expected error 'Test error', got %q", err.Error)
	}
}

func TestOpeningRequest_Structure(t *testing.T) {
	req := OpeningRequest{
		Moves: []string{"e4", "e5", "Nf3"},
	}

	if len(req.Moves) != 3 {
		t.Errorf("Expected 3 moves, got %d", len(req.Moves))
	}
}

func TestOpeningResponse_Structure(t *testing.T) {
	resp := OpeningResponse{
		ECO:  "C50",
		Name: "Italian Game",
		PGN:  "1. e4 e5 2. Nf3 Nc6 3. Bc4",
	}

	if resp.ECO != "C50" {
		t.Errorf("Expected ECO 'C50', got %q", resp.ECO)
	}
	if resp.Name != "Italian Game" {
		t.Errorf("Expected name 'Italian Game', got %q", resp.Name)
	}
}
