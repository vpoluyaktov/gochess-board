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
	FEN string `json:"fen"`
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

// handleComputerMove handles the computer move calculation
func (s *Server) handleComputerMove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Track request
	gameState := GetGameState()
	gameState.IncrementRequests()

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

	// Initialize Stockfish engine
	engine, err := NewUCIEngine(stockfishPath)
	if err != nil {
		log.Printf("Failed to initialize Stockfish: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Engine initialization failed"})
		return
	}
	defer engine.Close()

	// Get best move from Stockfish (track time)
	startTime := time.Now()
	bestMoveUCI, err := engine.GetBestMove(req.FEN, moveTime)
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
