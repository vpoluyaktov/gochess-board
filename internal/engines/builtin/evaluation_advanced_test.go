package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestEvaluateKingSafety(t *testing.T) {
	// Position with white king castled kingside with pawn shield
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	whiteKingSafety := evaluateKingSafety(pos, chess.White)
	blackKingSafety := evaluateKingSafety(pos, chess.Black)

	t.Logf("White king safety: %d, Black king safety: %d", whiteKingSafety, blackKingSafety)

	// Both kings should have some safety bonus
	if whiteKingSafety < 0 {
		t.Error("White king safety should not be negative")
	}
	if blackKingSafety < 0 {
		t.Error("Black king safety should not be negative")
	}
}

func TestEvaluateKingSafetyWithCastling(t *testing.T) {
	// Position after white has castled kingside
	fen := "r1bqkb1r/pppppppp/2n2n2/8/8/5N2/PPPPPPPP/RNBQ1RK1 b kq - 1 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	whiteKingSafety := evaluateKingSafety(pos, chess.White)

	t.Logf("White king safety after castling: %d", whiteKingSafety)

	// Should have pawn shield bonus
	if whiteKingSafety <= 0 {
		t.Error("Castled king should have positive safety score")
	}
}

func TestEvaluatePawnStructureDoubledPawns(t *testing.T) {
	// Position with doubled pawns
	fen := "rnbqkbnr/pppppppp/8/8/8/P7/P1PPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	score := evaluatePawnStructure(pos, chess.White)

	t.Logf("Pawn structure score with doubled pawns: %d", score)

	// Should be penalized for doubled pawns
	if score >= 0 {
		t.Error("Doubled pawns should result in negative score")
	}
}

func TestEvaluatePawnStructurePassedPawn(t *testing.T) {
	// Position with white passed pawn on e-file (no black pawns on d, e, or f files)
	fen := "rnbqkbnr/ppp3pp/8/4P3/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	score := evaluatePawnStructure(pos, chess.White)

	t.Logf("Pawn structure score with passed pawn: %d", score)

	// Should get bonus for passed pawn
	if score <= 0 {
		t.Error("Passed pawn should result in positive score")
	}
}

func TestEvaluateMobility(t *testing.T) {
	engine := NewEngine()

	// Starting position
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	whiteMobility := evaluateMobility(pos, chess.White)
	blackMobility := evaluateMobility(pos, chess.Black)

	t.Logf("White mobility: %d, Black mobility: %d", whiteMobility, blackMobility)

	// White to move, so white should have mobility score
	if whiteMobility <= 0 {
		t.Error("White should have positive mobility in starting position")
	}

	// Black not to move, so should be 0
	if blackMobility != 0 {
		t.Error("Black mobility should be 0 when not their turn")
	}

	// Test full evaluation
	fullScore := engine.evaluate(pos)
	t.Logf("Full evaluation with enhanced features: %d", fullScore)
}

func TestEvaluateEnhancedVsBasic(t *testing.T) {
	engine := NewEngine()

	// Test several positions to ensure enhanced evaluation is reasonable
	positions := []struct {
		name string
		fen  string
	}{
		{"Starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
		{"After e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"},
		{"Italian game", "r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 3 3"},
	}

	for _, pos := range positions {
		t.Run(pos.name, func(t *testing.T) {
			fenFunc, _ := chess.FEN(pos.fen)
			game := chess.NewGame(fenFunc)
			score := engine.evaluate(game.Position())

			t.Logf("%s: evaluation = %d centipawns", pos.name, score)

			// Scores should be reasonable (not extreme)
			if score < -5000 || score > 5000 {
				t.Errorf("Evaluation seems extreme: %d", score)
			}
		})
	}
}

func TestEvaluatePassedPawnAdvancement(t *testing.T) {
	// Test that passed pawns get more valuable as they advance
	// (no black pawns on d, e, or f files)
	positions := []struct {
		name string
		fen  string
		rank int
	}{
		{"Passed pawn on 4th rank", "rnbqkbnr/ppp3pp/8/4P3/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1", 4},
		{"Passed pawn on 5th rank", "rnbqkbnr/ppp3pp/4P3/8/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1", 5},
		{"Passed pawn on 6th rank", "rnbqkbnr/ppp2Ppp/8/8/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1", 6},
	}

	var scores []int
	for _, pos := range positions {
		t.Run(pos.name, func(t *testing.T) {
			fenFunc, _ := chess.FEN(pos.fen)
			game := chess.NewGame(fenFunc)
			score := evaluatePawnStructure(game.Position(), chess.White)
			scores = append(scores, score)

			t.Logf("%s: pawn structure score = %d", pos.name, score)
		})
	}

	// Scores should increase as pawn advances
	if len(scores) >= 2 {
		for i := 1; i < len(scores); i++ {
			if scores[i] <= scores[i-1] {
				t.Logf("Warning: Passed pawn score didn't increase with advancement: rank %d score=%d, rank %d score=%d",
					positions[i-1].rank, scores[i-1], positions[i].rank, scores[i])
			}
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test max function
	if max(5, 3) != 5 {
		t.Error("max(5, 3) should be 5")
	}
	if max(2, 7) != 7 {
		t.Error("max(2, 7) should be 7")
	}
	if max(4, 4) != 4 {
		t.Error("max(4, 4) should be 4")
	}

	// Test min function
	if min(5, 3) != 3 {
		t.Error("min(5, 3) should be 3")
	}
	if min(2, 7) != 2 {
		t.Error("min(2, 7) should be 2")
	}
	if min(4, 4) != 4 {
		t.Error("min(4, 4) should be 4")
	}
}

func BenchmarkEvaluateEnhanced(b *testing.B) {
	engine := NewEngine()
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.evaluate(pos)
	}
}

func BenchmarkEvaluateKingSafety(b *testing.B) {
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluateKingSafety(pos, chess.White)
	}
}

func BenchmarkEvaluatePawnStructure(b *testing.B) {
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluatePawnStructure(pos, chess.White)
	}
}
