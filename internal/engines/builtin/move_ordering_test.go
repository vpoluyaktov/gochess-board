package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestMoveOrderScore(t *testing.T) {
	engine := NewEngine()

	// Create a position to test move scoring
	fen := "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	moves := pos.ValidMoves()

	// Find specific move types
	var captureMove, checkMove, quietMove *chess.Move

	for _, move := range moves {
		if move.HasTag(chess.Capture) && captureMove == nil {
			captureMove = move
		}
		if move.HasTag(chess.Check) && checkMove == nil {
			checkMove = move
		}
		if !move.HasTag(chess.Capture) && !move.HasTag(chess.Check) && quietMove == nil {
			quietMove = move
		}
	}

	// Test that captures are scored higher than quiet moves
	if captureMove != nil && quietMove != nil {
		captureScore := engine.moveOrderScore(captureMove, nil, 0)
		quietScore := engine.moveOrderScore(quietMove, nil, 0)

		if captureScore <= quietScore {
			t.Errorf("Capture move should score higher than quiet move: capture=%d, quiet=%d",
				captureScore, quietScore)
		}

		t.Logf("Capture score: %d, Quiet score: %d", captureScore, quietScore)
	}

	// Test that checks are scored higher than quiet moves
	if checkMove != nil && quietMove != nil {
		checkScore := engine.moveOrderScore(checkMove, nil, 0)
		quietScore := engine.moveOrderScore(quietMove, nil, 0)

		if checkScore <= quietScore {
			t.Errorf("Check move should score higher than quiet move: check=%d, quiet=%d",
				checkScore, quietScore)
		}

		t.Logf("Check score: %d, Quiet score: %d", checkScore, quietScore)
	}
}

func TestOrderMoves(t *testing.T) {
	engine := NewEngine()

	// Position with various move types
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	moves := pos.ValidMoves()
	originalMoves := make([]*chess.Move, len(moves))
	copy(originalMoves, moves)

	// Order the moves
	engine.orderMoves(moves, nil, 0)

	// Verify moves are reordered (at least some change should occur if there are captures)
	hasCapture := false
	for _, move := range originalMoves {
		if move.HasTag(chess.Capture) {
			hasCapture = true
			break
		}
	}

	if hasCapture {
		// After ordering, captures should tend to be earlier
		capturesInFirstHalf := 0
		halfPoint := len(moves) / 2

		for i := 0; i < halfPoint && i < len(moves); i++ {
			if moves[i].HasTag(chess.Capture) {
				capturesInFirstHalf++
			}
		}

		t.Logf("Captures in first half after ordering: %d out of %d moves",
			capturesInFirstHalf, halfPoint)
	}

	// Verify all moves are still present (just reordered)
	if len(moves) != len(originalMoves) {
		t.Errorf("Move count changed after ordering: before=%d, after=%d",
			len(originalMoves), len(moves))
	}
}

func TestOrderMovesWithCaptures(t *testing.T) {
	engine := NewEngine()

	// Position where white can capture
	fen := "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	moves := pos.ValidMoves()
	engine.orderMoves(moves, nil, 0)

	// Count captures and their positions
	firstCaptureIndex := -1
	lastQuietIndex := -1

	for i, move := range moves {
		if move.HasTag(chess.Capture) && firstCaptureIndex == -1 {
			firstCaptureIndex = i
		}
		if !move.HasTag(chess.Capture) && !move.HasTag(chess.Check) {
			lastQuietIndex = i
		}
	}

	// If we have both captures and quiet moves, captures should come first
	if firstCaptureIndex != -1 && lastQuietIndex != -1 {
		if firstCaptureIndex > lastQuietIndex {
			t.Errorf("Captures should be ordered before quiet moves: first capture at %d, last quiet at %d",
				firstCaptureIndex, lastQuietIndex)
		}

		t.Logf("First capture at index %d, last quiet move at index %d",
			firstCaptureIndex, lastQuietIndex)
	}
}

func TestOrderMovesEmptyList(t *testing.T) {
	engine := NewEngine()

	// Test with empty move list (shouldn't crash)
	var moves []*chess.Move
	engine.orderMoves(moves, nil, 0)

	if len(moves) != 0 {
		t.Errorf("Empty move list should remain empty")
	}
}

func TestOrderMovesSingleMove(t *testing.T) {
	engine := NewEngine()

	// Create a slice with just one move from a position
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	allMoves := pos.ValidMoves()
	if len(allMoves) == 0 {
		t.Skip("Position has no moves")
	}

	// Take just one move
	moves := []*chess.Move{allMoves[0]}
	originalMove := moves[0]

	engine.orderMoves(moves, nil, 0)

	// Single move should remain unchanged
	if len(moves) != 1 {
		t.Errorf("Single move list should remain single move, got %d", len(moves))
	}

	if moves[0] != originalMove {
		t.Errorf("Single move should not change")
	}
}

func TestMoveOrderingConsistency(t *testing.T) {
	engine := NewEngine()

	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Order moves multiple times - should be consistent
	moves1 := pos.ValidMoves()
	moves2 := pos.ValidMoves()

	engine.orderMoves(moves1, nil, 0)
	engine.orderMoves(moves2, nil, 0)

	// Results should be identical
	if len(moves1) != len(moves2) {
		t.Errorf("Inconsistent move ordering: different lengths")
	}

	for i := range moves1 {
		if moves1[i].String() != moves2[i].String() {
			t.Errorf("Inconsistent move ordering at index %d: %s vs %s",
				i, moves1[i].String(), moves2[i].String())
		}
	}
}

func BenchmarkOrderMoves(b *testing.B) {
	engine := NewEngine()
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		moves := pos.ValidMoves()
		engine.orderMoves(moves, nil, 0)
	}
}

func BenchmarkMoveOrderScore(b *testing.B) {
	engine := NewEngine()
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, move := range moves {
			engine.moveOrderScore(move, nil, 0)
		}
	}
}
