package builtin

import "github.com/notnil/chess"

// MVV-LVA (Most Valuable Victim - Least Valuable Attacker) values
// Used for ordering captures
var mvvLvaScores = [7][7]int{
	// Victim: None, Pawn, Knight, Bishop, Rook, Queen, King
	{0, 0, 0, 0, 0, 0, 0},             // Attacker: None
	{0, 105, 205, 305, 405, 505, 605}, // Attacker: Pawn
	{0, 104, 204, 304, 404, 504, 604}, // Attacker: Knight
	{0, 103, 203, 303, 403, 503, 603}, // Attacker: Bishop
	{0, 102, 202, 302, 402, 502, 602}, // Attacker: Rook
	{0, 101, 201, 301, 401, 501, 601}, // Attacker: Queen
	{0, 100, 200, 300, 400, 500, 600}, // Attacker: King
}

// orderMoves orders moves to improve alpha-beta pruning
// Uses MVV-LVA for captures, killer moves for quiet moves, TT move gets highest priority
// ttMove is the move from the transposition table (searched first if not nil)
// ply is the current search depth (for killer move lookup)
func (e *InternalEngine) orderMoves(moves []*chess.Move, ttMove *chess.Move, ply int) {
	// Get the current position to evaluate captures
	// Note: We need the position before the move to get victim piece
	// This is a simple bubble sort - could be optimized with a better algorithm
	for i := 0; i < len(moves); i++ {
		for j := i + 1; j < len(moves); j++ {
			if e.moveOrderScore(moves[j], ttMove, ply) > e.moveOrderScore(moves[i], ttMove, ply) {
				moves[i], moves[j] = moves[j], moves[i]
			}
		}
	}
}

// moveOrderScore assigns a score for move ordering
func (e *InternalEngine) moveOrderScore(move *chess.Move, ttMove *chess.Move, ply int) int {
	score := 0

	// Highest priority: TT move
	if ttMove != nil && move.String() == ttMove.String() {
		score += 100000
	}

	// MVV-LVA for captures
	if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
		// For now, use a simplified heuristic
		// All captures get a high base score
		// TODO: Enhance with actual piece type detection for true MVV-LVA
		score += 10000

		// Bonus for captures with promotion
		if move.Promo() != chess.NoPieceType {
			score += 5000 // Promotions are very valuable
		}
	}

	// Killer moves (for quiet moves)
	if e.killerMoves != nil {
		score += e.killerMoves.getKillerScore(move, ply)
	}

	// Prioritize checks
	if move.HasTag(chess.Check) {
		score += 50
	}

	// Promotions
	if move.Promo() != chess.NoPieceType {
		score += 8000
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
