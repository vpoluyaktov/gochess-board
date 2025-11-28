package builtin

import "github.com/notnil/chess"

// search performs minimax search with alpha-beta pruning
func (e *InternalEngine) search(pos *chess.Position, depth int, alpha, beta int) (int, *chess.Move) {
	// Get zobrist key for this position
	zobristKey := getZobristKey(pos)
	alphaOrig := alpha

	// Probe transposition table
	if found, score, ttMove := e.tt.probe(zobristKey, depth, alpha, beta); found {
		return score, ttMove
	}

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
	// If we have a TT move, search it first
	var ttMove *chess.Move
	if _, _, move := e.tt.probe(zobristKey, 0, alpha, beta); move != nil {
		ttMove = move
	}
	// Calculate ply from depth (for killer moves)
	ply := 0 // We don't track ply in basic search, use 0
	e.orderMoves(moves, ttMove, ply)

	var bestMove *chess.Move

	for _, move := range moves {
		// Create new position with the move
		newPos := pos.Update(move)
		score, _ := e.search(newPos, depth-1, -beta, -alpha)
		score = -score

		if score >= beta {
			// Beta cutoff - store lower bound
			e.tt.store(zobristKey, depth, beta, TTBeta, move)
			return beta, move
		}

		if score > alpha {
			alpha = score
			bestMove = move
		}
	}

	// Store result in transposition table
	if alpha > alphaOrig {
		// Exact score
		e.tt.store(zobristKey, depth, alpha, TTExact, bestMove)
	} else {
		// Upper bound (failed low)
		e.tt.store(zobristKey, depth, alpha, TTAlpha, bestMove)
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
