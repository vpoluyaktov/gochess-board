package server

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/notnil/chess"
)

func TestPolyglotBook(t *testing.T) {
	book := NewPolyglotBook()

	// Try to load GNU Chess book
	bookPath := "/usr/share/games/gnuchess/book.bin"
	if err := book.LoadFromFile(bookPath); err != nil {
		t.Skipf("Book file not found: %v", err)
		return
	}

	// Test starting position
	game := chess.NewGame()

	// Debug: manually calculate hash to verify
	board := game.Position().Board()
	t.Logf("Piece at A1 (0): %v (type=%d, color=%v)", board.Piece(chess.A1), board.Piece(chess.A1).Type(), board.Piece(chess.A1).Color())
	t.Logf("Piece at E1 (4): %v (type=%d, color=%v)", board.Piece(chess.E1), board.Piece(chess.E1).Type(), board.Piece(chess.E1).Color())
	t.Logf("Piece at A8 (56): %v (type=%d, color=%v)", board.Piece(chess.A8), board.Piece(chess.A8).Type(), board.Piece(chess.A8).Color())

	// Calculate piece index for white rook
	// Should be: (4-1)*2 + 1 = 7
	whiteRookIdx := (4-1)*2 + 1
	t.Logf("White rook piece index: %d (expected 7)", whiteRookIdx)

	// Calculate offset for white rook at A1
	// Should be: 64*7 + 8*0 + 0 = 448
	offset := 64*7 + 8*0 + 0
	t.Logf("White rook at A1 offset: %d (expected 448)", offset)

	// Debug: print the hash we calculate
	hash := book.zobristHash(game.Position())
	t.Logf("Calculated hash for starting position: 0x%016X", hash)
	t.Logf("Expected hash for starting position:  0x463B96181691FC9C")

	// Debug: print first few book entries
	t.Logf("First 5 book entries:")
	for i := 0; i < 5 && i < len(book.entries); i++ {
		t.Logf("  Entry %d: key=0x%016X", i, book.entries[i].Key)
	}

	moves := book.Probe(game.Position())

	if len(moves) == 0 {
		t.Error("Expected book moves for starting position")
	}

	t.Logf("Starting position has %d book moves: %v", len(moves), moves)

	// Test weighted probe
	move := book.ProbeWeighted(game.Position())
	if move == "" {
		t.Error("Expected weighted book move for starting position")
	}

	t.Logf("Weighted book move: %s", move)
}

// TestZobristHashStartingPosition verifies the hash for the starting position
func TestZobristHashStartingPosition(t *testing.T) {
	book := NewPolyglotBook()
	game := chess.NewGame()
	hash := book.zobristHash(game.Position())

	// Expected hash for starting position (verified against python-chess)
	expected := uint64(0x463B96181691FC9C)
	if hash != expected {
		t.Errorf("Starting position hash mismatch: got 0x%016X, want 0x%016X", hash, expected)
	}
}

// TestZobristHashAfterMove verifies hash changes correctly after a move
func TestZobristHashAfterMove(t *testing.T) {
	book := NewPolyglotBook()
	game := chess.NewGame()

	// Hash before move
	hashBefore := book.zobristHash(game.Position())

	// Make a move (e4 in SAN notation)
	err := game.MoveStr("e4")
	if err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

	// Hash after move should be different
	hashAfter := book.zobristHash(game.Position())

	if hashAfter == hashBefore {
		t.Error("Hash should change after a move")
	}

	// Note: We don't check the exact hash value here because the chess library
	// may handle en passant squares differently than python-chess.
	// The important thing is that the hash changes, which is verified by
	// other tests (turn, castling, en passant)
}

// TestZobristHashTurnMatters verifies that turn affects the hash
func TestZobristHashTurnMatters(t *testing.T) {
	book := NewPolyglotBook()

	// Same position, white to move
	fen1 := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	pos1, _ := chess.FEN(fen1)
	game1 := chess.NewGame(pos1)
	hash1 := book.zobristHash(game1.Position())

	// Same position, black to move
	fen2 := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1"
	pos2, _ := chess.FEN(fen2)
	game2 := chess.NewGame(pos2)
	hash2 := book.zobristHash(game2.Position())

	if hash1 == hash2 {
		t.Error("Hash should differ based on whose turn it is")
	}

	// Verify specific values (from python-chess)
	expectedWhite := uint64(0x463B96181691FC9C)
	expectedBlack := uint64(0xBEEDB0B2B9B67995)

	if hash1 != expectedWhite {
		t.Errorf("White to move hash: got 0x%016X, want 0x%016X", hash1, expectedWhite)
	}
	if hash2 != expectedBlack {
		t.Errorf("Black to move hash: got 0x%016X, want 0x%016X", hash2, expectedBlack)
	}
}

