package utils

import "testing"

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple version with space",
			input:    "Stockfish 16",
			expected: "16",
		},
		{
			name:     "version with dot",
			input:    "Fruit 2.1",
			expected: "2.1",
		},
		{
			name:     "version with roman numerals",
			input:    "Toga II 3.0",
			expected: "3.0",
		},
		{
			name:     "version with hyphen",
			input:    "Crafty-23.4",
			expected: "23.4",
		},
		{
			name:     "version with v prefix",
			input:    "Engine v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with V prefix",
			input:    "Engine V2.0",
			expected: "2.0",
		},
		{
			name:     "complex version",
			input:    "GNU Chess 6.2.9",
			expected: "6.2.9",
		},
		{
			name:     "no version",
			input:    "Chess Engine",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only numbers",
			input:    "123",
			expected: "123",
		},
		{
			name:     "version at start",
			input:    "16 Stockfish",
			expected: "16",
		},
		{
			name:     "multiple versions",
			input:    "Engine 1.0 Beta 2",
			expected: "2",
		},
		{
			name:     "version with build number",
			input:    "Stockfish-16.1.2345",
			expected: "16.1.2345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVersion(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase word",
			input:    "stockfish",
			expected: "Stockfish",
		},
		{
			name:     "hyphenated words",
			input:    "chess-engine",
			expected: "Chess Engine",
		},
		{
			name:     "multiple hyphens",
			input:    "my-chess-engine",
			expected: "My Chess Engine",
		},
		{
			name:     "already capitalized",
			input:    "Stockfish",
			expected: "Stockfish",
		},
		{
			name:     "mixed case",
			input:    "stockFish-engine",
			expected: "StockFish Engine",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "single hyphen",
			input:    "-",
			expected: " ",
		},
		{
			name:     "trailing hyphen",
			input:    "engine-",
			expected: "Engine ",
		},
		{
			name:     "leading hyphen",
			input:    "-engine",
			expected: " Engine",
		},
		{
			name:     "numbers",
			input:    "engine-16",
			expected: "Engine 16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TitleCase(tt.input)
			if result != tt.expected {
				t.Errorf("TitleCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkExtractVersion(b *testing.B) {
	testCases := []string{
		"Stockfish 16",
		"Crafty-23.4",
		"GNU Chess 6.2.9",
		"Engine v1.2.3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			ExtractVersion(tc)
		}
	}
}

func BenchmarkTitleCase(b *testing.B) {
	testCases := []string{
		"stockfish",
		"chess-engine",
		"my-chess-engine-name",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			TitleCase(tc)
		}
	}
}
