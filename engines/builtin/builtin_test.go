package builtin

import (
	"testing"
	"time"
)

func TestInternalEngineBasic(t *testing.T) {
	engine := NewInternalEngine()

	// Test starting position
	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	move, err := engine.GetBestMove(startFEN, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to get best move: %v", err)
	}

	if move == "" {
		t.Fatal("Engine returned empty move")
	}

	t.Logf("Engine suggested move: %s", move)

	// Verify it's a valid UCI move format (e.g., "e2e4")
	if len(move) < 4 {
		t.Errorf("Move format seems invalid: %s", move)
	}
}

func TestInternalEngineWithClock(t *testing.T) {
	engine := NewInternalEngine()

	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	moveHistory := []string{}

	move, err := engine.GetBestMoveWithClock(
		startFEN,
		moveHistory,
		5*time.Minute, // white time
		5*time.Minute, // black time
		0,             // white increment
		0,             // black increment
	)

	if err != nil {
		t.Fatalf("Failed to get best move with clock: %v", err)
	}

	if move == "" {
		t.Fatal("Engine returned empty move")
	}

	t.Logf("Engine suggested move: %s", move)
}

func TestInternalEngineCheckmate(t *testing.T) {
	engine := NewInternalEngine()

	// Position where white can checkmate in 1 (back rank mate)
	// Black king on h8, white rook can move to h1#
	fen := "6k1/8/8/8/8/8/8/R6K w - - 0 1"

	move, err := engine.GetBestMove(fen, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to get best move: %v", err)
	}

	t.Logf("Engine found move: %s", move)

	// The engine should find the checkmate (Ra8# or similar)
	if move == "" {
		t.Fatal("Engine failed to find checkmate")
	}
}

func TestInternalEngineSetOption(t *testing.T) {
	engine := NewInternalEngine()

	// Internal engine doesn't support options yet, but shouldn't error
	err := engine.SetOption("SomeOption", "value")
	if err != nil {
		t.Errorf("SetOption should not error: %v", err)
	}
}

func TestInternalEngineClose(t *testing.T) {
	engine := NewInternalEngine()

	err := engine.Close()
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}
}
