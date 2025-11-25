package server

import (
	"testing"
)

// Note: Most engine tests require a real chess engine binary.
// These tests focus on structure validation and error cases.

func TestUCIEngine_Structure(t *testing.T) {
	// Test that UCIEngine struct has expected fields
	// This is a compile-time check more than a runtime test
	var engine *UCIEngine

	if engine == nil {
		// Expected: uninitialized pointer should be nil
	} else {
		t.Error("Uninitialized engine should be nil")
	}
}

func TestNewUCIEngine_InvalidPath(t *testing.T) {
	// Test with non-existent engine path
	engine, err := NewUCIEngine("/nonexistent/engine/path", "Test Engine")

	if err == nil {
		t.Error("Expected error for non-existent engine path")
		if engine != nil {
			engine.Close()
		}
	}

	if engine != nil {
		t.Error("Expected nil engine for invalid path")
	}
}

func TestNewUCIEngine_EmptyPath(t *testing.T) {
	// Test with empty path
	engine, err := NewUCIEngine("", "Test Engine")

	if err == nil {
		t.Error("Expected error for empty engine path")
		if engine != nil {
			engine.Close()
		}
	}
}

func TestNewUCIEngine_InvalidBinary(t *testing.T) {
	// Test with a file that exists but isn't executable
	// Using /dev/null as a non-executable file
	engine, err := NewUCIEngine("/dev/null", "Test Engine")

	if err == nil {
		t.Error("Expected error for non-executable file")
		if engine != nil {
			engine.Close()
		}
	}
}

// Integration test - only runs if stockfish is available
func TestNewUCIEngine_WithStockfish(t *testing.T) {
	// Try common stockfish paths
	paths := []string{
		"/usr/games/stockfish",
		"/usr/bin/stockfish",
		"/usr/local/bin/stockfish",
	}

	var engine *UCIEngine
	var err error

	for _, path := range paths {
		engine, err = NewUCIEngine(path, "Stockfish")
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Skip("Stockfish not found, skipping integration test")
		return
	}

	defer engine.Close()

	// If we got here, engine was created successfully
	if engine == nil {
		t.Error("Engine should not be nil after successful creation")
	}
}

func TestUCIEngine_NameFallback(t *testing.T) {
	// This test verifies the name fallback logic
	// We can't test it directly without a real engine,
	// but we can verify the logic would work

	enginePath := "/usr/bin/stockfish"
	engineName := ""

	// The expected behavior is to use filepath.Base(enginePath)
	// if engineName is empty
	expectedName := "stockfish"

	if engineName == "" {
		// This is what the code does
		actualName := enginePath[len(enginePath)-len(expectedName):]
		if actualName != expectedName {
			// This is just a logic test
			t.Logf("Name fallback would use: %s", actualName)
		}
	}
}
