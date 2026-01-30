package daemon

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MonitorInterval != 30*time.Second {
		t.Errorf("expected MonitorInterval to be 30s, got %v", cfg.MonitorInterval)
	}

	if cfg.ServerPort != 9999 {
		t.Errorf("expected ServerPort to be 9999, got %d", cfg.ServerPort)
	}

	if !cfg.AutoRestart {
		t.Error("expected AutoRestart to be true")
	}

	if !cfg.ServerEnabled {
		t.Error("expected ServerEnabled to be true")
	}
}

func TestMonitorStats(t *testing.T) {
	stats := &MonitorStats{
		LastCheck:      time.Now(),
		CheckCount:     10,
		RestartCount:   2,
		FailedChecks:   1,
		ContainerState: make(map[string]string),
	}

	stats.ContainerState["backend"] = "running"
	stats.ContainerState["frontend"] = "running"

	if len(stats.ContainerState) != 2 {
		t.Errorf("expected 2 container states, got %d", len(stats.ContainerState))
	}

	if stats.CheckCount != 10 {
		t.Errorf("expected CheckCount to be 10, got %d", stats.CheckCount)
	}
}
