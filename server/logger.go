package server

import (
	"fmt"
	"log"
)

// Log levels
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

// Logger helper functions for consistent logging format
// Format: YYYY/MM/DD HH:MM:SS.microseconds LEVEL [component] message

func logMessage(level, component, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("%-5s [%s] %s", level, component, message)
}

// Debug logs a debug message
func Debug(component, format string, args ...interface{}) {
	logMessage(LogLevelDebug, component, format, args...)
}

// Info logs an info message
func Info(component, format string, args ...interface{}) {
	logMessage(LogLevelInfo, component, format, args...)
}

// Warn logs a warning message
func Warn(component, format string, args ...interface{}) {
	logMessage(LogLevelWarn, component, format, args...)
}

// Error logs an error message
func Error(component, format string, args ...interface{}) {
	logMessage(LogLevelError, component, format, args...)
}
