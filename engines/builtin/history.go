package builtin

import "github.com/notnil/chess"

// HistoryTable tracks the success of quiet moves across the search
// Indexed by [color][from_square][to_square]
type HistoryTable struct {
	table [2][64][64]int // [color][from][to]
	max   int            // Maximum value for scaling
}

// NewHistoryTable creates a new history table
func NewHistoryTable() *HistoryTable {
	return &HistoryTable{
		max: 1, // Start at 1 to avoid division by zero
	}
}

// update adds a bonus to a move that caused a beta cutoff
// depth is used to weight the bonus (deeper cutoffs are more valuable)
func (h *HistoryTable) update(move *chess.Move, color chess.Color, depth int) {
	if move == nil {
		return
	}

	from := int(move.S1())
	to := int(move.S2())
	colorIdx := 0
	if color == chess.Black {
		colorIdx = 1
	}

	// Bonus increases with depth squared (deeper cutoffs are more significant)
	bonus := depth * depth

	h.table[colorIdx][from][to] += bonus

	// Track maximum for scaling
	if h.table[colorIdx][from][to] > h.max {
		h.max = h.table[colorIdx][from][to]
	}

	// Age/scale down if values get too large
	if h.max > 10000 {
		h.scale()
	}
}

// getScore returns the history score for a move
func (h *HistoryTable) getScore(move *chess.Move, color chess.Color) int {
	if move == nil {
		return 0
	}

	from := int(move.S1())
	to := int(move.S2())
	colorIdx := 0
	if color == chess.Black {
		colorIdx = 1
	}

	return h.table[colorIdx][from][to]
}

// scale reduces all values to prevent overflow
func (h *HistoryTable) scale() {
	for c := 0; c < 2; c++ {
		for f := 0; f < 64; f++ {
			for t := 0; t < 64; t++ {
				h.table[c][f][t] /= 2
			}
		}
	}
	h.max /= 2
}

// clear resets the history table
func (h *HistoryTable) clear() {
	for c := 0; c < 2; c++ {
		for f := 0; f < 64; f++ {
			for t := 0; t < 64; t++ {
				h.table[c][f][t] = 0
			}
		}
	}
	h.max = 1
}

// updateBadMove penalizes moves that didn't cause a cutoff
// Called for moves searched before the cutoff move
func (h *HistoryTable) updateBadMove(move *chess.Move, color chess.Color, depth int) {
	if move == nil {
		return
	}

	from := int(move.S1())
	to := int(move.S2())
	colorIdx := 0
	if color == chess.Black {
		colorIdx = 1
	}

	// Penalty is smaller than bonus
	penalty := depth

	h.table[colorIdx][from][to] -= penalty
	if h.table[colorIdx][from][to] < 0 {
		h.table[colorIdx][from][to] = 0
	}
}
