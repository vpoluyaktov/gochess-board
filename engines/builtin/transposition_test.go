package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestNewTranspositionTable(t *testing.T) {
	tt := NewTranspositionTable(1) // 1MB

	if tt == nil {
		t.Fatal("NewTranspositionTable returned nil")
	}

	if tt.size == 0 {
		t.Error("TranspositionTable size is 0")
	}

	if len(tt.entries) != tt.size {
		t.Errorf("Expected %d entries, got %d", tt.size, len(tt.entries))
	}

	t.Logf("Created transposition table with %d entries", tt.size)
}

func TestTranspositionTableStoreAndProbe(t *testing.T) {
	tt := NewTranspositionTable(1)

	// Create a test position
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)
	depth := 5
	score := 100
	move := pos.ValidMoves()[0]

	// Store entry
	tt.store(zobristKey, depth, score, TTExact, move)

	// Probe entry
	found, retrievedScore, retrievedMove := tt.probe(zobristKey, depth, -1000, 1000)

	if !found {
		t.Error("Expected to find stored entry")
	}

	if retrievedScore != score {
		t.Errorf("Expected score %d, got %d", score, retrievedScore)
	}

	if retrievedMove == nil {
		t.Error("Expected to retrieve move")
	} else if retrievedMove.String() != move.String() {
		t.Errorf("Expected move %s, got %s", move.String(), retrievedMove.String())
	}
}

func TestTranspositionTableDepthReplacement(t *testing.T) {
	tt := NewTranspositionTable(1)

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)
	moves := pos.ValidMoves()

	// Store shallow search
	tt.store(zobristKey, 3, 50, TTExact, moves[0])

	// Store deeper search - should replace
	tt.store(zobristKey, 5, 100, TTExact, moves[1])

	// Probe should return deeper search result
	found, score, move := tt.probe(zobristKey, 5, -1000, 1000)

	if !found {
		t.Error("Expected to find entry")
	}

	if score != 100 {
		t.Errorf("Expected score from deeper search (100), got %d", score)
	}

	if move.String() != moves[1].String() {
		t.Errorf("Expected move from deeper search (%s), got %s", moves[1].String(), move.String())
	}
}

func TestTranspositionTableExactScore(t *testing.T) {
	tt := NewTranspositionTable(1)

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)

	// Store exact score
	tt.store(zobristKey, 5, 100, TTExact, nil)

	// Probe with same depth should return exact score
	found, score, _ := tt.probe(zobristKey, 5, -1000, 1000)

	if !found {
		t.Error("Expected to find entry")
	}

	if score != 100 {
		t.Errorf("Expected exact score 100, got %d", score)
	}
}

func TestTranspositionTableAlphaBound(t *testing.T) {
	tt := NewTranspositionTable(1)

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)

	// Store alpha bound (upper bound)
	tt.store(zobristKey, 5, 50, TTAlpha, nil)

	// Probe with alpha=100 - stored score (50) <= alpha, should use it
	found, score, _ := tt.probe(zobristKey, 5, 100, 1000)

	if !found {
		t.Error("Expected to find entry")
	}

	if score != 100 {
		t.Errorf("Expected alpha cutoff score 100, got %d", score)
	}
}

func TestTranspositionTableBetaBound(t *testing.T) {
	tt := NewTranspositionTable(1)

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)

	// Store beta bound (lower bound)
	tt.store(zobristKey, 5, 200, TTBeta, nil)

	// Probe with beta=100 - stored score (200) >= beta, should use it
	found, score, _ := tt.probe(zobristKey, 5, -1000, 100)

	if !found {
		t.Error("Expected to find entry")
	}

	if score != 100 {
		t.Errorf("Expected beta cutoff score 100, got %d", score)
	}
}

func TestTranspositionTableClear(t *testing.T) {
	tt := NewTranspositionTable(1)

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)

	// Store entry
	tt.store(zobristKey, 5, 100, TTExact, nil)

	// Verify it's there
	found, _, _ := tt.probe(zobristKey, 5, -1000, 1000)
	if !found {
		t.Error("Expected to find entry before clear")
	}

	// Clear table
	tt.clear()

	// Should not find entry after clear
	found, _, _ = tt.probe(zobristKey, 5, -1000, 1000)
	if found {
		t.Error("Should not find entry after clear")
	}
}

func TestZobristKeyDifferentPositions(t *testing.T) {
	// Starting position
	fen1 := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc1, _ := chess.FEN(fen1)
	game1 := chess.NewGame(fenFunc1)
	key1 := getZobristKey(game1.Position())

	// After e4
	fen2 := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	fenFunc2, _ := chess.FEN(fen2)
	game2 := chess.NewGame(fenFunc2)
	key2 := getZobristKey(game2.Position())

	if key1 == key2 {
		t.Error("Different positions should have different zobrist keys")
	}

	t.Logf("Key1: %x, Key2: %x", key1, key2)
}

func TestZobristKeySamePosition(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	// Create same position twice
	fenFunc1, _ := chess.FEN(fen)
	game1 := chess.NewGame(fenFunc1)
	key1 := getZobristKey(game1.Position())

	fenFunc2, _ := chess.FEN(fen)
	game2 := chess.NewGame(fenFunc2)
	key2 := getZobristKey(game2.Position())

	if key1 != key2 {
		t.Error("Same positions should have same zobrist key")
	}
}

func TestTranspositionTableIncrementAge(t *testing.T) {
	tt := NewTranspositionTable(1)

	initialAge := tt.age
	tt.incrementAge()

	if tt.age != initialAge+1 {
		t.Errorf("Expected age %d, got %d", initialAge+1, tt.age)
	}
}

func BenchmarkTranspositionTableStore(b *testing.B) {
	tt := NewTranspositionTable(64)

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)
	move := pos.ValidMoves()[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tt.store(zobristKey, 5, 100, TTExact, move)
	}
}

func BenchmarkTranspositionTableProbe(b *testing.B) {
	tt := NewTranspositionTable(64)

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	zobristKey := getZobristKey(pos)
	move := pos.ValidMoves()[0]

	tt.store(zobristKey, 5, 100, TTExact, move)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tt.probe(zobristKey, 5, -1000, 1000)
	}
}

func BenchmarkZobristKey(b *testing.B) {
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)
	pos := game.Position()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getZobristKey(pos)
	}
}
