package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestCounterMoveBasic(t *testing.T) {
	cm := NewCounterMoveTable()

	// Create test moves
	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	if len(moves) < 2 {
		t.Fatal("Not enough moves")
	}

	prevMove := moves[0]    // e.g., e7e5
	counterMove := moves[1] // e.g., d7d5

	// Initially no counter move
	got := cm.get(prevMove, chess.Pawn)
	if got != nil {
		t.Error("Expected nil counter move initially")
	}

	// Update counter move
	cm.update(prevMove, counterMove, chess.Pawn)

	// Now should have counter move
	got = cm.get(prevMove, chess.Pawn)
	if got == nil {
		t.Error("Expected counter move after update")
	}

	if got.String() != counterMove.String() {
		t.Errorf("Expected counter move %s, got %s", counterMove.String(), got.String())
	}
}

func TestCounterMoveIsCounterMove(t *testing.T) {
	cm := NewCounterMoveTable()

	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	if len(moves) < 3 {
		t.Fatal("Not enough moves")
	}

	prevMove := moves[0]
	counterMove := moves[1]
	otherMove := moves[2]

	cm.update(prevMove, counterMove, chess.Pawn)

	// Should be counter move
	if !cm.isCounterMove(counterMove, prevMove, chess.Pawn) {
		t.Error("Expected isCounterMove to return true")
	}

	// Other move should not be counter move
	if cm.isCounterMove(otherMove, prevMove, chess.Pawn) {
		t.Error("Expected isCounterMove to return false for other move")
	}
}

func TestCounterMoveClear(t *testing.T) {
	cm := NewCounterMoveTable()

	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	prevMove := moves[0]
	counterMove := moves[1]

	cm.update(prevMove, counterMove, chess.Pawn)

	// Verify it's set
	if cm.get(prevMove, chess.Pawn) == nil {
		t.Error("Counter move should be set")
	}

	// Clear
	cm.clear()

	// Should be nil now
	if cm.get(prevMove, chess.Pawn) != nil {
		t.Error("Counter move should be nil after clear")
	}
}

func TestCounterMoveDoesNotStoreCaptures(t *testing.T) {
	cm := NewCounterMoveTable()

	// Position with captures available
	fen := "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	var captureMove *chess.Move
	var quietMove *chess.Move
	for _, m := range moves {
		if m.HasTag(chess.Capture) && captureMove == nil {
			captureMove = m
		} else if !m.HasTag(chess.Capture) && quietMove == nil {
			quietMove = m
		}
	}

	if captureMove == nil || quietMove == nil {
		t.Skip("Need both capture and quiet moves")
	}

	// Try to store capture as counter move
	cm.update(quietMove, captureMove, chess.Pawn)

	// Should not be stored (captures are not stored as counter moves)
	if cm.get(quietMove, chess.Pawn) != nil {
		t.Error("Captures should not be stored as counter moves")
	}

	// Store quiet move
	cm.update(captureMove, quietMove, chess.Pawn)

	// Should be stored
	if cm.get(captureMove, chess.Pawn) == nil {
		t.Error("Quiet moves should be stored as counter moves")
	}
}
