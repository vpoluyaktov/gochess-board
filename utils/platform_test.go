package utils

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
			expected: getExpectedName("stockfish"),
		},
		{
			name:     "name with version",
			input:    "stockfish-16",
			expected: getExpectedName("stockfish-16"),
		},
		{
			name:     "name with path",
			input:    "chess-engine",
			expected: getExpectedName("chess-engine"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: getExpectedName(""),
		},
		{
			name:     "name with spaces",
			input:    "my engine",
			expected: getExpectedName("my engine"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExecutableName(tt.input)
			if result != tt.expected {
				t.Errorf("GetExecutableName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to get expected name based on platform
func getExpectedName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func TestGetExecutableNameWindows(t *testing.T) {
	// Test that on Windows, .exe is added
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	result := GetExecutableName("test")
	if result != "test.exe" {
		t.Errorf("Expected 'test.exe', got %q", result)
	}
}

func TestGetExecutableNameUnix(t *testing.T) {
	// Test that on Unix systems, no extension is added
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test")
	}

	result := GetExecutableName("test")
	if result != "test" {
		t.Errorf("Expected 'test', got %q", result)
	}
}
