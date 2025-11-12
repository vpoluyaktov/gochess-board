package server

import (
	"sync"
	"time"
)

// GameState tracks the current game statistics
type GameState struct {
	mu              sync.RWMutex
	MovesPlayed     int
	LastMove        string
	LastMoveTime    time.Time
	StockfishTime   time.Duration
	CurrentFEN      string
	GameStarted     time.Time
	TotalRequests   int
	WhiteMoves      int
	BlackMoves      int
	WhitePlayTime   time.Duration
	BlackPlayTime   time.Duration
	CurrentTurnStart time.Time
	IsWhiteTurn     bool
}

var globalGameState = &GameState{
	GameStarted:      time.Now(),
	CurrentTurnStart: time.Now(),
	IsWhiteTurn:      true,
}

// GetGameState returns the global game state
func GetGameState() *GameState {
	return globalGameState
}

// UpdateMove updates the game state with a new move (Stockfish/Black)
func (gs *GameState) UpdateMove(move string, fen string, thinkTime time.Duration) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	// Add elapsed time to black's total
	if !gs.IsWhiteTurn {
		gs.BlackPlayTime += time.Since(gs.CurrentTurnStart)
	}
	
	gs.MovesPlayed++
	gs.LastMove = move
	gs.LastMoveTime = time.Now()
	gs.StockfishTime = thinkTime
	gs.CurrentFEN = fen
	gs.BlackMoves++
	
	// Switch to white's turn
	gs.IsWhiteTurn = true
	gs.CurrentTurnStart = time.Now()
}

// UpdatePlayerMove updates the game state with a player move (White)
func (gs *GameState) UpdatePlayerMove() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.WhiteMoves++
}

// IncrementRequests increments the total request counter
func (gs *GameState) IncrementRequests() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	// Add elapsed time to white's total
	if gs.IsWhiteTurn {
		gs.WhitePlayTime += time.Since(gs.CurrentTurnStart)
	}
	
	gs.TotalRequests++
	// Each request means the player made a move (white)
	gs.WhiteMoves++
	
	// Switch to black's turn
	gs.IsWhiteTurn = false
	gs.CurrentTurnStart = time.Now()
}

// GetStats returns a snapshot of current stats
func (gs *GameState) GetStats() (int, string, time.Time, time.Duration, string, int, int, int, time.Duration, time.Duration) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	// Calculate current play times including active turn
	whiteTime := gs.WhitePlayTime
	blackTime := gs.BlackPlayTime
	
	if gs.IsWhiteTurn {
		whiteTime += time.Since(gs.CurrentTurnStart)
	} else {
		blackTime += time.Since(gs.CurrentTurnStart)
	}
	
	return gs.MovesPlayed, gs.LastMove, gs.LastMoveTime, gs.StockfishTime, 
	       gs.CurrentFEN, gs.TotalRequests, gs.WhiteMoves, gs.BlackMoves, whiteTime, blackTime
}

// Reset resets the game state
func (gs *GameState) Reset() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	gs.MovesPlayed = 0
	gs.LastMove = ""
	gs.LastMoveTime = time.Time{}
	gs.StockfishTime = 0
	gs.CurrentFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	gs.GameStarted = time.Now()
	gs.WhiteMoves = 0
	gs.BlackMoves = 0
	gs.WhitePlayTime = 0
	gs.BlackPlayTime = 0
	gs.CurrentTurnStart = time.Now()
	gs.IsWhiteTurn = true
}
