package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/notnil/chess"
)

const (
	stockfishPath = "/usr/games/stockfish"
	moveTime      = 1000 * time.Millisecond // 1 second per move
)

// MoveRequest represents a move request from the client
type MoveRequest struct {
	FEN        string `json:"fen"`
	EnginePath string `json:"enginePath"` // Path to the chess engine to use
	LastMove   string `json:"lastMove"`   // The player's last move in UCI notation (e.g., "e2e4")
}

// MoveResponse represents the server's move response
type MoveResponse struct {
	Move string `json:"move"`
	FEN  string `json:"fen"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatsResponse represents game statistics
type StatsResponse struct {
	WhiteMoves    int    `json:"whiteMoves"`
	BlackMoves    int    `json:"blackMoves"`
	TotalMoves    int    `json:"totalMoves"`
	WhiteTime     string `json:"whiteTime"`
	BlackTime     string `json:"blackTime"`
	GameDuration  string `json:"gameDuration"`
}

// ClockRequest represents a request to set time control
type ClockRequest struct {
	InitialMinutes   int `json:"initialMinutes"`
	IncrementSeconds int `json:"incrementSeconds"`
}

// ClockResponse represents the current clock state
type ClockResponse struct {
	WhiteTimeLeft int  `json:"whiteTimeLeft"` // milliseconds
	BlackTimeLeft int  `json:"blackTimeLeft"` // milliseconds
	IsWhiteTurn   bool `json:"isWhiteTurn"`
	ClockRunning  bool `json:"clockRunning"`
}

// handleComputerMove handles the computer move calculation
func (s *Server) handleComputerMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request"})
		return
	}

	// Track player's move in game state
	gameState := GetGameState()
	if req.LastMove != "" {
		gameState.IncrementRequests(req.LastMove)
	}

	// Parse the FEN position
	fen, err := chess.FEN(req.FEN)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid FEN"})
		return
	}

	game := chess.NewGame(fen)

	// Check if game is over
	if game.Outcome() != chess.NoOutcome {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Game over"})
		return
	}

	// Determine which engine to use
	enginePath := req.EnginePath
	if enginePath == "" {
		// Default to stockfish if no engine specified
		enginePath = stockfishPath
	}

	// Initialize chess engine
	engine, err := NewUCIEngine(enginePath)
	if err != nil {
		log.Printf("Failed to initialize engine %s: %v", enginePath, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Engine initialization failed"})
		return
	}
	defer engine.Close()

	// Get clock times and move history from game state
	whiteTime, blackTime, _ := gameState.GetClockTimes()
	whiteInc := gameState.TimeControl.Increment
	blackInc := gameState.TimeControl.Increment
	moveHistory := gameState.GetMoveHistory()
	
	// Get best move from engine (track time)
	startTime := time.Now()
	var bestMoveUCI string
	
	// Use clock-based time management if time control is active (not unlimited)
	if gameState.TimeControl.InitialTime > 0 {
		bestMoveUCI, err = engine.GetBestMoveWithClock(req.FEN, moveHistory, whiteTime, blackTime, whiteInc, blackInc)
	} else {
		// Fallback to fixed time for unlimited games
		bestMoveUCI, err = engine.GetBestMove(req.FEN, moveTime)
	}
	thinkTime := time.Since(startTime)
	
	if err != nil {
		log.Printf("Failed to get best move: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to calculate move"})
		return
	}

	// Parse the UCI move (e.g., "e2e4")
	move, err := chess.UCINotation{}.Decode(game.Position(), bestMoveUCI)
	if err != nil {
		log.Printf("Failed to parse move %s: %v", bestMoveUCI, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid move from engine"})
		return
	}

	// Make the move
	if err := game.Move(move); err != nil {
		log.Printf("Failed to make move %s: %v", bestMoveUCI, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to make move"})
		return
	}

	// Update game state
	newFEN := game.FEN()
	gameState.UpdateMove(bestMoveUCI, newFEN, thinkTime)

	// Return the move and new FEN
	response := MoveResponse{
		Move: bestMoveUCI,
		FEN:  newFEN,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleStats returns current game statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameState := GetGameState()
	_, _, _, _, _, _, whiteMoves, blackMoves, whiteTime, blackTime := gameState.GetStats()
	
	gameDuration := time.Since(gameState.GameStarted)
	
	response := StatsResponse{
		WhiteMoves:   whiteMoves,
		BlackMoves:   blackMoves,
		TotalMoves:   whiteMoves + blackMoves,
		WhiteTime:    whiteTime.Round(time.Second).String(),
		BlackTime:    blackTime.Round(time.Second).String(),
		GameDuration: gameDuration.Round(time.Second).String(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleSetTimeControl sets the time control for the game
func (s *Server) handleSetTimeControl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req ClockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request"})
		return
	}

	gameState := GetGameState()
	gameState.SetTimeControl(req.InitialMinutes, req.IncrementSeconds)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleGetClock returns the current clock state
func (s *Server) handleGetClock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameState := GetGameState()
	whiteTime, blackTime, isWhiteTurn := gameState.GetClockTimes()

	response := ClockResponse{
		WhiteTimeLeft: int(whiteTime.Milliseconds()),
		BlackTimeLeft: int(blackTime.Milliseconds()),
		IsWhiteTurn:   isWhiteTurn,
		ClockRunning:  gameState.ClockRunning,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleStartClock starts the chess clock
func (s *Server) handleStartClock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	gameState := GetGameState()
	gameState.StartClock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleGetMoveHistory returns the move history
func (s *Server) handleGetMoveHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gameState := GetGameState()
	history := gameState.GetMoveHistoryDisplay()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}
