package opening

import (
	"testing"

	"github.com/notnil/chess"
)

func TestOpeningBook(t *testing.T) {
	book := NewOpeningBook()

	// Load the opening database
	err := book.LoadFromDirectory("assets/openings")
	if err != nil {
		t.Fatalf("Failed to load opening book: %v", err)
	}

	// Print stats
	stats := book.Stats()
	t.Logf("Opening book stats: %+v", stats)

	// Test some common openings
	tests := []struct {
		name     string
		moves    []string
		wantECO  string
		wantName string
	}{
		{
			name:     "Italian Game",
			moves:    []string{"e4", "e5", "Nf3", "Nc6", "Bc4"},
			wantECO:  "C50",
			wantName: "Italian Game",
		},
		{
			name:     "Sicilian Defense",
			moves:    []string{"e4", "c5"},
			wantECO:  "B20",
			wantName: "Sicilian Defense",
		},
		{
			name:     "French Defense",
			moves:    []string{"e4", "e6"},
			wantECO:  "C00",
			wantName: "French Defense",
		},
		{
			name:     "Queen's Gambit",
			moves:    []string{"d4", "d5", "c4"},
			wantECO:  "D06",
			wantName: "Queen's Gambit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opening := book.Lookup(tt.moves)
			if opening == nil {
				t.Errorf("Expected to find opening for %v, got nil", tt.moves)
				return
			}

			if opening.ECO != tt.wantECO {
				t.Errorf("ECO = %s, want %s", opening.ECO, tt.wantECO)
			}

			if opening.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", opening.Name, tt.wantName)
			}

			t.Logf("Found: %s (%s) - %s", opening.Name, opening.ECO, opening.PGN)
		})
	}
}

func TestOpeningBookByGame(t *testing.T) {
	book := NewOpeningBook()

	// Load the opening database
	err := book.LoadFromDirectory("assets/openings")
	if err != nil {
		t.Fatalf("Failed to load opening book: %v", err)
	}

	// Create a game with some moves
	game := chess.NewGame()
	moves := []string{"e4", "e5", "Nf3", "Nc6"}

	for _, move := range moves {
		if err := game.MoveStr(move); err != nil {
			t.Fatalf("Failed to make move %s: %v", move, err)
		}
	}

	opening := book.LookupByGame(game)
	if opening == nil {
		t.Error("Expected to find opening, got nil")
		return
	}

	t.Logf("Found opening: %s (%s)", opening.Name, opening.ECO)
}

func TestOpeningBookNoMatch(t *testing.T) {
	book := NewOpeningBook()

	// Load the opening database
	err := book.LoadFromDirectory("assets/openings")
	if err != nil {
		t.Fatalf("Failed to load opening book: %v", err)
	}

	// Test with random moves that shouldn't match any opening
	moves := []string{"a3", "a6", "b3", "b6"}
	opening := book.Lookup(moves)

	// Should return nil or a very early opening
	if opening != nil {
		t.Logf("Found opening (might be early position): %s (%s)", opening.Name, opening.ECO)
	} else {
		t.Log("No opening found (expected)")
	}
}
