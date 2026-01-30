package daemon

import (
	"testing"
)

func TestAPILogger(t *testing.T) {
	log := NewAPILogger()

	log.Info("info message")
	log.Success("success message")
	log.Warn("warning message")
	log.Error("error message")
	log.Debug("debug message")

	logs := log.GetLogs()
	if len(logs) != 5 {
		t.Errorf("Expected 5 log entries, got %d", len(logs))
	}

	expectedLevels := []string{"info", "success", "warn", "error", "debug"}
	for i, entry := range logs {
		if entry.Level != expectedLevels[i] {
			t.Errorf("Entry %d: expected level '%s', got '%s'", i, expectedLevels[i], entry.Level)
		}
		if entry.Timestamp == "" {
			t.Errorf("Entry %d: timestamp is empty", i)
		}
		if entry.Message == "" {
			t.Errorf("Entry %d: message is empty", i)
		}
	}
}

func TestAPILoggerFormatting(t *testing.T) {
	log := NewAPILogger()

	log.Info("test %s %d", "string", 42)

	logs := log.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	expected := "test string 42"
	if logs[0].Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, logs[0].Message)
	}
}

func TestAPILoggerBufferLimit(t *testing.T) {
	log := NewAPILogger()
	log.maxSize = 10

	// Add 15 entries (exceeds maxSize)
	for i := 0; i < 15; i++ {
		log.Info("message %d", i)
	}

	logs := log.GetLogs()
	if len(logs) != 10 {
		t.Errorf("Expected 10 log entries (maxSize), got %d", len(logs))
	}

	// Should keep the last 10 entries (5-14)
	if logs[0].Message != "message 5" {
		t.Errorf("Expected oldest entry to be 'message 5', got '%s'", logs[0].Message)
	}
	if logs[9].Message != "message 14" {
		t.Errorf("Expected newest entry to be 'message 14', got '%s'", logs[9].Message)
	}
}

func TestAPILoggerConcurrency(t *testing.T) {
	log := NewAPILogger()
	done := make(chan bool)

	// Run concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			log.Info("message from goroutine %d", n)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	logs := log.GetLogs()
	if len(logs) != 10 {
		t.Errorf("Expected 10 log entries, got %d", len(logs))
	}
}

func TestAPILoggerCopy(t *testing.T) {
	log := NewAPILogger()
	log.Info("message 1")

	logs1 := log.GetLogs()
	log.Info("message 2")
	logs2 := log.GetLogs()

	if len(logs1) == len(logs2) {
		t.Error("Expected GetLogs to return a copy, not a reference")
	}
}
