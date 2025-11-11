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
}

var globalGameState = &GameState{
	GameStarted: time.Now(),
}

// GetGameState returns the global game state
func GetGameState() *GameState {
	return globalGameState
}

// UpdateMove updates the game state with a new move (Stockfish/Black)
func (gs *GameState) UpdateMove(move string, fen string, thinkTime time.Duration) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	gs.MovesPlayed++
	gs.LastMove = move
	gs.LastMoveTime = time.Now()
	gs.StockfishTime = thinkTime
	gs.CurrentFEN = fen
	gs.BlackMoves++
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
	gs.TotalRequests++
	// Each request means the player made a move (white)
	gs.WhiteMoves++
}

// GetStats returns a snapshot of current stats
func (gs *GameState) GetStats() (int, string, time.Time, time.Duration, string, int, int, int) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	return gs.MovesPlayed, gs.LastMove, gs.LastMoveTime, gs.StockfishTime, 
	       gs.CurrentFEN, gs.TotalRequests, gs.WhiteMoves, gs.BlackMoves
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
}
