package builtin

import "github.com/notnil/chess"

// search performs minimax search with alpha-beta pruning
func (e *InternalEngine) search(pos *chess.Position, depth int, alpha, beta int) (int, *chess.Move) {
	// Base case: depth reached
	if depth == 0 {
		score := e.quiescence(pos, alpha, beta, 4)
		return score, nil
	}

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		// Checkmate or stalemate
		return e.evaluate(pos), nil
	}

	// Order moves: captures first, then others
	e.orderMoves(moves)

	var bestMove *chess.Move

	for _, move := range moves {
		// Create new position with the move
		newPos := pos.Update(move)
		score, _ := e.search(newPos, depth-1, -beta, -alpha)
		score = -score

		if score >= beta {
			// Beta cutoff
			return beta, move
		}

		if score > alpha {
			alpha = score
			bestMove = move
		}
	}

	return alpha, bestMove
}

// quiescence performs a quiescence search (only captures)
func (e *InternalEngine) quiescence(pos *chess.Position, alpha, beta int, depth int) int {
	// Limit quiescence depth to prevent infinite loops
	if depth <= 0 {
		return e.evaluate(pos)
	}

	standPat := e.evaluate(pos)

	if standPat >= beta {
		return beta
	}
	if alpha < standPat {
		alpha = standPat
	}

	// Only search captures
	moves := pos.ValidMoves()
	for _, move := range moves {
		// Only consider captures
		if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
			// Create new position with the move
			newPos := pos.Update(move)
			score := -e.quiescence(newPos, -beta, -alpha, depth-1)

			if score >= beta {
				return beta
			}
			if score > alpha {
				alpha = score
			}
		}
	}

	return alpha
}
