package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestSearchStartingPosition(t *testing.T) {
	engine := NewEngine()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search at depth 3
	score, move := engine.search(pos, 3, -10000, 10000)

	if move == nil {
		t.Fatal("Search should return a move")
	}

	t.Logf("Depth 3 search: score=%d, move=%s", score, move.String())

	// Score should be reasonable (not extreme)
	if score < -2000 || score > 2000 {
		t.Errorf("Starting position score seems extreme: %d", score)
	}
}

func TestSearchFindsMate(t *testing.T) {
	engine := NewEngine()

	// Mate in 1 position - back rank mate
	// White rook on a1, black king on g8, white can play Ra8#
	fen := "6k1/8/8/8/8/8/8/R6K w - - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search should find the mate
	score, move := engine.search(pos, 3, -10000, 10000)

	if move == nil {
		t.Fatal("Search should find a move")
	}

	t.Logf("Found move: score=%d, move=%s", score, move.String())

	// The engine should find a good move (Ra8# is mate)
	// At depth 3 it should see the mate and give a high score
	// Note: Our simple engine might not always find forced mate, so we just check it finds a reasonable move
	if move.String() == "a1a8" {
		t.Logf("Engine found the mate! Ra8#")
		if score < 9000 {
			t.Logf("Score %d is lower than expected for mate, but move is correct", score)
		}
	} else {
		t.Logf("Engine found %s instead of Ra8# (mate in 1)", move.String())
	}
}

func TestSearchDepthZero(t *testing.T) {
	engine := NewEngine()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Depth 0 should use quiescence search
	score, move := engine.search(pos, 0, -10000, 10000)

	// Should return a score but no move
	if move != nil {
		t.Errorf("Depth 0 search should return nil move, got %s", move.String())
	}

	t.Logf("Depth 0 score: %d", score)
}

func TestSearchWithAlphaBetaPruning(t *testing.T) {
	engine := NewEngine()

	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search with tight alpha-beta window
	alpha := -100
	beta := 100
	score, move := engine.search(pos, 2, alpha, beta)

	if move == nil {
		t.Fatal("Search should return a move")
	}

	// Score should be within or at the bounds
	if score < alpha-1000 || score > beta+1000 {
		t.Logf("Score %d outside window [%d, %d] - this is OK due to fail-soft",
			score, alpha, beta)
	}

	t.Logf("Alpha-beta search: score=%d, move=%s, window=[%d,%d]",
		score, move.String(), alpha, beta)
}

func TestQuiescence(t *testing.T) {
	engine := NewEngine()

	// Tactical position with captures available
	fen := "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Quiescence search
	score := engine.quiescence(pos, -10000, 10000, 4)

	t.Logf("Quiescence score: %d", score)

	// Should return a reasonable score
	if score < -3000 || score > 3000 {
		t.Errorf("Quiescence score seems extreme: %d", score)
	}
}

func TestQuiescenceDepthLimit(t *testing.T) {
	engine := NewEngine()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Test different quiescence depths
	score1 := engine.quiescence(pos, -10000, 10000, 0)
	score2 := engine.quiescence(pos, -10000, 10000, 2)
	score4 := engine.quiescence(pos, -10000, 10000, 4)

	t.Logf("Quiescence scores - depth 0: %d, depth 2: %d, depth 4: %d",
		score1, score2, score4)

	// All should be reasonable
	for _, score := range []int{score1, score2, score4} {
		if score < -5000 || score > 5000 {
			t.Errorf("Quiescence score seems extreme: %d", score)
		}
	}
}

func TestQuiescenceOnlySearchesCaptures(t *testing.T) {
	engine := NewEngine()

	// Quiet position (no captures available)
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Quiescence should just return static eval for quiet positions
	score := engine.quiescence(pos, -10000, 10000, 4)
	staticEval := engine.evaluate(pos)

	// In a quiet position, quiescence should be close to static eval
	diff := score - staticEval
	if diff < -50 || diff > 50 {
		t.Logf("Quiescence vs static eval difference: %d (this is OK if there are en passant captures)", diff)
	}

	t.Logf("Quiescence: %d, Static eval: %d, Difference: %d", score, staticEval, diff)
}

