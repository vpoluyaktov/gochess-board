package builtin

import "github.com/notnil/chess"

// orderMoves orders moves to improve alpha-beta pruning
// Captures and checks are searched first
func (e *InternalEngine) orderMoves(moves []*chess.Move) {
	// Simple ordering: captures first
	// This is a basic implementation - could be improved with MVV-LVA, killer moves, etc.
	for i := 0; i < len(moves); i++ {
		for j := i + 1; j < len(moves); j++ {
			if e.moveOrderScore(moves[j]) > e.moveOrderScore(moves[i]) {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}
}

// moveOrderScore assigns a score for move ordering
func (e *InternalEngine) moveOrderScore(move *chess.Move) int {
	score := 0

	// Prioritize captures
	if move.HasTag(chess.Capture) {
		score += 10
	}

	// Prioritize checks
	if move.HasTag(chess.Check) {
		score += 5
	}

	return score
}
