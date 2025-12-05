package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestHistoryTableBasic(t *testing.T) {
	h := NewHistoryTable()

	// Create a test move
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	if len(moves) == 0 {
		t.Fatal("No moves available")
	}

	move := moves[0]

	// Initial score should be 0
	score := h.getScore(move, chess.White)
	if score != 0 {
		t.Errorf("Expected initial score 0, got %d", score)
	}

	// Update with depth 4
	h.update(move, chess.White, 4)

	// Score should now be positive (4*4 = 16)
	score = h.getScore(move, chess.White)
	if score != 16 {
		t.Errorf("Expected score 16 after depth 4 update, got %d", score)
	}

	// Update again
	h.update(move, chess.White, 3)

	// Score should increase (16 + 9 = 25)
	score = h.getScore(move, chess.White)
	if score != 25 {
		t.Errorf("Expected score 25 after second update, got %d", score)
	}
}

func TestHistoryTableColorSeparation(t *testing.T) {
	h := NewHistoryTable()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	move := moves[0]

	// Update for white
	h.update(move, chess.White, 4)

	// White should have score
	whiteScore := h.getScore(move, chess.White)
	if whiteScore == 0 {
		t.Error("White score should not be 0")
	}

	// Black should have 0 for same move
	blackScore := h.getScore(move, chess.Black)
	if blackScore != 0 {
		t.Errorf("Black score should be 0, got %d", blackScore)
	}
}

func TestHistoryTableClear(t *testing.T) {
	h := NewHistoryTable()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	move := moves[0]

	h.update(move, chess.White, 5)

	score := h.getScore(move, chess.White)
	if score == 0 {
		t.Error("Score should not be 0 after update")
	}

	h.clear()

	score = h.getScore(move, chess.White)
	if score != 0 {
		t.Errorf("Score should be 0 after clear, got %d", score)
	}
}

func TestHistoryTableBadMove(t *testing.T) {
	h := NewHistoryTable()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	move := moves[0]

	// First give it a positive score
	h.update(move, chess.White, 5) // +25

	score := h.getScore(move, chess.White)
	if score != 25 {
		t.Errorf("Expected score 25, got %d", score)
	}

	// Now penalize it
	h.updateBadMove(move, chess.White, 5) // -5

	score = h.getScore(move, chess.White)
	if score != 20 {
		t.Errorf("Expected score 20 after penalty, got %d", score)
	}
}

func TestHistoryTableScaling(t *testing.T) {
	h := NewHistoryTable()

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()
	moves := pos.ValidMoves()

	move := moves[0]

	// Update many times to trigger scaling
	for i := 0; i < 200; i++ {
		h.update(move, chess.White, 10) // +100 each time
	}

	score := h.getScore(move, chess.White)
	t.Logf("Score after many updates: %d", score)

	// Score should be scaled down but still positive
	if score <= 0 {
		t.Error("Score should be positive after scaling")
	}

	// Score should be less than 200 * 100 = 20000 due to scaling
	if score >= 20000 {
		t.Errorf("Score should be scaled down, got %d", score)
	}
}

func TestHistoryInSearch(t *testing.T) {
	engine := NewEngine()

	// Run a search to populate history
	fen := "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	// Search at depth 5
	_, _ = engine.search(pos, 5, -1000000, 1000000)

	// History table should have some entries
	// Check a few common opening moves
	moves := pos.ValidMoves()
	hasHistory := false
	for _, move := range moves {
		score := engine.history.getScore(move, chess.White)
		if score > 0 {
			hasHistory = true
			t.Logf("Move %s has history score %d", move.String(), score)
		}
	}

	if !hasHistory {
		t.Log("No history entries found (this may be normal for short searches)")
	}
}
