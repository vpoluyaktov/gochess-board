package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestSetLogLevel(t *testing.T) {
	// Save original log level
	originalLevel := currentLogLevel
	defer func() { currentLogLevel = originalLevel }()

	tests := []struct {
		name          string
		level         string
		expectedLevel string
	}{
		{
			name:          "set to DEBUG",
			level:         "DEBUG",
			expectedLevel: LogLevelDebug,
		},
		{
			name:          "set to INFO",
			level:         "INFO",
			expectedLevel: LogLevelInfo,
		},
		{
			name:          "set to WARN",
			level:         "WARN",
			expectedLevel: LogLevelWarn,
		},
		{
			name:          "set to ERROR",
			level:         "ERROR",
			expectedLevel: LogLevelError,
		},
		{
			name:          "lowercase debug",
			level:         "debug",
			expectedLevel: LogLevelDebug,
		},
		{
			name:          "mixed case info",
			level:         "InFo",
			expectedLevel: LogLevelInfo,
		},
		{
			name:          "invalid level - should not change",
			level:         "INVALID",
			expectedLevel: originalLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to original before each test
			currentLogLevel = originalLevel

			SetLogLevel(tt.level)

			if currentLogLevel != tt.expectedLevel {
				t.Errorf("SetLogLevel(%q) = %q, want %q", tt.level, currentLogLevel, tt.expectedLevel)
			}
		})
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// Save original log level
	originalLevel := currentLogLevel
	defer func() { currentLogLevel = originalLevel }()

	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	tests := []struct {
		name         string
		setLevel     string
		logFunc      func()
		shouldLog    bool
		expectedText string
	}{
		{
			name:     "DEBUG level - should log debug",
			setLevel: LogLevelDebug,
			logFunc: func() {
				Debug("TEST", "debug message")
			},
			shouldLog:    true,
			expectedText: "DEBUG",
		},
		{
			name:     "INFO level - should not log debug",
			setLevel: LogLevelInfo,
			logFunc: func() {
				Debug("TEST", "debug message")
			},
			shouldLog: false,
		},
		{
			name:     "INFO level - should log info",
			setLevel: LogLevelInfo,
			logFunc: func() {
				Info("TEST", "info message")
			},
			shouldLog:    true,
			expectedText: "INFO",
		},
		{
			name:     "WARN level - should not log info",
			setLevel: LogLevelWarn,
			logFunc: func() {
				Info("TEST", "info message")
			},
			shouldLog: false,
		},
		{
			name:     "WARN level - should log warn",
			setLevel: LogLevelWarn,
			logFunc: func() {
				Warn("TEST", "warning message")
			},
			shouldLog:    true,
			expectedText: "WARN",
		},
		{
			name:     "ERROR level - should log error",
			setLevel: LogLevelError,
			logFunc: func() {
				Error("TEST", "error message")
			},
			shouldLog:    true,
			expectedText: "ERROR",
		},
		{
			name:     "ERROR level - should not log warn",
			setLevel: LogLevelError,
			logFunc: func() {
				Warn("TEST", "warning message")
			},
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			SetLogLevel(tt.setLevel)

			tt.logFunc()

			output := buf.String()
			if tt.shouldLog {
				if output == "" {
					t.Errorf("Expected log output but got none")
				}
				if tt.expectedText != "" && !strings.Contains(output, tt.expectedText) {
					t.Errorf("Expected output to contain %q, got: %q", tt.expectedText, output)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no log output but got: %q", output)
				}
			}
		})
	}
}

func TestLogMessageFormat(t *testing.T) {
	// Save original log level
	originalLevel := currentLogLevel
	defer func() { currentLogLevel = originalLevel }()

	SetLogLevel(LogLevelDebug)

	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	tests := []struct {
		name          string
		logFunc       func()
		expectedLevel string
		expectedComp  string
		expectedMsg   string
	}{
		{
			name: "Debug message",
			logFunc: func() {
				Debug("COMPONENT", "test message")
			},
			expectedLevel: "DEBUG",
			expectedComp:  "COMPONENT",
			expectedMsg:   "test message",
		},
		{
			name: "Info message",
			logFunc: func() {
				Info("SERVER", "server started")
			},
			expectedLevel: "INFO",
			expectedComp:  "SERVER",
			expectedMsg:   "server started",
		},
		{
			name: "Warn message",
			logFunc: func() {
				Warn("ENGINE", "engine slow")
			},
			expectedLevel: "WARN",
			expectedComp:  "ENGINE",
			expectedMsg:   "engine slow",
		},
		{
			name: "Error message",
			logFunc: func() {
				Error("DATABASE", "connection failed")
			},
			expectedLevel: "ERROR",
			expectedComp:  "DATABASE",
			expectedMsg:   "connection failed",
		},
		{
			name: "Message with formatting",
			logFunc: func() {
				Info("TEST", "value: %d, name: %s", 42, "test")
			},
			expectedLevel: "INFO",
			expectedComp:  "TEST",
			expectedMsg:   "value: 42, name: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()

			tt.logFunc()

			output := buf.String()
			if !strings.Contains(output, tt.expectedLevel) {
				t.Errorf("Expected output to contain level %q, got: %q", tt.expectedLevel, output)
			}
			if !strings.Contains(output, tt.expectedComp) {
				t.Errorf("Expected output to contain component %q, got: %q", tt.expectedComp, output)
			}
			if !strings.Contains(output, tt.expectedMsg) {
				t.Errorf("Expected output to contain message %q, got: %q", tt.expectedMsg, output)
			}
		})
	}
}

func TestLogLevelPriorities(t *testing.T) {
	// Verify the priority order
	if logLevelPriority[LogLevelDebug] >= logLevelPriority[LogLevelInfo] {
		t.Error("DEBUG should have lower priority than INFO")
	}
	if logLevelPriority[LogLevelInfo] >= logLevelPriority[LogLevelWarn] {
		t.Error("INFO should have lower priority than WARN")
	}
	if logLevelPriority[LogLevelWarn] >= logLevelPriority[LogLevelError] {
		t.Error("WARN should have lower priority than ERROR")
	}
}
