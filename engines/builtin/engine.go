package builtin

import (
	"fmt"
	"math"
	"time"

	"github.com/notnil/chess"
)

// InternalEngine is a simple built-in chess engine
type InternalEngine struct {
	name    string
	stopped bool // Flag to stop analysis
}

// NewEngine creates a new built-in chess engine
func NewEngine() *InternalEngine {
	return &InternalEngine{
		name: "GoChess",
	}
}

// NewInternalEngine is deprecated, use NewEngine instead
func NewInternalEngine() *InternalEngine {
	return NewEngine()
}

// GetBestMove implements the ChessEngine interface
func (e *InternalEngine) GetBestMove(fen string, moveTime time.Duration) (string, error) {
	// Parse FEN
	fenFunc, err := chess.FEN(fen)
	if err != nil {
		return "", fmt.Errorf("invalid FEN: %w", err)
	}

	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Use iterative deepening with time limit
	var bestMove *chess.Move
	startTime := time.Now()

	// Search depths 1-5 (adjust based on time available)
	maxDepth := 4
	if moveTime > 5*time.Second {
		maxDepth = 5
	} else if moveTime < 1*time.Second {
		maxDepth = 3
	}

	for depth := 1; depth <= maxDepth; depth++ {
		// Check if we're running out of time
		if time.Since(startTime) > moveTime*8/10 {
			break
		}

		_, move := e.search(pos, depth, -math.MaxInt32, math.MaxInt32)
		if move != nil {
			bestMove = move
		}
	}

	if bestMove == nil {
		// Fallback: return first legal move
		moves := pos.ValidMoves()
		if len(moves) > 0 {
			bestMove = moves[0]
		} else {
			return "", fmt.Errorf("no legal moves available")
		}
	}

	return bestMove.String(), nil
}

// GetBestMoveWithClock implements the ChessEngine interface
func (e *InternalEngine) GetBestMoveWithClock(fen string, moveHistory []string, whiteTime, blackTime, whiteInc, blackInc time.Duration) (string, error) {
	// Simple time management: use a fraction of remaining time
	var timeToUse time.Duration

	fenFunc, err := chess.FEN(fen)
	if err != nil {
		return "", fmt.Errorf("invalid FEN: %w", err)
	}
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	if pos.Turn() == chess.White {
		timeToUse = whiteTime / 20
		if whiteInc > 0 {
			timeToUse += whiteInc / 2
		}
	} else {
		timeToUse = blackTime / 20
		if blackInc > 0 {
			timeToUse += blackInc / 2
		}
	}

	// Minimum 100ms, maximum 10s
	if timeToUse < 100*time.Millisecond {
		timeToUse = 100 * time.Millisecond
	}
	if timeToUse > 10*time.Second {
		timeToUse = 10 * time.Second
	}

	return e.GetBestMove(fen, timeToUse)
}

// SetOption implements the ChessEngine interface
func (e *InternalEngine) SetOption(name, value string) error {
	// Internal engine doesn't have configurable options yet
	return nil
}

// Close implements the ChessEngine interface
func (e *InternalEngine) Close() error {
	// Nothing to clean up
	return nil
}
