package daemon

import (
	"fmt"
	"sync"
	"time"
)

// APILogger captures log entries for structured JSON responses
type APILogger struct {
	entries []LogEntry
	mu      sync.Mutex
	maxSize int
}

// NewAPILogger creates a new API logger
func NewAPILogger() *APILogger {
	return &APILogger{
		entries: make([]LogEntry, 0, 1000),
		maxSize: 1000,
	}
}

// Info logs an info message
func (l *APILogger) Info(format string, args ...interface{}) {
	l.log("info", format, args...)
}

// Success logs a success message
func (l *APILogger) Success(format string, args ...interface{}) {
	l.log("success", format, args...)
}

// Warn logs a warning message
func (l *APILogger) Warn(format string, args ...interface{}) {
	l.log("warn", format, args...)
}

// Error logs an error message
func (l *APILogger) Error(format string, args ...interface{}) {
	l.log("error", format, args...)
}

// Debug logs a debug message
func (l *APILogger) Debug(format string, args ...interface{}) {
	l.log("debug", format, args...)
}

// log is the internal logging implementation
func (l *APILogger) log(level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   fmt.Sprintf(format, args...),
	}

	l.entries = append(l.entries, entry)

	// Limit buffer size
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
}

// GetLogs returns all captured log entries
func (l *APILogger) GetLogs() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Return a copy to prevent external modification
	logs := make([]LogEntry, len(l.entries))
	copy(logs, l.entries)
	return logs
}
