package server

import (
	"testing"
)

func TestOpeningBookUnknownPosition(t *testing.T) {
	book := NewOpeningBook()

	// Load the opening database
	err := book.LoadFromDirectory("assets/openings")
	if err != nil {
		t.Fatalf("Failed to load opening book: %v", err)
	}

	tests := []struct {
		name        string
		moves       []string
		shouldFind  bool
		description string
	}{
		{
			name:        "Italian Game - Known Opening",
			moves:       []string{"e4", "e5", "Nf3", "Nc6", "Bc4"},
			shouldFind:  true,
			description: "Standard Italian Game opening",
		},
		{
			name:        "Italian Game + Unknown Move",
			moves:       []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Bc5", "h4"},
			shouldFind:  false,
			description: "Italian Game followed by unusual move h4 - should return nil",
		},
		{
			name:        "Sicilian + Random Moves",
			moves:       []string{"e4", "c5", "Nf3", "d6", "d4", "cxd4", "Nxd4", "Nf6", "Nc3", "a6", "h4", "h5"},
			shouldFind:  false,
			description: "Sicilian followed by random moves - should return nil",
		},
		{
			name:        "Unknown Opening From Start",
			moves:       []string{"a3", "a6", "b3", "b6", "c3", "c6"},
			shouldFind:  false,
			description: "Completely unknown opening - should return nil",
		},
		{
			name:        "Single Known Move",
			moves:       []string{"e4"},
			shouldFind:  true,
			description: "First move of many openings",
		},
		{
			name:        "Known Opening Exact",
			moves:       []string{"e4", "e5", "Nf3"},
			shouldFind:  true,
			description: "Exact match in opening book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opening := book.Lookup(tt.moves)

			if tt.shouldFind {
				if opening == nil {
					t.Errorf("Expected to find opening, got nil. Description: %s", tt.description)
				} else {
					t.Logf("✓ Found: %s (%s)", opening.Name, opening.ECO)
				}
			} else {
				if opening != nil {
					t.Errorf("Expected nil (unknown position), but found: %s (%s). Description: %s",
						opening.Name, opening.ECO, tt.description)
				} else {
					t.Logf("✓ Correctly returned nil for unknown position")
				}
			}
		})
	}
}
