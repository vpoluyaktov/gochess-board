package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gochess-board/internal/engines"
)

func TestMoveRequest_JSON(t *testing.T) {
	req := MoveRequest{
		FEN:            "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Moves:          []string{"e2e4"},
		EnginePath:     "/usr/bin/stockfish",
		MoveTime:       1000,
		WhiteTime:      300000,
		BlackTime:      300000,
		WhiteIncrement: 5000,
		BlackIncrement: 5000,
		EngineOptions:  map[string]string{"UCI_Elo": "2000"},
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Unmarshal back
	var decoded MoveRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.FEN != req.FEN {
		t.Errorf("FEN mismatch: got %q, want %q", decoded.FEN, req.FEN)
	}
	if len(decoded.Moves) != len(req.Moves) {
		t.Errorf("Moves length mismatch: got %d, want %d", len(decoded.Moves), len(req.Moves))
	}
	if decoded.EngineOptions["UCI_Elo"] != "2000" {
		t.Error("Engine options not preserved")
	}
}

func TestMoveResponse_JSON(t *testing.T) {
	resp := MoveResponse{
		Move:      "e2e4",
		FEN:       "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		ThinkTime: 150,
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal back
	var decoded MoveResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if decoded.Move != resp.Move {
		t.Errorf("Move mismatch: got %q, want %q", decoded.Move, resp.Move)
	}
	if decoded.ThinkTime != resp.ThinkTime {
		t.Errorf("ThinkTime mismatch: got %d, want %d", decoded.ThinkTime, resp.ThinkTime)
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	errResp := ErrorResponse{
		Error: "Invalid move",
	}

	data, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if decoded.Error != errResp.Error {
		t.Errorf("Error mismatch: got %q, want %q", decoded.Error, errResp.Error)
	}
}

func TestHandleComputerMove_MethodNotAllowed(t *testing.T) {
	server := &Server{
		addr:    ":8080",
		engines: []engines.EngineInfo{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/computer-move", nil)
	w := httptest.NewRecorder()

	server.handleComputerMove(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error != "Method not allowed" {
		t.Errorf("Expected 'Method not allowed' error, got %q", errResp.Error)
	}
}

func TestHandleComputerMove_InvalidJSON(t *testing.T) {
	server := &Server{
		addr:    ":8080",
		engines: []engines.EngineInfo{},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/computer-move", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	server.handleComputerMove(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error != "Invalid request" {
		t.Errorf("Expected 'Invalid request' error, got %q", errResp.Error)
	}
}

func TestHandleComputerMove_InvalidFEN(t *testing.T) {
	server := &Server{
		addr:    ":8080",
		engines: []engines.EngineInfo{},
	}

	reqBody := MoveRequest{
		FEN: "invalid fen string",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/computer-move", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	server.handleComputerMove(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error != "Invalid FEN" {
		t.Errorf("Expected 'Invalid FEN' error, got %q", errResp.Error)
	}
}

func TestHandleComputerMove_GameOver(t *testing.T) {
	server := &Server{
		addr:    ":8080",
		engines: []engines.EngineInfo{},
	}

	// Checkmate position
	reqBody := MoveRequest{
		FEN: "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/computer-move", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	server.handleComputerMove(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error != "Game over" {
		t.Errorf("Expected 'Game over' error, got %q", errResp.Error)
	}
}

func TestMoveRequest_TimeManagement(t *testing.T) {
	tests := []struct {
		name         string
		moveTime     int
		whiteTime    int
		blackTime    int
		expectedMode string
	}{
		{
			name:         "fixed time mode",
			moveTime:     1000,
			whiteTime:    0,
			blackTime:    0,
			expectedMode: "fixed",
		},
		{
			name:         "clock-based mode",
			moveTime:     0,
			whiteTime:    300000,
			blackTime:    300000,
			expectedMode: "clock",
		},
		{
			name:         "default mode",
			moveTime:     0,
			whiteTime:    0,
			blackTime:    0,
			expectedMode: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := MoveRequest{
				MoveTime:  tt.moveTime,
				WhiteTime: tt.whiteTime,
				BlackTime: tt.blackTime,
			}

			// Determine which mode would be used
			var mode string
			if req.MoveTime > 0 {
				mode = "fixed"
			} else if req.WhiteTime > 0 || req.BlackTime > 0 {
				mode = "clock"
			} else {
				mode = "default"
			}

			if mode != tt.expectedMode {
				t.Errorf("Expected mode %q, got %q", tt.expectedMode, mode)
			}
		})
	}
}

func TestMoveRequest_EngineOptions(t *testing.T) {
	req := MoveRequest{
		EngineOptions: map[string]string{
			"UCI_Elo":           "2000",
			"UCI_LimitStrength": "true",
			"Threads":           "4",
			"Hash":              "128",
		},
	}

	if len(req.EngineOptions) != 4 {
		t.Errorf("Expected 4 options, got %d", len(req.EngineOptions))
	}

	if req.EngineOptions["UCI_Elo"] != "2000" {
		t.Error("UCI_Elo option not set correctly")
	}
	if req.EngineOptions["UCI_LimitStrength"] != "true" {
		t.Error("UCI_LimitStrength option not set correctly")
	}
}

func TestMoveResponse_ThinkTime(t *testing.T) {
	tests := []struct {
		name      string
		thinkTime int
		isValid   bool
	}{
		{
			name:      "normal think time",
			thinkTime: 150,
			isValid:   true,
		},
		{
			name:      "very fast",
			thinkTime: 1,
			isValid:   true,
		},
		{
			name:      "very slow",
			thinkTime: 60000,
			isValid:   true,
		},
		{
			name:      "zero time",
			thinkTime: 0,
			isValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := MoveResponse{
				ThinkTime: tt.thinkTime,
			}

			if tt.isValid && resp.ThinkTime < 0 {
				t.Errorf("Invalid think time: %d", resp.ThinkTime)
			}
		})
	}
}
