package builtin

import "github.com/notnil/chess"

// KillerMoves stores killer moves for each ply
// Killer moves are quiet moves that caused beta cutoffs
type KillerMoves struct {
	// Two killer moves per ply (primary and secondary)
	moves [maxDepth][2]*chess.Move
}

const maxDepth = 64 // Maximum search depth

// NewKillerMoves creates a new killer moves table
func NewKillerMoves() *KillerMoves {
	return &KillerMoves{}
}

// add adds a killer move at the given ply
func (km *KillerMoves) add(move *chess.Move, ply int) {
	if ply < 0 || ply >= maxDepth {
		return
	}

	// Don't store captures as killer moves (they're already prioritized)
	if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
		return
	}

	// If this move is already the primary killer, don't change anything
	if km.moves[ply][0] != nil && km.moves[ply][0].String() == move.String() {
		return
	}

	// Shift primary to secondary, new move becomes primary
	km.moves[ply][1] = km.moves[ply][0]
	km.moves[ply][0] = move
}

// isKiller checks if a move is a killer move at the given ply
func (km *KillerMoves) isKiller(move *chess.Move, ply int) bool {
	if ply < 0 || ply >= maxDepth {
		return false
	}

	if km.moves[ply][0] != nil && km.moves[ply][0].String() == move.String() {
		return true
	}

	if km.moves[ply][1] != nil && km.moves[ply][1].String() == move.String() {
		return true
	}

	return false
}

// clear resets all killer moves
func (km *KillerMoves) clear() {
	km.moves = [maxDepth][2]*chess.Move{}
}

// getKillerScore returns the score bonus for a killer move
func (km *KillerMoves) getKillerScore(move *chess.Move, ply int) int {
	if ply < 0 || ply >= maxDepth {
		return 0
	}

	// Primary killer gets higher score
	if km.moves[ply][0] != nil && km.moves[ply][0].String() == move.String() {
		return 9000
	}

	// Secondary killer gets lower score
	if km.moves[ply][1] != nil && km.moves[ply][1].String() == move.String() {
		return 8000
	}

	return 0
}
