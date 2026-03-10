package builtin

import (
	"github.com/notnil/chess"
)

// Mate score constants
const (
	MateScore     = 100000 // Base mate score
	MateThreshold = 99000  // Scores above this are mate scores
)

// isInCheck returns true if the side to move is in check
// We detect this by checking if any opponent piece attacks the king
func isInCheck(pos *chess.Position) bool {
	board := pos.Board()
	turn := pos.Turn()

	// Find king position
	var kingSquare chess.Square
	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece.Type() == chess.King && piece.Color() == turn {
			kingSquare = chess.Square(sq)
			break
		}
	}

	// Check if any opponent piece attacks the king square
	opponent := chess.Black
	if turn == chess.Black {
		opponent = chess.White
	}

	return isSquareAttacked(board, kingSquare, opponent)
}

// isSquareAttacked checks if a square is attacked by any piece of the given color
func isSquareAttacked(board *chess.Board, sq chess.Square, byColor chess.Color) bool {
	sqRank := int(sq) / 8
	sqFile := int(sq) % 8

	for fromSq := 0; fromSq < 64; fromSq++ {
		piece := board.Piece(chess.Square(fromSq))
		if piece == chess.NoPiece || piece.Color() != byColor {
			continue
		}

		fromRank := fromSq / 8
		fromFile := fromSq % 8
		rankDiff := sqRank - fromRank
		fileDiff := sqFile - fromFile
		absRankDiff := rankDiff
		if absRankDiff < 0 {
			absRankDiff = -absRankDiff
		}
		absFileDiff := fileDiff
		if absFileDiff < 0 {
			absFileDiff = -absFileDiff
		}

		switch piece.Type() {
		case chess.Pawn:
			// Pawns attack diagonally
			if absFileDiff == 1 {
				if byColor == chess.White && rankDiff == 1 {
					return true
				} else if byColor == chess.Black && rankDiff == -1 {
					return true
				}
			}

		case chess.Knight:
			if (absRankDiff == 2 && absFileDiff == 1) || (absRankDiff == 1 && absFileDiff == 2) {
				return true
			}

		case chess.Bishop:
			if absRankDiff == absFileDiff && absRankDiff > 0 {
				if isDiagonalClear(board, chess.Square(fromSq), sq) {
					return true
				}
			}

		case chess.Rook:
			if (absRankDiff == 0 || absFileDiff == 0) && (absRankDiff > 0 || absFileDiff > 0) {
				if isStraightClear(board, chess.Square(fromSq), sq) {
					return true
				}
			}

		case chess.Queen:
			if absRankDiff == absFileDiff && absRankDiff > 0 {
				if isDiagonalClear(board, chess.Square(fromSq), sq) {
					return true
				}
			} else if (absRankDiff == 0 || absFileDiff == 0) && (absRankDiff > 0 || absFileDiff > 0) {
				if isStraightClear(board, chess.Square(fromSq), sq) {
					return true
				}
			}

		case chess.King:
			if absRankDiff <= 1 && absFileDiff <= 1 && (absRankDiff > 0 || absFileDiff > 0) {
				return true
			}
		}
	}

	return false
}

// isDiagonalClear checks if diagonal path is clear (for search.go)
func isDiagonalClear(board *chess.Board, from, to chess.Square) bool {
	fromRank := int(from) / 8
	fromFile := int(from) % 8
	toRank := int(to) / 8
	toFile := int(to) % 8

	rankDir := 1
	if toRank < fromRank {
		rankDir = -1
	}
	fileDir := 1
	if toFile < fromFile {
		fileDir = -1
	}

	rank := fromRank + rankDir
	file := fromFile + fileDir

	for rank != toRank && file != toFile {
		sq := chess.Square(rank*8 + file)
		if board.Piece(sq) != chess.NoPiece {
			return false
		}
		rank += rankDir
		file += fileDir
	}

	return true
}

// isStraightClear checks if straight path is clear (for search.go)
func isStraightClear(board *chess.Board, from, to chess.Square) bool {
	fromRank := int(from) / 8
	fromFile := int(from) % 8
	toRank := int(to) / 8
	toFile := int(to) % 8

	if fromRank == toRank {
		dir := 1
		if toFile < fromFile {
			dir = -1
		}
		for file := fromFile + dir; file != toFile; file += dir {
			sq := chess.Square(fromRank*8 + file)
			if board.Piece(sq) != chess.NoPiece {
				return false
			}
		}
	} else {
		dir := 1
		if toRank < fromRank {
			dir = -1
		}
		for rank := fromRank + dir; rank != toRank; rank += dir {
			sq := chess.Square(rank*8 + fromFile)
			if board.Piece(sq) != chess.NoPiece {
				return false
			}
		}
	}

	return true
}

