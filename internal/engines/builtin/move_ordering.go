package builtin

import "github.com/notnil/chess"

// MVV-LVA (Most Valuable Victim - Least Valuable Attacker) values
// Used for ordering captures - higher score = search first
// Formula: 10 * victim_value - attacker_value (captures more valuable pieces with less valuable attackers first)
var mvvLvaScores = [7][7]int{
	// Victim: None, Pawn, Knight, Bishop, Rook, Queen, King
	{0, 0, 0, 0, 0, 0, 0},             // Attacker: None
	{0, 105, 205, 305, 405, 505, 605}, // Attacker: Pawn (best attacker)
	{0, 104, 204, 304, 404, 504, 604}, // Attacker: Knight
	{0, 103, 203, 303, 403, 503, 603}, // Attacker: Bishop
	{0, 102, 202, 302, 402, 502, 602}, // Attacker: Rook
	{0, 101, 201, 301, 401, 501, 601}, // Attacker: Queen
	{0, 100, 200, 300, 400, 500, 600}, // Attacker: King (worst attacker)
}

// Move ordering score constants
const (
	ttMoveBonus         = 100000 // TT move - highest priority
	winningCaptureBonus = 15000  // Winning captures (SEE > 0)
	equalCaptureBonus   = 14000  // Equal captures (SEE = 0)
	killerMove1Bonus    = 9000   // Primary killer move
	killerMove2Bonus    = 8000   // Secondary killer move
	counterMoveBonus    = 7500   // Counter move heuristic
	promotionBonus      = 12000  // Promotions
	checkBonus          = 7000   // Checking moves
	losingCaptureBonus  = 1000   // Losing captures (still search, but last)
)

// orderMoves orders moves to improve alpha-beta pruning
// Uses MVV-LVA for captures, killer moves for quiet moves, TT move gets highest priority
// ttMove is the move from the transposition table (searched first if not nil)
// ply is the current search depth (for killer move lookup)
func (e *InternalEngine) orderMoves(moves []*chess.Move, ttMove *chess.Move, ply int) {
	// Cache scores to avoid recalculating
	scores := make([]int, len(moves))
	for i, move := range moves {
		scores[i] = e.moveOrderScore(move, ttMove, ply)
	}

	// Selection sort (simple and efficient for small arrays)
	for i := 0; i < len(moves)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(moves); j++ {
			if scores[j] > scores[maxIdx] {
				maxIdx = j
			}
		}
		if maxIdx != i {
			moves[i], moves[maxIdx] = moves[maxIdx], moves[i]
			scores[i], scores[maxIdx] = scores[maxIdx], scores[i]
		}
	}
}

// orderMovesWithPos orders moves using position info for better MVV-LVA
func (e *InternalEngine) orderMovesWithPos(moves []*chess.Move, pos *chess.Position, ttMove *chess.Move, ply int) {
	e.orderMovesWithPrevMove(moves, pos, ttMove, ply, nil, chess.NoPieceType)
}

// orderMovesWithPrevMove orders moves with counter move heuristic
func (e *InternalEngine) orderMovesWithPrevMove(moves []*chess.Move, pos *chess.Position, ttMove *chess.Move, ply int, prevMove *chess.Move, prevPieceType chess.PieceType) {
	// Cache scores to avoid recalculating
	scores := make([]int, len(moves))
	for i, move := range moves {
		scores[i] = e.moveOrderScoreWithPrevMove(move, pos, ttMove, ply, prevMove, prevPieceType)
	}

	// Selection sort
	for i := 0; i < len(moves)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(moves); j++ {
			if scores[j] > scores[maxIdx] {
				maxIdx = j
			}
		}
		if maxIdx != i {
			moves[i], moves[maxIdx] = moves[maxIdx], moves[i]
			scores[i], scores[maxIdx] = scores[maxIdx], scores[i]
		}
	}
}

// moveOrderScore assigns a score for move ordering (without position)
func (e *InternalEngine) moveOrderScore(move *chess.Move, ttMove *chess.Move, ply int) int {
	score := 0

	// Highest priority: TT move
	if ttMove != nil && move.String() == ttMove.String() {
		return ttMoveBonus
	}

	// Promotions are very valuable
	if move.Promo() != chess.NoPieceType {
		score += promotionBonus
		// Queen promotion is best
		if move.Promo() == chess.Queen {
			score += 1000
		}
	}

	// Captures - use basic MVV-LVA heuristic
	if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
		score += winningCaptureBonus // Assume winning until proven otherwise
	}

	// Killer moves (for quiet moves only)
	if !move.HasTag(chess.Capture) && e.killerMoves != nil {
		killerScore := e.killerMoves.getKillerScore(move, ply)
		if killerScore > 0 {
			score += killerScore
		}
	}

	// Checking moves get bonus
	if move.HasTag(chess.Check) {
		score += checkBonus
	}

	return score
}

// moveOrderScoreWithPos assigns a score using position info for true MVV-LVA
func (e *InternalEngine) moveOrderScoreWithPos(move *chess.Move, pos *chess.Position, ttMove *chess.Move, ply int) int {
	return e.moveOrderScoreWithPrevMove(move, pos, ttMove, ply, nil, chess.NoPieceType)
}

// moveOrderScoreWithPrevMove assigns a score with counter move support
func (e *InternalEngine) moveOrderScoreWithPrevMove(move *chess.Move, pos *chess.Position, ttMove *chess.Move, ply int, prevMove *chess.Move, prevPieceType chess.PieceType) int {
	score := 0

	// Highest priority: TT move
	if ttMove != nil && move.String() == ttMove.String() {
		return ttMoveBonus
	}

	// Promotions are very valuable
	if move.Promo() != chess.NoPieceType {
		score += promotionBonus
		if move.Promo() == chess.Queen {
			score += 1000
		}
	}

	// Captures - use true MVV-LVA with position info
	if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
		board := pos.Board()
		attacker := board.Piece(move.S1())
		victim := board.Piece(move.S2())

		// Handle en passant
		if move.HasTag(chess.EnPassant) {
			victim = chess.Piece(chess.Pawn)
			if pos.Turn() == chess.White {
				victim = chess.BlackPawn
			} else {
				victim = chess.WhitePawn
			}
		}

		if attacker != chess.NoPiece && victim != chess.NoPiece {
			// True MVV-LVA score
			mvvLva := getMVVLVAScore(attacker.Type(), victim.Type())
			score += 10000 + mvvLva
		} else {
			score += winningCaptureBonus
		}
	} else {
		// Quiet moves: use history heuristic
		if e.history != nil {
			historyScore := e.history.getScore(move, pos.Turn())
			// Scale history score to be below killer moves but above nothing
			score += historyScore / 10
		}

		// Counter move heuristic
		if e.counterMoves != nil && prevMove != nil {
			if e.counterMoves.isCounterMove(move, prevMove, prevPieceType) {
				score += counterMoveBonus
			}
		}
	}

	// Killer moves (for quiet moves only)
	if !move.HasTag(chess.Capture) && e.killerMoves != nil {
		killerScore := e.killerMoves.getKillerScore(move, ply)
		if killerScore > 0 {
			score += killerScore
		}
	}

	// Checking moves get bonus
	if move.HasTag(chess.Check) {
		score += checkBonus
	}

	return score
}

// getMVVLVAScore returns the MVV-LVA score for a capture
// This requires knowing both the attacker and victim piece types
func getMVVLVAScore(attacker, victim chess.PieceType) int {
	if attacker == chess.NoPieceType || victim == chess.NoPieceType {
		return 0
	}
	return mvvLvaScores[attacker][victim]
}
