package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/notnil/chess"

	"gochess-board/engines"
	"gochess-board/engines/builtin"
	"gochess-board/logger"
)

const (
	moveTime = 1000 * time.Millisecond // 1 second per move (used for unlimited mode and default)
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
	IsUnlimited    bool              `json:"isUnlimited"`    // True for unlimited time (0+0) mode
	EngineOptions  map[string]string `json:"engineOptions"`  // UCI engine options (e.g., UCI_Elo, UCI_LimitStrength)
	GameID         string            `json:"gameId"`         // Game identifier for engine pooling (optional)
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
	logger.Info("CHESS", "Computer move request: FEN=%s, Turn=%v, Moves=%d",
		req.FEN, game.Position().Turn(), len(req.Moves))

	// Check if game is over
	if game.Outcome() != chess.NoOutcome {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Game over"})
		return
	}

	// Check opening book first (if available)
	if s.polyglotBook != nil {
		logger.Debug("POLYGLOT_BOOK", "Probing book for position: %s", req.FEN)
		bookMove := s.polyglotBook.ProbeWeighted(game.Position())
		if bookMove != "" {
			logger.Info("POLYGLOT_BOOK", "Book move found: %s", bookMove)

			// Parse and apply the book move
			move, err := chess.UCINotation{}.Decode(game.Position(), bookMove)
			if err != nil {
				logger.Warn("POLYGLOT_BOOK", "Failed to parse book move %s: %v", bookMove, err)
				// Continue to engine if book move is invalid
			} else {
				if err := game.Move(move); err != nil {
					logger.Warn("POLYGLOT_BOOK", "Failed to make book move %s: %v", bookMove, err)
					// Continue to engine if book move fails
				} else {
					// Book move successful - return immediately
					newFEN := game.FEN()
					logger.Info("POLYGLOT_BOOK", "Book move applied: %s", bookMove)

					response := MoveResponse{
						Move:      bookMove,
						FEN:       newFEN,
						ThinkTime: 0, // Book moves are instant
					}

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
					return
				}
			}
		} else {
			logger.Debug("POLYGLOT_BOOK", "No book move found for position")
		}
	}

	// Determine which engine to use
	enginePath := req.EnginePath
	if enginePath == "" {
		// This should never happen as client always sends enginePath from discovered engines
		logger.Error("CHESS", "No engine path specified in request")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "No engine specified"})
		return
	}

	// Generate session ID for tracking
	sessionID := time.Now().Format("20060102-150405.000000")

	// Get engine name and type from discovered engines
	engineName := "Unknown"
	engineType := "uci" // default to UCI
	eloValue := 0
	for _, e := range s.engines {
		if e.Path == enginePath {
			engineName = e.Name
			engineType = e.Type
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
	activeEngine := &engines.ActiveEngine{
		Name:           engineName,
		Path:           enginePath,
		Type:           engines.EngineTypeMove,
		ELO:            eloValue,
		WhiteTime:      req.WhiteTime,
		BlackTime:      req.BlackTime,
		WhiteIncrement: req.WhiteIncrement,
		BlackIncrement: req.BlackIncrement,
		StartTime:      time.Now(),
		SessionID:      sessionID,
		GameID:         req.GameID,
	}
	engines.GlobalMonitor.RegisterEngine(sessionID, activeEngine)
	defer engines.GlobalMonitor.UnregisterEngine(sessionID)

	// Initialize chess engine based on mode (persistent pool or per-request)
	var chessEngine engines.ChessEngine
	var usePooledEngine bool

	if engines.GlobalEnginePool != nil && req.GameID != "" {
		// Persistent engine mode - use engine pool
		chessEngine, err = engines.GlobalEnginePool.GetOrCreateEngine(
			req.GameID, enginePath, engineName, engineType, req.EngineOptions)
		if err != nil {
			logger.Error("CHESS", "Failed to get engine from pool: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Engine initialization failed"})
			return
		}
		usePooledEngine = true
		// Don't close pooled engines - they are managed by the pool
	} else {
		// Per-request engine mode (original behavior)
		if engineType == "internal" || enginePath == "internal" {
			// Use built-in internal engine
			chessEngine = builtin.NewInternalEngine()
			logger.Info("CHESS", "Using built-in internal engine")
		} else if engineType == "cecp" {
			chessEngine, err = engines.NewCECPEngine(enginePath, engineName)
			if err != nil {
				logger.Error("CHESS", "Failed to initialize CECP engine %s: %v", engineName, err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Engine initialization failed"})
				return
			}
		} else {
			chessEngine, err = engines.NewUCIEngine(enginePath, engineName)
			if err != nil {
				logger.Error("CHESS", "Failed to initialize UCI engine %s: %v", engineName, err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Engine initialization failed"})
				return
			}
		}
		defer chessEngine.Close()

		// Apply engine options if provided (only for non-pooled engines)
		if len(req.EngineOptions) > 0 {
			for optionName, optionValue := range req.EngineOptions {
				if err := chessEngine.SetOption(optionName, optionValue); err != nil {
					logger.Warn("CHESS", "Failed to set option %s=%s: %v", optionName, optionValue, err)
				}
			}
		}
	}

	// For pooled engines, release back to pool after use
	if usePooledEngine {
		defer engines.GlobalEnginePool.ReleaseEngine(req.GameID, enginePath)
	}

	// Get best move from engine (track time)
	startTime := time.Now()
	var bestMoveUCI string

	// Determine time management strategy
	if req.MoveTime > 0 {
		// Fixed time per move
		bestMoveUCI, err = chessEngine.GetBestMove(req.FEN, time.Duration(req.MoveTime)*time.Millisecond)
	} else if req.IsUnlimited {
		// Unlimited time mode (0+0): use fixed time per move
		bestMoveUCI, err = chessEngine.GetBestMove(req.FEN, moveTime)
	} else if req.WhiteTime > 0 || req.BlackTime > 0 {
		// Clock-based time management (timed games)
		whiteTime := time.Duration(req.WhiteTime) * time.Millisecond
		blackTime := time.Duration(req.BlackTime) * time.Millisecond
		whiteInc := time.Duration(req.WhiteIncrement) * time.Millisecond
		blackInc := time.Duration(req.BlackIncrement) * time.Millisecond
		bestMoveUCI, err = chessEngine.GetBestMoveWithClock(req.FEN, req.Moves, whiteTime, blackTime, whiteInc, blackInc)
	} else {
		// Default: 1 second per move
		bestMoveUCI, err = chessEngine.GetBestMove(req.FEN, moveTime)
	}
	thinkTime := time.Since(startTime)

	if err != nil {
		logger.Error("CHESS", "Failed to get best move: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to calculate move"})
		return
	}

	// Parse the move - try UCI notation first, then SAN notation (for CECP engines)
	move, err := chess.UCINotation{}.Decode(game.Position(), bestMoveUCI)
	if err != nil {
		// Try parsing as SAN notation (e.g., "Nf6" from CECP engines like Crafty)
		move, err = chess.AlgebraicNotation{}.Decode(game.Position(), bestMoveUCI)
		if err != nil {
			logger.Error("CHESS", "Failed to parse move %s as UCI or SAN: %v", bestMoveUCI, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid move from engine"})
			return
		}
		// Convert SAN move to UCI notation for consistency
		bestMoveUCI = chess.UCINotation{}.Encode(game.Position(), move)
		logger.Debug("CHESS", "Converted SAN to UCI: %s", bestMoveUCI)
	}

	// Make the move
	if err := game.Move(move); err != nil {
		logger.Error("CHESS", "Failed to make move %s: %v", bestMoveUCI, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to make move"})
		return
	}

	// Get new FEN after move
	newFEN := game.FEN()

	logger.Info("CHESS", "Engine move: %s, think time: %v", bestMoveUCI, thinkTime)

	// Return the move in UCI notation, new FEN, and think time
	response := MoveResponse{
		Move:      bestMoveUCI,
		FEN:       newFEN,
		ThinkTime: int(thinkTime.Milliseconds()),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
