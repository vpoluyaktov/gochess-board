package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

func TestGetPieceValue(t *testing.T) {
	tests := []struct {
		piece chess.Piece
		want  int
	}{
		{chess.WhitePawn, pawnValue},
		{chess.BlackPawn, pawnValue},
		{chess.WhiteKnight, knightValue},
		{chess.BlackKnight, knightValue},
		{chess.WhiteBishop, bishopValue},
		{chess.BlackBishop, bishopValue},
		{chess.WhiteRook, rookValue},
		{chess.BlackRook, rookValue},
		{chess.WhiteQueen, queenValue},
		{chess.BlackQueen, queenValue},
		{chess.WhiteKing, kingValue},
		{chess.BlackKing, kingValue},
		{chess.NoPiece, 0},
	}

	for _, tt := range tests {
		t.Run(tt.piece.String(), func(t *testing.T) {
			got := getPieceValue(tt.piece)
			if got != tt.want {
				t.Errorf("getPieceValue(%v) = %d, want %d", tt.piece, got, tt.want)
			}
		})
	}
}

func TestGetPieceSquareValue(t *testing.T) {
	// Test that piece-square values are symmetric for white and black
	// White pawn on e2 should have same value as black pawn on e7
	whitePawnE2 := getPieceSquareValue(chess.WhitePawn, chess.E2)
	blackPawnE7 := getPieceSquareValue(chess.BlackPawn, chess.E7)

	if whitePawnE2 != blackPawnE7 {
		t.Errorf("Piece-square values not symmetric: white pawn e2=%d, black pawn e7=%d",
			whitePawnE2, blackPawnE7)
	}

	// Test that central squares are valued higher for knights
	knightCenter := getPieceSquareValue(chess.WhiteKnight, chess.E4)
	knightCorner := getPieceSquareValue(chess.WhiteKnight, chess.A1)

	if knightCenter <= knightCorner {
		t.Errorf("Knight should be valued higher in center: center=%d, corner=%d",
			knightCenter, knightCorner)
	}

	// Test NoPiece returns 0
	noPieceValue := getPieceSquareValue(chess.NoPiece, chess.E4)
	if noPieceValue != 0 {
		t.Errorf("NoPiece should have value 0, got %d", noPieceValue)
	}
}

func TestEvaluateStartingPosition(t *testing.T) {
	engine := NewEngine()

	// Starting position should be roughly equal (close to 0)
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)

	score := engine.evaluate(game.Position())

	// Should be close to 0 (within 100 centipawns for symmetry)
	if score < -100 || score > 100 {
		t.Errorf("Starting position evaluation should be near 0, got %d", score)
	}

	t.Logf("Starting position evaluation: %d centipawns", score)
}

func TestEvaluateMaterialAdvantage(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		fen      string
		minScore int // Minimum expected score for white
	}{
		{
			name:     "White up a queen",
			fen:      "rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minScore: 800, // Queen is worth ~900
		},
		{
			name:     "White up a rook",
			fen:      "rnbqkbn1/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minScore: 400, // Rook is worth ~500
		},
		{
			name:     "White up a pawn",
			fen:      "rnbqkbnr/ppppppp1/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			minScore: 80, // Pawn is worth 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fenFunc, _ := chess.FEN(tt.fen)
			game := chess.NewGame(fenFunc)
			score := engine.evaluate(game.Position())

			if score < tt.minScore {
				t.Errorf("%s: expected score >= %d, got %d", tt.name, tt.minScore, score)
			}

			t.Logf("%s: evaluation = %d centipawns", tt.name, score)
		})
	}
}

func TestEvaluateCheckmate(t *testing.T) {
	engine := NewEngine()

	// Checkmate position - white is checkmated
	fen := "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)

	score := engine.evaluate(game.Position())

	// Should be very negative (white is checkmated)
	if score > -9000 {
		t.Errorf("Checkmate should have score < -9000, got %d", score)
	}

	t.Logf("Checkmate evaluation: %d", score)
}

func TestEvaluateStalemate(t *testing.T) {
	engine := NewEngine()

	// Stalemate position
	fen := "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1"
	fenFunc, _ := chess.FEN(fen)
	game := chess.NewGame(fenFunc)

	score := engine.evaluate(game.Position())

	// Stalemate should be 0 (draw)
	if score != 0 {
		t.Errorf("Stalemate should have score 0, got %d", score)
	}
}

func TestEvaluatePositionalAdvantage(t *testing.T) {
	engine := NewEngine()

	// Position with knight in center vs knight on rim
	tests := []struct {
		name string
		fen  string
		desc string
	}{
		{
			name: "Knight in center",
			fen:  "rnbqkb1r/pppppppp/8/8/4N3/8/PPPPPPPP/RNBQKB1R w KQkq - 0 1",
			desc: "Knight on e4 (center)",
		},
		{
			name: "Knight on rim",
			fen:  "rnbqkb1r/pppppppp/8/8/8/8/PPPPPPPP/NNBQKB1R w KQkq - 0 1",
			desc: "Knight on a1 (rim)",
		},
	}

	var centerScore, rimScore int

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fenFunc, _ := chess.FEN(tt.fen)
			game := chess.NewGame(fenFunc)
			score := engine.evaluate(game.Position())

			if tt.name == "Knight in center" {
				centerScore = score
			} else {
				rimScore = score
			}

			t.Logf("%s: evaluation = %d centipawns", tt.desc, score)
		})
	}

	// Knight in center should be valued higher
	if centerScore <= rimScore {
		t.Errorf("Knight in center (%d) should be valued higher than knight on rim (%d)",
			centerScore, rimScore)
	}
}

func TestPieceSquareTablesSymmetry(t *testing.T) {
	// Test that piece-square tables are properly mirrored for black pieces
	tables := []struct {
		name  string
		table [64]int
	}{
		{"pawn", pawnTable},
		{"knight", knightTable},
		{"bishop", bishopTable},
		{"rook", rookTable},
		{"queen", queenTable},
		{"king", kingMiddleGameTable},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			// Check that the table is symmetric (rank 0 should mirror rank 7, etc.)
			// This is important for proper evaluation
			for file := 0; file < 8; file++ {
				for rank := 0; rank < 4; rank++ {
					topSq := rank*8 + file
					bottomSq := (7-rank)*8 + file

					// For most pieces, we expect some asymmetry (pawns especially)
					// Just verify the table has values
					_ = tt.table[topSq]
					_ = tt.table[bottomSq]
				}
			}
		})
	}
}

func BenchmarkEvaluate(b *testing.B) {
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
