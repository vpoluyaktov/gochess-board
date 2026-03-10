package builtin

import "github.com/notnil/chess"

// CounterMoveTable tracks moves that refute the opponent's previous move
// Indexed by [piece_type][to_square] of the previous move
type CounterMoveTable struct {
	// For each opponent move (piece moving to square), store the best counter move
	table [7][64]*chess.Move // [piece_type][to_square]
}

// NewCounterMoveTable creates a new counter move table
func NewCounterMoveTable() *CounterMoveTable {
	return &CounterMoveTable{}
}

// update stores a counter move that caused a beta cutoff
func (cm *CounterMoveTable) update(prevMove, counterMove *chess.Move, prevPieceType chess.PieceType) {
	if prevMove == nil || counterMove == nil {
		return
	}

	// Don't store captures as counter moves (they're already prioritized)
	if counterMove.HasTag(chess.Capture) {
		return
	}

	toSq := int(prevMove.S2())
	pieceIdx := int(prevPieceType)

	if pieceIdx >= 0 && pieceIdx < 7 && toSq >= 0 && toSq < 64 {
		cm.table[pieceIdx][toSq] = counterMove
	}
}

// get returns the counter move for a given previous move
func (cm *CounterMoveTable) get(prevMove *chess.Move, prevPieceType chess.PieceType) *chess.Move {
	if prevMove == nil {
		return nil
	}

	toSq := int(prevMove.S2())
	pieceIdx := int(prevPieceType)

	if pieceIdx >= 0 && pieceIdx < 7 && toSq >= 0 && toSq < 64 {
		return cm.table[pieceIdx][toSq]
	}

	return nil
}

// isCounterMove checks if a move is the counter move for the previous move
func (cm *CounterMoveTable) isCounterMove(move, prevMove *chess.Move, prevPieceType chess.PieceType) bool {
	counter := cm.get(prevMove, prevPieceType)
	if counter == nil || move == nil {
		return false
	}
	return move.String() == counter.String()
}

// clear resets the counter move table
func (cm *CounterMoveTable) clear() {
	for p := 0; p < 7; p++ {
		for s := 0; s < 64; s++ {
			cm.table[p][s] = nil
		}
	}
}