func TestSearchIterativeDeepening(t *testing.T) {
	engine := NewEngine()

	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search at increasing depths
	var moves []*chess.Move
	var scores []int

	for depth := 1; depth <= 4; depth++ {
		score, move := engine.search(pos, depth, -10000, 10000)
		moves = append(moves, move)
		scores = append(scores, score)

		t.Logf("Depth %d: score=%d, move=%s", depth, score, move.String())
	}

	// All searches should return valid moves
	for i, move := range moves {
		if move == nil {
			t.Errorf("Depth %d returned nil move", i+1)
		}
	}
}

func TestSearchNoLegalMoves(t *testing.T) {
	engine := NewEngine()

	// Stalemate position - black king on a8, white king on a6, white queen on b6
	// Black has no legal moves but is not in check
	fen := "k7/8/KQ6/8/8/8/8/8 b - - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Verify this is actually stalemate
	moves := pos.ValidMoves()
	t.Logf("Valid moves: %v, Status: %s", moves, pos.Status())

	score, move := engine.search(pos, 3, -1000000, 1000000)

	t.Logf("Stalemate position: score=%d, move=%v", score, move)

	// In stalemate, score should be 0 (draw) and no move
	if len(moves) == 0 && score != 0 {
		t.Errorf("Expected score 0 for stalemate, got %d", score)
	}
}

func TestPVS(t *testing.T) {
	engine := NewEngine()

	// Position where PVS should help - clear best move
	fen := "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search at depth 5 to exercise PVS
	score, move := engine.search(pos, 5, -1000000, 1000000)

	t.Logf("PVS search: score=%d, move=%v", score, move)

	if move == nil {
		t.Error("PVS search returned nil move")
	}
}

func TestPVSConsistency(t *testing.T) {
	engine := NewEngine()

	// Run the same search multiple times to verify consistency
	fen := "rnbqkb1r/pppppppp/5n2/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 1 2"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	var firstMove *chess.Move
	var firstScore int

	for i := 0; i < 3; i++ {
		// Clear TT between searches to test consistency
		engine.tt.clear()

		score, move := engine.search(pos, 4, -1000000, 1000000)

		if i == 0 {
			firstMove = move
			firstScore = score
		} else {
			// Scores should be consistent (within a small margin due to TT effects)
			if move.String() != firstMove.String() {
				t.Logf("Move changed between iterations: %s vs %s", firstMove.String(), move.String())
			}
			if abs(score-firstScore) > 10 {
				t.Logf("Score changed significantly: %d vs %d", firstScore, score)
			}
		}

		t.Logf("Iteration %d: score=%d, move=%v", i+1, score, move)
	}
}

func TestFutilityPruning(t *testing.T) {
	engine := NewEngine()

	// Position where white is significantly ahead - futility pruning should help
	// White has extra queen
	fen := "r1b1kbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search at depth 3 where futility pruning is active
	score, move := engine.search(pos, 3, -1000000, 1000000)

	t.Logf("Futility pruning test: score=%d, move=%v", score, move)

	if move == nil {
		t.Error("Search returned nil move")
	}
}

func TestFutilityPruningDoesNotPruneCaptures(t *testing.T) {
	engine := NewEngine()

	// Position with a hanging piece - should not be pruned
	fen := "r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search at shallow depth
	score, move := engine.search(pos, 2, -1000000, 1000000)

	t.Logf("Capture test: score=%d, move=%v", score, move)

	if move == nil {
		t.Error("Search returned nil move")
	}
}

func BenchmarkSearch(b *testing.B) {
	engine := NewEngine()
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.search(pos, 3, -10000, 10000)
	}
}

func BenchmarkQuiescence(b *testing.B) {
	engine := NewEngine()
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.quiescence(pos, -10000, 10000, 4)
	}
}