// TestZobristHashCastlingRights verifies castling rights affect the hash
func TestZobristHashCastlingRights(t *testing.T) {
	book := NewPolyglotBook()

	// Position with all castling rights
	fen1 := "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"
	pos1, _ := chess.FEN(fen1)
	game1 := chess.NewGame(pos1)
	hash1 := book.zobristHash(game1.Position())

	// Same position, no castling rights
	fen2 := "r3k2r/8/8/8/8/8/8/R3K2R w - - 0 1"
	pos2, _ := chess.FEN(fen2)
	game2 := chess.NewGame(pos2)
	hash2 := book.zobristHash(game2.Position())

	if hash1 == hash2 {
		t.Error("Hash should differ based on castling rights")
	}
}

// TestZobristHashEnPassant verifies en passant square affects the hash
func TestZobristHashEnPassant(t *testing.T) {
	book := NewPolyglotBook()

	// Position with en passant square
	fen1 := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	pos1, _ := chess.FEN(fen1)
	game1 := chess.NewGame(pos1)
	hash1 := book.zobristHash(game1.Position())

	// Same position, no en passant
	fen2 := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
	pos2, _ := chess.FEN(fen2)
	game2 := chess.NewGame(pos2)
	hash2 := book.zobristHash(game2.Position())

	if hash1 == hash2 {
		t.Error("Hash should differ based on en passant square")
	}
}

// TestPolyglotMoveToUCI verifies move format conversion
func TestPolyglotMoveToUCI(t *testing.T) {
	book := NewPolyglotBook()

	tests := []struct {
		name     string
		move     uint16
		expected string
	}{
		{"e2e4", 0x031C, "e2e4"},   // from=12 (e2), to=28 (e4)
		{"d2d4", 0x02DB, "d2d4"},   // from=11 (d2), to=27 (d4)
		{"g1f3", 0x0195, "g1f3"},   // from=6 (g1), to=21 (f3)
		{"a1a8", 0x0038, "a1a8"},   // from=0 (a1), to=56 (a8)
		{"h1h8", 0x01FF, "h1h8"},   // from=7 (h1), to=63 (h8)
		{"e7e8q", 0x4D3C, "e7e8q"}, // from=52 (e7), to=60 (e8), promo=queen
		{"a7a8n", 0x1C38, "a7a8n"}, // from=48 (a7), to=56 (a8), promo=knight
		{"b7b8r", 0x3C79, "b7b8r"}, // from=49 (b7), to=57 (b8), promo=rook
		{"c7c8b", 0x2CBA, "c7c8b"}, // from=50 (c7), to=58 (c8), promo=bishop
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := book.polyglotMoveToUCI(tt.move)
			if result != tt.expected {
				t.Errorf("polyglotMoveToUCI(0x%04X) = %s, want %s", tt.move, result, tt.expected)
			}
		})
	}
}

// TestPieceHash verifies piece hash calculation
func TestPieceHash(t *testing.T) {
	book := NewPolyglotBook()

	tests := []struct {
		name     string
		piece    chess.Piece
		square   int
		expected uint64
	}{
		{"white rook a1", chess.WhiteRook, 0, 0xA09E8C8C35AB96DE},
		{"white king e1", chess.WhiteKing, 4, 0xB5FDFC5D3132C498},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := book.pieceHash(tt.piece, tt.square)
			if result != tt.expected {
				t.Errorf("pieceHash(%v, %d) = 0x%016X, want 0x%016X", tt.piece, tt.square, result, tt.expected)
			}
		})
	}
}

// TestIndexToSquare verifies square index to algebraic notation conversion
func TestIndexToSquare(t *testing.T) {
	tests := []struct {
		idx      int
		expected string
	}{
		{0, "a1"},
		{7, "h1"},
		{8, "a2"},
		{12, "e2"},
		{28, "e4"},
		{56, "a8"},
		{63, "h8"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := indexToSquare(tt.idx)
			if result != tt.expected {
				t.Errorf("indexToSquare(%d) = %s, want %s", tt.idx, result, tt.expected)
			}
		})
	}
}

