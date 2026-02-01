package daemon

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ServerPort != 9999 {
		t.Errorf("expected ServerPort to be 9999, got %d", cfg.ServerPort)
	}

	if !cfg.ServerEnabled {
		t.Error("expected ServerEnabled to be true")
	}

	if cfg.ServerBindAddress != "127.0.0.1" {
		t.Errorf("expected ServerBindAddress to be 127.0.0.1, got %s", cfg.ServerBindAddress)
	}
}
