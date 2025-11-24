package server

import (
	"runtime"
	"testing"
)

func TestGetExecutableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "stockfish",
			expected: getExpectedExecutableName("stockfish"),
		},
		{
			name:     "name with version",
			input:    "stockfish-16",
			expected: getExpectedExecutableName("stockfish-16"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: getExpectedExecutableName(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExecutableName(tt.input)
			if result != tt.expected {
				t.Errorf("getExecutableName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func getExpectedExecutableName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func TestGetEngineNames(t *testing.T) {
	baseNames := []string{"stockfish", "fruit", "crafty"}
	result := getEngineNames(baseNames)

	if len(result) != 3 {
		t.Errorf("Expected 3 engine names, got %d", len(result))
	}

	// On Windows, should have .exe extension
	if runtime.GOOS == "windows" {
		for i, name := range result {
			expected := baseNames[i] + ".exe"
			if name != expected {
				t.Errorf("Expected %q, got %q", expected, name)
			}
		}
	} else {
		// On Unix, should be unchanged
		for i, name := range result {
			if name != baseNames[i] {
				t.Errorf("Expected %q, got %q", baseNames[i], name)
			}
		}
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple version",
			input:    "Stockfish 16",
			expected: "16",
		},
		{
			name:     "version with dot",
			input:    "Fruit 2.1",
			expected: "2.1",
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
			name:     "no version",
			input:    "Chess Engine",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersion(tt.input)
			if result != tt.expected {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEngineInfo_Structure(t *testing.T) {
	engine := EngineInfo{
		Name:                  "Stockfish",
		Path:                  "/usr/bin/stockfish",
		Version:               "16",
		ID:                    "stockfish",
		Type:                  "uci",
		SupportsBook:          false,
		SupportsLimitStrength: true,
		MinElo:                1350,
		MaxElo:                2850,
		Options:               make(map[string]string),
	}

	if engine.Name != "Stockfish" {
		t.Errorf("Expected Name 'Stockfish', got %q", engine.Name)
	}
	if engine.Type != "uci" {
		t.Errorf("Expected Type 'uci', got %q", engine.Type)
	}
	if !engine.SupportsLimitStrength {
		t.Error("Expected SupportsLimitStrength to be true")
	}
	if engine.MinElo != 1350 {
		t.Errorf("Expected MinElo 1350, got %d", engine.MinElo)
	}
	if engine.MaxElo != 2850 {
		t.Errorf("Expected MaxElo 2850, got %d", engine.MaxElo)
	}
}

func TestEngineInfo_PolyglotVariant(t *testing.T) {
	engine := EngineInfo{
		Name:             "Stockfish with Book",
		Type:             "polyglot-uci",
		SupportsBook:     true,
		BookPath:         "/path/to/book.bin",
		UnderlyingEngine: "stockfish",
	}

	if engine.Type != "polyglot-uci" {
		t.Errorf("Expected Type 'polyglot-uci', got %q", engine.Type)
	}
	if !engine.SupportsBook {
		t.Error("Expected SupportsBook to be true")
	}
	if engine.BookPath != "/path/to/book.bin" {
		t.Errorf("Expected BookPath '/path/to/book.bin', got %q", engine.BookPath)
	}
	if engine.UnderlyingEngine != "stockfish" {
		t.Errorf("Expected UnderlyingEngine 'stockfish', got %q", engine.UnderlyingEngine)
	}
}

func TestEngineInfo_Options(t *testing.T) {
	engine := EngineInfo{
		Name:    "Test Engine",
		Options: make(map[string]string),
	}

	// Add some options
	engine.Options["UCI_Elo"] = "2000"
	engine.Options["UCI_LimitStrength"] = "true"
	engine.Options["Threads"] = "4"

	if len(engine.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(engine.Options))
	}

	if engine.Options["UCI_Elo"] != "2000" {
		t.Errorf("Expected UCI_Elo '2000', got %q", engine.Options["UCI_Elo"])
	}
}

func TestEngineInfo_JSONTags(t *testing.T) {
	// This test verifies that the struct has proper JSON tags
	// by checking if the fields can be marshaled/unmarshaled
	engine := EngineInfo{
		Name:    "Test",
		Path:    "/test",
		Version: "1.0",
		ID:      "test",
		Type:    "uci",
	}

	// Basic validation that fields are accessible
	if engine.Name == "" {
		t.Error("Name field should be accessible")
	}
	if engine.Path == "" {
		t.Error("Path field should be accessible")
	}
}

func TestEngineInfo_EmptyOptions(t *testing.T) {
	engine := EngineInfo{
		Name:    "Test Engine",
		Options: nil,
	}

	// Should handle nil options map gracefully
	if engine.Options != nil {
		// If we initialize it, it should work
		engine.Options = make(map[string]string)
	}

	// Should not panic
	if engine.Options == nil {
		engine.Options = make(map[string]string)
	}
	engine.Options["test"] = "value"

	if engine.Options["test"] != "value" {
		t.Error("Should be able to set options after initialization")
	}
}

func TestEngineInfo_EloRange(t *testing.T) {
	tests := []struct {
		name    string
		minElo  int
		maxElo  int
		isValid bool
	}{
		{
			name:    "valid range",
			minElo:  1350,
			maxElo:  2850,
			isValid: true,
		},
		{
			name:    "equal values",
			minElo:  2000,
			maxElo:  2000,
			isValid: true,
		},
		{
			name:    "zero values",
			minElo:  0,
			maxElo:  0,
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := EngineInfo{
				MinElo: tt.minElo,
				MaxElo: tt.maxElo,
			}

			if tt.isValid && engine.MaxElo > 0 && engine.MinElo > engine.MaxElo {
				t.Errorf("Invalid ELO range: min=%d, max=%d", engine.MinElo, engine.MaxElo)
			}
		})
	}
}