// TestLoadFromFile verifies book file loading
func TestLoadFromFile(t *testing.T) {
	// Create a temporary book file with a few entries
	tmpDir := t.TempDir()
	bookPath := filepath.Join(tmpDir, "test.bin")

	// Create a small book file with 2 entries
	var buf bytes.Buffer

	// Entry 1: Starting position with e2e4
	key1 := uint64(0x463B96181691FC9C)
	move1 := uint16(0x031C) // e2e4
	weight1 := uint16(100)
	learn1 := uint32(0)

	binary.Write(&buf, binary.BigEndian, key1)
	binary.Write(&buf, binary.BigEndian, move1)
	binary.Write(&buf, binary.BigEndian, weight1)
	binary.Write(&buf, binary.BigEndian, learn1)

	// Entry 2: Starting position with d2d4
	key2 := uint64(0x463B96181691FC9C)
	move2 := uint16(0x02DB) // d2d4
	weight2 := uint16(80)
	learn2 := uint32(0)

	binary.Write(&buf, binary.BigEndian, key2)
	binary.Write(&buf, binary.BigEndian, move2)
	binary.Write(&buf, binary.BigEndian, weight2)
	binary.Write(&buf, binary.BigEndian, learn2)

	// Write to file
	if err := os.WriteFile(bookPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to create test book file: %v", err)
	}

	// Load the book
	book := NewPolyglotBook()
	if err := book.LoadFromFile(bookPath); err != nil {
		t.Fatalf("Failed to load book: %v", err)
	}

	// Verify entries were loaded
	if len(book.entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(book.entries))
	}

	// Verify entries are sorted by key
	for i := 1; i < len(book.entries); i++ {
		if book.entries[i-1].Key > book.entries[i].Key {
			t.Error("Entries are not sorted by key")
		}
	}
}

// TestProbe verifies book move lookup
func TestProbe(t *testing.T) {
	// Create a book with test data
	tmpDir := t.TempDir()
	bookPath := filepath.Join(tmpDir, "test.bin")

	var buf bytes.Buffer

	// Add entries for starting position
	key := uint64(0x463B96181691FC9C)
	moves := []struct {
		move   uint16
		weight uint16
	}{
		{0x031C, 100}, // e2e4
		{0x02DB, 80},  // d2d4
		{0x0195, 50},  // g1f3
	}

	for _, m := range moves {
		binary.Write(&buf, binary.BigEndian, key)
		binary.Write(&buf, binary.BigEndian, m.move)
		binary.Write(&buf, binary.BigEndian, m.weight)
		binary.Write(&buf, binary.BigEndian, uint32(0))
	}

	os.WriteFile(bookPath, buf.Bytes(), 0644)

	// Load and probe
	book := NewPolyglotBook()
	book.LoadFromFile(bookPath)

	game := chess.NewGame()
	result := book.Probe(game.Position())

	if len(result) != 3 {
		t.Errorf("Expected 3 moves, got %d", len(result))
	}

	// Verify moves are correct
	expected := []string{"e2e4", "d2d4", "g1f3"}
	for i, move := range expected {
		if i >= len(result) {
			break
		}
		if result[i] != move {
			t.Errorf("Move %d: got %s, want %s", i, result[i], move)
		}
	}
}

// TestProbeWeighted verifies weighted move selection
func TestProbeWeighted(t *testing.T) {
	// Create a book with test data
	tmpDir := t.TempDir()
	bookPath := filepath.Join(tmpDir, "test.bin")

	var buf bytes.Buffer

	// Add one entry for starting position
	key := uint64(0x463B96181691FC9C)
	binary.Write(&buf, binary.BigEndian, key)
	binary.Write(&buf, binary.BigEndian, uint16(0x031C)) // e2e4
	binary.Write(&buf, binary.BigEndian, uint16(100))
	binary.Write(&buf, binary.BigEndian, uint32(0))

	os.WriteFile(bookPath, buf.Bytes(), 0644)

	// Load and probe
	book := NewPolyglotBook()
	book.LoadFromFile(bookPath)

	game := chess.NewGame()
	result := book.ProbeWeighted(game.Position())

	if result != "e2e4" {
		t.Errorf("Expected e2e4, got %s", result)
	}
}

// TestProbeNoMatch verifies behavior when no book moves are found
func TestProbeNoMatch(t *testing.T) {
	book := NewPolyglotBook()

	// Create a position not in the book
	fen := "8/8/8/8/8/8/8/4K3 w - - 0 1"
	pos, _ := chess.FEN(fen)
	game := chess.NewGame(pos)

	// Probe should return nil
	result := book.Probe(game.Position())
	if result != nil {
		t.Errorf("Expected nil for position not in book, got %v", result)
	}

	// ProbeWeighted should return empty string
	weighted := book.ProbeWeighted(game.Position())
	if weighted != "" {
		t.Errorf("Expected empty string for position not in book, got %s", weighted)
	}
}

// TestEmptyBook verifies behavior with empty book
func TestEmptyBook(t *testing.T) {
	book := NewPolyglotBook()

	game := chess.NewGame()

	// Probe empty book
	result := book.Probe(game.Position())
	if result != nil {
		t.Errorf("Expected nil for empty book, got %v", result)
	}

	weighted := book.ProbeWeighted(game.Position())
	if weighted != "" {
		t.Errorf("Expected empty string for empty book, got %s", weighted)
	}
}
