package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestNewKillerMoves(t *testing.T) {
	km := NewKillerMoves()

	if km == nil {
		t.Fatal("NewKillerMoves returned nil")
	}

	// All slots should be empty initially
	for ply := 0; ply < maxDepth; ply++ {
		if km.moves[ply][0] != nil || km.moves[ply][1] != nil {
			t.Errorf("Ply %d should have nil killer moves initially", ply)
		}
	}
}

func TestKillerMovesAdd(t *testing.T) {
	km := NewKillerMoves()

	// Create test moves
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	// Find a quiet move (non-capture)
	var quietMove *chess.Move
	for _, move := range moves {
		if !move.HasTag(chess.Capture) && !move.HasTag(chess.EnPassant) {
			quietMove = move
			break
		}
	}

	if quietMove == nil {
		t.Skip("No quiet moves available")
	}

	ply := 5
	km.add(quietMove, ply)

	// Should be stored as primary killer
	if km.moves[ply][0] == nil {
		t.Error("Primary killer should not be nil")
	} else if km.moves[ply][0].String() != quietMove.String() {
		t.Errorf("Primary killer should be %s, got %s", quietMove.String(), km.moves[ply][0].String())
	}
}

func TestKillerMovesAddTwoMoves(t *testing.T) {
	km := NewKillerMoves()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	// Get two quiet moves
	var quietMoves []*chess.Move
	for _, move := range moves {
		if !move.HasTag(chess.Capture) && !move.HasTag(chess.EnPassant) {
			quietMoves = append(quietMoves, move)
			if len(quietMoves) == 2 {
				break
			}
		}
	}

	if len(quietMoves) < 2 {
		t.Skip("Need at least 2 quiet moves")
	}

	ply := 3
	km.add(quietMoves[0], ply)
	km.add(quietMoves[1], ply)

	// First move should be shifted to secondary
	if km.moves[ply][1] == nil {
		t.Error("Secondary killer should not be nil")
	} else if km.moves[ply][1].String() != quietMoves[0].String() {
		t.Errorf("Secondary killer should be %s, got %s", quietMoves[0].String(), km.moves[ply][1].String())
	}

	// Second move should be primary
	if km.moves[ply][0] == nil {
		t.Error("Primary killer should not be nil")
	} else if km.moves[ply][0].String() != quietMoves[1].String() {
		t.Errorf("Primary killer should be %s, got %s", quietMoves[1].String(), km.moves[ply][0].String())
	}
}

func TestKillerMovesDoNotStoreCaptures(t *testing.T) {
	km := NewKillerMoves()

	// Position with captures
	fen := "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	// Find a capture move
	var captureMove *chess.Move
	for _, move := range moves {
		if move.HasTag(chess.Capture) || move.HasTag(chess.EnPassant) {
			captureMove = move
			break
		}
	}

	if captureMove == nil {
		t.Skip("No capture moves available")
	}

	ply := 2
	km.add(captureMove, ply)

	// Should not be stored
	if km.moves[ply][0] != nil {
		t.Error("Captures should not be stored as killer moves")
	}
}

func TestKillerMovesIsKiller(t *testing.T) {
	km := NewKillerMoves()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	var quietMove *chess.Move
	for _, move := range moves {
		if !move.HasTag(chess.Capture) {
			quietMove = move
			break
		}
	}

	if quietMove == nil {
		t.Skip("No quiet moves available")
	}

	ply := 4
	km.add(quietMove, ply)

	// Should be recognized as killer
	if !km.isKiller(quietMove, ply) {
		t.Error("Move should be recognized as killer")
	}

	// Should not be killer at different ply
	if km.isKiller(quietMove, ply+1) {
		t.Error("Move should not be killer at different ply")
	}
}

func TestKillerMovesGetKillerScore(t *testing.T) {
	km := NewKillerMoves()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	var quietMoves []*chess.Move
	for _, move := range moves {
		if !move.HasTag(chess.Capture) {
			quietMoves = append(quietMoves, move)
			if len(quietMoves) == 2 {
				break
			}
		}
	}

	if len(quietMoves) < 2 {
		t.Skip("Need at least 2 quiet moves")
	}

	ply := 6
	km.add(quietMoves[0], ply)
	km.add(quietMoves[1], ply)

	// Primary killer should have higher score
	primaryScore := km.getKillerScore(quietMoves[1], ply)
	secondaryScore := km.getKillerScore(quietMoves[0], ply)

	if primaryScore <= secondaryScore {
		t.Errorf("Primary killer score (%d) should be higher than secondary (%d)",
			primaryScore, secondaryScore)
	}

	if primaryScore != 9000 {
		t.Errorf("Primary killer score should be 9000, got %d", primaryScore)
	}

	if secondaryScore != 8000 {
		t.Errorf("Secondary killer score should be 8000, got %d", secondaryScore)
	}
}

func TestKillerMovesClear(t *testing.T) {
	km := NewKillerMoves()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	var quietMove *chess.Move
	for _, move := range moves {
		if !move.HasTag(chess.Capture) {
			quietMove = move
			break
		}
	}

	if quietMove == nil {
		t.Skip("No quiet moves available")
	}

	// Add some killer moves
	for ply := 0; ply < 10; ply++ {
		km.add(quietMove, ply)
	}

	// Clear
	km.clear()

	// All should be nil
	for ply := 0; ply < maxDepth; ply++ {
		if km.moves[ply][0] != nil || km.moves[ply][1] != nil {
			t.Errorf("Ply %d should have nil killer moves after clear", ply)
		}
	}
}

func TestKillerMovesBoundaryConditions(t *testing.T) {
	km := NewKillerMoves()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	var quietMove *chess.Move
	for _, move := range moves {
		if !move.HasTag(chess.Capture) {
			quietMove = move
			break
		}
	}

	if quietMove == nil {
		t.Skip("No quiet moves available")
	}

	// Test negative ply
	km.add(quietMove, -1)
	if km.isKiller(quietMove, -1) {
		t.Error("Should not store killer at negative ply")
	}

	// Test ply >= maxDepth
	km.add(quietMove, maxDepth)
	if km.isKiller(quietMove, maxDepth) {
		t.Error("Should not store killer at ply >= maxDepth")
	}

	// Test valid boundary
	km.add(quietMove, maxDepth-1)
	if !km.isKiller(quietMove, maxDepth-1) {
		t.Error("Should store killer at ply = maxDepth-1")
	}
}

func TestKillerMovesSameMoveAddedTwice(t *testing.T) {
	km := NewKillerMoves()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	var quietMove *chess.Move
	for _, move := range moves {
		if !move.HasTag(chess.Capture) {
			quietMove = move
			break
		}
	}

	if quietMove == nil {
		t.Skip("No quiet moves available")
	}

	ply := 5
	km.add(quietMove, ply)
	km.add(quietMove, ply) // Add same move again

	// Should still be primary, secondary should be nil
	if km.moves[ply][0] == nil || km.moves[ply][0].String() != quietMove.String() {
		t.Error("Primary killer should still be the same move")
	}

	if km.moves[ply][1] != nil {
		t.Error("Secondary killer should be nil when same move is added twice")
	}
}
