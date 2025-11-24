package utils

import (
	"runtime"
	"testing"
)

func TestOpenBrowser(t *testing.T) {
	t.Skip("Skipping browser tests - requires user interaction and won't work in CI/CD")

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			url:     "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: false, // Command will still execute, just with empty arg
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OpenBrowser(tt.url)

			// On unsupported platforms, we expect an error
			if runtime.GOOS != "linux" && runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
				if err == nil {
					t.Error("Expected error on unsupported platform, got nil")
				}
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("OpenBrowser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenBrowserUnsupportedPlatform(t *testing.T) {
	t.Skip("Skipping browser tests - requires user interaction and won't work in CI/CD")

	// This test verifies the error message for unsupported platforms
	// We can't actually test this without mocking runtime.GOOS
	// But we can verify the function exists and has the right signature

	// Just ensure the function is callable
	_ = OpenBrowser("http://test.com")
}
