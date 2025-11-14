package server

import (
	"fmt"
	"log"
	"strings"
)

// Log levels
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

// Log level priorities (lower number = more verbose)
var logLevelPriority = map[string]int{
	LogLevelDebug: 0,
	LogLevelInfo:  1,
	LogLevelWarn:  2,
	LogLevelError: 3,
}

// Current log level (default INFO)
var currentLogLevel = LogLevelInfo

// SetLogLevel sets the minimum log level to display
func SetLogLevel(level string) {
	level = strings.ToUpper(level)
	if _, ok := logLevelPriority[level]; ok {
		currentLogLevel = level
	}
}

// Logger helper functions for consistent logging format
// Format: YYYY/MM/DD HH:MM:SS.microseconds LEVEL [component] message

func logMessage(level, component, format string, args ...interface{}) {
	// Filter based on log level
	if logLevelPriority[level] < logLevelPriority[currentLogLevel] {
		return
	}

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
