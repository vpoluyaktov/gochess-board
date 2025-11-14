package server

import (
	"encoding/json"
	"fmt"
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
	FEN            string            `json:"fen"`
	Moves          []string          `json:"moves"`          // Move history in UCI notation
	EnginePath     string            `json:"enginePath"`     // Path to the chess engine to use
	MoveTime       int               `json:"moveTime"`       // Time in milliseconds (0 = use clock-based)
	WhiteTime      int               `json:"whiteTime"`      // White's remaining time in milliseconds
	BlackTime      int               `json:"blackTime"`      // Black's remaining time in milliseconds
	WhiteIncrement int               `json:"whiteIncrement"` // White's increment in milliseconds
	BlackIncrement int               `json:"blackIncrement"` // Black's increment in milliseconds
	EngineOptions  map[string]string `json:"engineOptions"`  // UCI engine options (e.g., UCI_Elo, UCI_LimitStrength)
}

// MoveResponse represents the server's move response
type MoveResponse struct {
	Move      string `json:"move"`
	FEN       string `json:"fen"`
	ThinkTime int    `json:"thinkTime"` // Time engine spent thinking in milliseconds
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// handleComputerMove handles the computer move calculation (stateless)
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

	// Parse the FEN position
	fen, err := chess.FEN(req.FEN)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid FEN"})
		return
	}

	game := chess.NewGame(fen)

	// Log the request for debugging
	Info("CHESS", "Computer move request: FEN=%s, Turn=%v, Moves=%d",
		req.FEN, game.Position().Turn(), len(req.Moves))

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

	// Generate session ID for tracking
	sessionID := time.Now().Format("20060102-150405.000000")

	// Get engine name from discovered engines
	engineName := "Unknown"
	eloValue := 0
	for _, e := range s.engines {
		if e.Path == enginePath {
			engineName = e.Name
			break
		}
	}

	// Extract ELO from engine options
	if eloStr, ok := req.EngineOptions["UCI_Elo"]; ok {
		// Parse ELO as integer
		var elo int
		if _, err := fmt.Sscanf(eloStr, "%d", &elo); err == nil {
			eloValue = elo
		}
	}

	// Register engine in monitor
	activeEngine := &ActiveEngine{
		Name:           engineName,
		Path:           enginePath,
		ELO:            eloValue,
		WhiteTime:      req.WhiteTime,
		BlackTime:      req.BlackTime,
		WhiteIncrement: req.WhiteIncrement,
		BlackIncrement: req.BlackIncrement,
		StartTime:      time.Now(),
		SessionID:      sessionID,
	}
	globalMonitor.RegisterEngine(sessionID, activeEngine)
	defer globalMonitor.UnregisterEngine(sessionID)

	// Initialize chess engine
	engine, err := NewUCIEngine(enginePath)
	if err != nil {
		Error("CHESS", "Failed to initialize engine %s: %v", enginePath, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Engine initialization failed"})
		return
	}
	defer engine.Close()

	// Apply engine options if provided
	if len(req.EngineOptions) > 0 {
		for optionName, optionValue := range req.EngineOptions {
			if err := engine.SetOption(optionName, optionValue); err != nil {
				Warn("CHESS", "Failed to set option %s=%s: %v", optionName, optionValue, err)
			}
		}
	}

	// Get best move from engine (track time)
	startTime := time.Now()
	var bestMoveUCI string

	// Determine time management strategy
	if req.MoveTime > 0 {
		// Fixed time per move
		bestMoveUCI, err = engine.GetBestMove(req.FEN, time.Duration(req.MoveTime)*time.Millisecond)
	} else if req.WhiteTime > 0 || req.BlackTime > 0 {
		// Clock-based time management
		whiteTime := time.Duration(req.WhiteTime) * time.Millisecond
		blackTime := time.Duration(req.BlackTime) * time.Millisecond
		whiteInc := time.Duration(req.WhiteIncrement) * time.Millisecond
		blackInc := time.Duration(req.BlackIncrement) * time.Millisecond
		bestMoveUCI, err = engine.GetBestMoveWithClock(req.FEN, req.Moves, whiteTime, blackTime, whiteInc, blackInc)
	} else {
		// Default: 1 second per move
		bestMoveUCI, err = engine.GetBestMove(req.FEN, moveTime)
	}
	thinkTime := time.Since(startTime)

	if err != nil {
		Error("CHESS", "Failed to get best move: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to calculate move"})
		return
	}

	// Parse the UCI move (e.g., "e2e4")
	move, err := chess.UCINotation{}.Decode(game.Position(), bestMoveUCI)
	if err != nil {
		Error("CHESS", "Failed to parse move %s: %v", bestMoveUCI, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid move from engine"})
		return
	}

	// Make the move
	if err := game.Move(move); err != nil {
		Error("CHESS", "Failed to make move %s: %v", bestMoveUCI, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to make move"})
		return
	}

	// Get new FEN after move
	newFEN := game.FEN()

	Info("CHESS", "Engine move: %s, think time: %v", bestMoveUCI, thinkTime)

	// Return the move, new FEN, and think time
	response := MoveResponse{
		Move:      bestMoveUCI,
		FEN:       newFEN,
		ThinkTime: int(thinkTime.Milliseconds()),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