// isEndgame returns true if the position is an endgame (few pieces)
// We avoid null move pruning in endgames due to zugzwang
func isEndgame(pos *chess.Position) bool {
	board := pos.Board()
	materialCount := 0

	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece == chess.NoPiece {
			continue
		}
		switch piece.Type() {
		case chess.Queen:
			materialCount += 9
		case chess.Rook:
			materialCount += 5
		case chess.Bishop, chess.Knight:
			materialCount += 3
		}
	}

	// Endgame if total non-pawn material is low
	return materialCount <= 13 // Roughly a rook + minor piece per side
}

// makeNullMove creates a position where the side to move passes
// Returns nil if we can't make a null move (e.g., in check)
func makeNullMove(pos *chess.Position) *chess.Position {
	// Create a new position with the turn flipped
	// We need to construct a FEN with the opposite side to move
	fen := pos.XFENString()

	// Parse and modify the FEN to flip the turn
	// FEN format: pieces turn castling en_passant halfmove fullmove
	parts := splitFEN(fen)
	if len(parts) < 4 {
		return nil
	}

	// Flip turn
	if parts[1] == "w" {
		parts[1] = "b"
	} else {
		parts[1] = "w"
	}

	// Clear en passant (null move clears it)
	parts[3] = "-"

	// Reconstruct FEN
	newFEN := parts[0] + " " + parts[1] + " " + parts[2] + " " + parts[3]
	if len(parts) > 4 {
		newFEN += " " + parts[4]
	} else {
		newFEN += " 0"
	}
	if len(parts) > 5 {
		newFEN += " " + parts[5]
	} else {
		newFEN += " 1"
	}

	fenFunc, err := chess.FEN(newFEN)
	if err != nil {
		return nil
	}

	game := chess.NewGame(fenFunc)
	return game.Position()
}

// splitFEN splits a FEN string into its components
func splitFEN(fen string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(fen); i++ {
		if fen[i] == ' ' {
			parts = append(parts, fen[start:i])
			start = i + 1
		}
	}
	if start < len(fen) {
		parts = append(parts, fen[start:])
	}
	return parts
}

// search performs minimax search with alpha-beta pruning
func (e *InternalEngine) search(pos *chess.Position, depth int, alpha, beta int) (int, *chess.Move) {
	return e.searchWithPly(pos, depth, alpha, beta, 0)
}

