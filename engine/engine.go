package engine

import "time"

// ChessEngine is an interface that both UCI and CECP engines implement
type ChessEngine interface {
	// GetBestMove gets the best move for a position with fixed time
	GetBestMove(fen string, moveTime time.Duration) (string, error)

	// GetBestMoveWithClock gets the best move using chess clock time management
	GetBestMoveWithClock(fen string, moveHistory []string, whiteTime, blackTime, whiteInc, blackInc time.Duration) (string, error)

	// SetOption sets an engine option
	SetOption(name, value string) error

	// Close closes the engine
	Close() error
}
