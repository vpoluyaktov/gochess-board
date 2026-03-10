package opening

import (
	"testing"
)

// BenchmarkOpeningBookLoad measures the time to load the opening database
func BenchmarkOpeningBookLoad(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		book := NewOpeningBook()
		if err := book.LoadFromDirectory("../server/assets/openings"); err != nil {
			b.Fatalf("Failed to load opening book: %v", err)
		}
	}
}

// BenchmarkOpeningBookLookup measures the time to lookup openings
func BenchmarkOpeningBookLookup(b *testing.B) {
	book := NewOpeningBook()
	if err := book.LoadFromDirectory("../server/assets/openings"); err != nil {
		b.Fatalf("Failed to load opening book: %v", err)
	}

	moves := []string{"e4", "e5", "Nf3", "Nc6", "Bc4"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = book.Lookup(moves)
	}
}

// BenchmarkOpeningBookLookupDeep measures lookup time for deep variations
func BenchmarkOpeningBookLookupDeep(b *testing.B) {
	book := NewOpeningBook()
	if err := book.LoadFromDirectory("../server/assets/openings"); err != nil {
		b.Fatalf("Failed to load opening book: %v", err)
	}

	// Evans Gambit - deep variation
	moves := []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Bc5", "b4", "Bxb4", "c3", "Bc5", "O-O", "d6", "d4", "exd4", "cxd4", "Bb6"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = book.Lookup(moves)
	}
}

// BenchmarkParsePGNMoves measures the time to parse PGN moves
func BenchmarkParsePGNMoves(b *testing.B) {
	pgn := "1. e4 e5 2. Nf3 Nc6 3. Bc4 Bc5 4. b4 Bxb4 5. c3 Bc5 6. O-O d6 7. d4 exd4 8. cxd4 Bb6"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parsePGNMoves(pgn)
		if err != nil {
			b.Fatalf("Failed to parse PGN: %v", err)
		}
	}
}

// BenchmarkAddOpening measures the time to add a single opening to the trie
func BenchmarkAddOpening(b *testing.B) {
	book := NewOpeningBook()
	eco := "C51"
	name := "Italian Game: Evans Gambit, McDonnell Defense"
	pgn := "1. e4 e5 2. Nf3 Nc6 3. Bc4 Bc5 4. b4 Bxb4 5. c3 Bc5 6. O-O d6 7. d4 exd4 8. cxd4 Bb6"

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := book.addOpening(eco, name, pgn); err != nil {
			b.Fatalf("Failed to add opening: %v", err)
		}
	}
}