// searchWithPly performs the actual search with ply tracking for mate distance
func (e *InternalEngine) searchWithPly(pos *chess.Position, depth int, alpha, beta int, ply int) (int, *chess.Move) {
	// Get zobrist key for this position
	zobristKey := getZobristKey(pos)
	alphaOrig := alpha

	// Probe transposition table
	if found, score, ttMove := e.tt.probe(zobristKey, depth, alpha, beta); found {
		// Adjust mate scores for ply
		if score > MateThreshold {
			score -= ply
		} else if score < -MateThreshold {
			score += ply
		}
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
		if pos.Status() == chess.Checkmate {
			// Return mate score adjusted by ply (prefer shorter mates)
			return -MateScore + ply, nil
		}
		return 0, nil // Stalemate
	}

	// Razoring: if static eval is far below alpha at depth 1, drop to quiescence
	if depth == 1 && !isInCheck(pos) && alpha > -MateThreshold && alpha < MateThreshold {
		staticEval := e.evaluate(pos)
		razoringMargin := 500
		if staticEval+razoringMargin < alpha {
			qScore := e.quiescence(pos, alpha, beta, 4)
			if qScore < alpha {
				return qScore, nil
			}
		}
	}

	// Null Move Pruning: if we can skip our turn and still have a good position,
	// the current position is likely very good (beta cutoff)
	// Don't do null move if:
	// - We're in check
	// - Depth is too shallow
	// - We just did a null move (handled by not passing null move flag)
	if depth >= 3 && !isInCheck(pos) && !isEndgame(pos) {
		// Make null move (skip our turn)
		nullPos := makeNullMove(pos)
		if nullPos != nil {
			// Search with reduced depth (R=2 or R=3)
			R := 2
			if depth >= 6 {
				R = 3
			}
			nullScore, _ := e.searchWithPly(nullPos, depth-1-R, -beta, -beta+1, ply+1)
			nullScore = -nullScore

			// If null move search fails high, we can prune
			if nullScore >= beta {
				// Verification search at reduced depth to avoid zugzwang issues
				if depth >= 6 {
					verifyScore, _ := e.searchWithPly(pos, depth-R, beta-1, beta, ply)
					if verifyScore >= beta {
						return beta, nil
					}
				} else {
					return beta, nil
				}
			}
		}
	}

	// Order moves: captures first, then others
	// If we have a TT move, search it first
	var ttMove *chess.Move
	if _, _, move := e.tt.probe(zobristKey, 0, alpha, beta); move != nil {
		ttMove = move
	}

	// Internal Iterative Deepening (IID): if no TT move at sufficient depth,
	// do a reduced search to find a good move to search first
	if ttMove == nil && depth >= 4 {
		iidDepth := depth - 2
		if iidDepth < 1 {
			iidDepth = 1
		}
		_, iidMove := e.searchWithPly(pos, iidDepth, alpha, beta, ply)
		if iidMove != nil {
			ttMove = iidMove
		}
	}

	e.orderMovesWithPos(moves, pos, ttMove, ply)

	var bestMove *chess.Move

	// Futility pruning margins (indexed by depth)
	futilityMargins := [4]int{0, 300, 500, 900}

	// Calculate static eval for futility pruning (only if not in check)
	staticEval := 0
	canFutilityPrune := depth <= 3 && !isInCheck(pos)
	if canFutilityPrune {
		staticEval = e.evaluate(pos)
	}

	// Singular extension: DISABLED - caused tactical regression
	// The overhead of searching all moves at reduced depth hurts more than it helps

	for i, move := range moves {
		// Create new position with the move
		newPos := pos.Update(move)

		var score int
		// Check extension: extend search when giving check
		extension := 0
		if move.HasTag(chess.Check) {
			extension = 1
		}

		// Futility pruning: DISABLED - causes tactical regression
		_ = canFutilityPrune
		_ = futilityMargins
		_ = staticEval

		// Late Move Reductions (LMR): reduce depth for later quiet moves
		reduction := 0
		if i >= 4 && depth >= 3 && !move.HasTag(chess.Capture) && !move.HasTag(chess.Check) && extension == 0 {
			reduction = 1
			if i >= 8 {
				reduction = 2
			}
		}

		// Search with potential reduction
		searchDepth := depth - 1 + extension - reduction
		if searchDepth < 0 {
			searchDepth = 0
		}

		// Principal Variation Search (PVS)
		if i == 0 {
			// First move: search with full window
			score, _ = e.searchWithPly(newPos, searchDepth, -beta, -alpha, ply+1)
			score = -score
		} else {
			// Other moves: search with null window first
			score, _ = e.searchWithPly(newPos, searchDepth, -alpha-1, -alpha, ply+1)
			score = -score

			// If it beats alpha, re-search with full window
			if score > alpha && score < beta {
				score, _ = e.searchWithPly(newPos, searchDepth, -beta, -alpha, ply+1)
				score = -score
			}
		}

		// Re-search at full depth if reduced search found a better move
		if reduction > 0 && score > alpha {
			score, _ = e.searchWithPly(newPos, depth-1+extension, -beta, -alpha, ply+1)
			score = -score
		}

		if score >= beta {
			// Beta cutoff - store lower bound
			// Store killer move and update history for quiet moves
			if !move.HasTag(chess.Capture) {
				if e.killerMoves != nil {
					e.killerMoves.add(move, ply)
				}
				if e.history != nil {
					e.history.update(move, pos.Turn(), depth)
					// Penalize quiet moves that were searched before this one
					for j := 0; j < i; j++ {
						if !moves[j].HasTag(chess.Capture) {
							e.history.updateBadMove(moves[j], pos.Turn(), depth)
						}
					}
				}
			}
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

	// Delta pruning: if we're so far behind that even capturing the queen won't help, prune
	const deltaMargin = 975 // Queen value + some margin
	if standPat+deltaMargin < alpha {
		return alpha
	}

	// Only search captures
	moves := pos.ValidMoves()

	// Order captures by SEE value (best captures first)
	captures := make([]*chess.Move, 0, len(moves))
	for _, move := range moves {
		if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
			captures = append(captures, move)
		}
	}

	// Sort captures by SEE (descending)
	e.orderCapturesBySEE(pos, captures)

	for _, move := range captures {
		// SEE pruning disabled - hurts tactical performance
		// SEE is still used for move ordering above
		_ = e.see(pos, move)

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

	return alpha
}

// orderCapturesBySEE orders captures by their SEE value (best first)
func (e *InternalEngine) orderCapturesBySEE(pos *chess.Position, moves []*chess.Move) {
	// Simple insertion sort (captures list is usually small)
	for i := 1; i < len(moves); i++ {
		key := moves[i]
		keyScore := e.see(pos, key)
		j := i - 1
		for j >= 0 && e.see(pos, moves[j]) < keyScore {
			moves[j+1] = moves[j]
			j--
		}
		moves[j+1] = key
	}
}
