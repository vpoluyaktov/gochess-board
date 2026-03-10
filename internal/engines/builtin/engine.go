package builtin

import (
	"fmt"
	"math"
	"time"

	"github.com/notnil/chess"
)

// InternalEngine is a simple built-in chess engine
type InternalEngine struct {
	name         string
	stopped      bool                // Flag to stop analysis
	tt           *TranspositionTable // Transposition table for position caching
	killerMoves  *KillerMoves        // Killer moves heuristic
	history      *HistoryTable       // History heuristic for move ordering
	counterMoves *CounterMoveTable   // Counter move heuristic
}

// NewEngine creates a new built-in chess engine
func NewEngine() *InternalEngine {
	return &InternalEngine{
		name:         "GoChess",
		tt:           NewTranspositionTable(64), // 64MB transposition table
		killerMoves:  NewKillerMoves(),
		history:      NewHistoryTable(),
		counterMoves: NewCounterMoveTable(),
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
	var bestScore int
	startTime := time.Now()

	// Clear killer moves for new search
	if e.killerMoves != nil {
		e.killerMoves.clear()
	}

	// Increment TT age for new search
	if e.tt != nil {
		e.tt.incrementAge()
	}

	// Search depths 1-7 (adjust based on time available)
	// With LMR, null move pruning, and other optimizations, we can search deeper
	maxDepth := 6
	if moveTime > 5*time.Second {
		maxDepth = 7
	} else if moveTime > 2*time.Second {
		maxDepth = 6
	} else if moveTime < 500*time.Millisecond {
		maxDepth = 5
	}

	// Aspiration window parameters
	const aspirationWindow = 50 // Initial window size in centipawns
	prevScore := 0

	for depth := 1; depth <= maxDepth; depth++ {
		// Check if we're running out of time
		elapsed := time.Since(startTime)
		if elapsed > moveTime*7/10 {
			break
		}

		var score int
		var move *chess.Move

		// Use aspiration windows for depth >= 4
		if depth >= 4 && bestMove != nil {
			// Start with a narrow window around the previous score
			alpha := prevScore - aspirationWindow
			beta := prevScore + aspirationWindow

			score, move = e.search(pos, depth, alpha, beta)

			// If the search failed low or high, re-search with wider window
			if score <= alpha {
				// Failed low - re-search with wider lower bound
				score, move = e.search(pos, depth, -math.MaxInt32, beta)
			} else if score >= beta {
				// Failed high - re-search with wider upper bound
				score, move = e.search(pos, depth, alpha, math.MaxInt32)
			}

			// If still outside bounds, do a full-width search
			if score <= alpha || score >= beta {
				score, move = e.search(pos, depth, -math.MaxInt32, math.MaxInt32)
			}
		} else {
			// Full-width search for shallow depths
			score, move = e.search(pos, depth, -math.MaxInt32, math.MaxInt32)
		}

		if move != nil {
			bestMove = move
			bestScore = score
			prevScore = score

			// If we found a forced mate, stop searching
			if score > MateThreshold || score < -MateThreshold {
				break
			}
		}
	}

	// Suppress unused variable warning
	_ = bestScore

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
