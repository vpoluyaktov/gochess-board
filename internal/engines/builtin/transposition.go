package builtin

import (
	"github.com/notnil/chess"
)

// TTEntryType represents the type of transposition table entry
type TTEntryType uint8

const (
	TTExact TTEntryType = iota // Exact score
	TTAlpha                    // Upper bound (failed low)
	TTBeta                     // Lower bound (failed high)
)

// TTEntry represents a transposition table entry
type TTEntry struct {
	zobristKey uint64      // Position hash key
	depth      int         // Search depth
	score      int         // Position score
	entryType  TTEntryType // Type of score
	bestMove   *chess.Move // Best move from this position
	age        uint8       // Age for replacement strategy
}

// TranspositionTable is a hash table for storing position evaluations
type TranspositionTable struct {
	entries []TTEntry
	size    int
	age     uint8
}

// NewTranspositionTable creates a new transposition table
// Size is in MB (e.g., 64 for 64MB)
func NewTranspositionTable(sizeMB int) *TranspositionTable {
	// Calculate number of entries based on size
	// Each entry is approximately 32 bytes
	entriesPerMB := (1024 * 1024) / 32
	numEntries := sizeMB * entriesPerMB

	// Round to power of 2 for efficient modulo
	size := 1
	for size < numEntries {
		size *= 2
	}

	return &TranspositionTable{
		entries: make([]TTEntry, size),
		size:    size,
		age:     0,
	}
}

// probe looks up a position in the transposition table
func (tt *TranspositionTable) probe(zobristKey uint64, depth int, alpha, beta int) (found bool, score int, move *chess.Move) {
	if tt == nil {
		return false, 0, nil
	}

	index := int(zobristKey) & (tt.size - 1)
	entry := &tt.entries[index]

	// Check if this is the same position
	if entry.zobristKey != zobristKey {
		return false, 0, nil
	}

	// Entry found
	found = true
	move = entry.bestMove

	// Check if we can use the stored score
	if entry.depth >= depth {
		switch entry.entryType {
		case TTExact:
			score = entry.score
			return true, score, move
		case TTAlpha:
			// Upper bound - score is at most this value
			if entry.score <= alpha {
				score = alpha
				return true, score, move
			}
		case TTBeta:
			// Lower bound - score is at least this value
			if entry.score >= beta {
				score = beta
				return true, score, move
			}
		}
	}

	// Can't use score, but return the move for move ordering
	return false, 0, move
}

// store saves a position evaluation to the transposition table
func (tt *TranspositionTable) store(zobristKey uint64, depth int, score int, entryType TTEntryType, bestMove *chess.Move) {
	if tt == nil {
		return
	}

	index := int(zobristKey) & (tt.size - 1)
	entry := &tt.entries[index]

	// Replacement strategy: always replace if:
	// 1. Empty slot
	// 2. Same position (update)
	// 3. Deeper search
	// 4. Older entry
	if entry.zobristKey == 0 ||
		entry.zobristKey == zobristKey ||
		depth >= entry.depth ||
		entry.age < tt.age {
		entry.zobristKey = zobristKey
		entry.depth = depth
		entry.score = score
		entry.entryType = entryType
		entry.bestMove = bestMove
		entry.age = tt.age
	}
}

// clear resets the transposition table
func (tt *TranspositionTable) clear() {
	if tt == nil {
		return
	}
	tt.entries = make([]TTEntry, tt.size)
	tt.age = 0
}

// incrementAge increments the age counter (call at start of new search)
func (tt *TranspositionTable) incrementAge() {
	if tt == nil {
		return
	}
	tt.age++
}

// Zobrist hashing for position keys
// We'll use the chess library's built-in hash if available,
// otherwise we'll compute a simple hash from the FEN
func getZobristKey(pos *chess.Position) uint64 {
	// Simple hash based on board state
	// In a production engine, you'd use proper Zobrist hashing
	var hash uint64
	board := pos.Board()

	// Hash pieces
	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece != chess.NoPiece {
			// Simple hash combining piece type, color, and square
			pieceVal := uint64(piece.Type()) + uint64(piece.Color())*16
			hash ^= pieceVal * uint64(sq*67+1)
		}
	}

	// Hash turn
	if pos.Turn() == chess.White {
		hash ^= 0x123456789ABCDEF0
	}

	// Hash castling rights
	castleRights := pos.CastleRights()
	if castleRights.CanCastle(chess.White, chess.KingSide) {
		hash ^= 0x1111111111111111
	}
	if castleRights.CanCastle(chess.White, chess.QueenSide) {
		hash ^= 0x2222222222222222
	}
	if castleRights.CanCastle(chess.Black, chess.KingSide) {
		hash ^= 0x3333333333333333
	}
	if castleRights.CanCastle(chess.Black, chess.QueenSide) {
		hash ^= 0x4444444444444444
	}

	// Hash en passant
	if pos.EnPassantSquare() != chess.NoSquare {
		hash ^= uint64(pos.EnPassantSquare()) * 0x5555555555555555
	}

	return hash
}
