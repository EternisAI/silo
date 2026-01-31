package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/pkg/logger"
)

// Monitor handles container health monitoring
type Monitor struct {
	paths  *config.Paths
	config *Config
	logger *logger.Logger
	stats  *MonitorStats
	mu     sync.RWMutex
}

// MonitorStats tracks monitoring statistics
type MonitorStats struct {
	LastCheck      time.Time
	CheckCount     int64
	RestartCount   int64
	FailedChecks   int64
	ContainerState map[string]string // service -> state
}

// NewMonitor creates a new container monitor
func NewMonitor(paths *config.Paths, config *Config, log *logger.Logger) *Monitor {
	return &Monitor{
		paths:  paths,
		config: config,
		logger: log,
		stats: &MonitorStats{
			ContainerState: make(map[string]string),
		},
	}
}

// Start begins monitoring containers
func (m *Monitor) Start(ctx context.Context) {
	m.logger.Info("Starting container monitor (interval: %s)", m.config.MonitorInterval)

	ticker := time.NewTicker(m.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Container monitor stopped")
			return
		case <-ticker.C:
			m.check()
		}
	}
}

// check performs a health check on all containers
func (m *Monitor) check() {
	ctx := context.Background()
	containers, err := docker.Ps(ctx, m.paths.ComposeFile)
	if err != nil {
		m.logger.Error("Failed to check container status: %v", err)
		m.recordFailedCheck()
		return
	}

	// Update stats with lock
	m.mu.Lock()
	m.stats.LastCheck = time.Now()
	m.stats.CheckCount++

	// Track state changes and containers needing restart
	var needsRestart []string
	for _, container := range containers {
		previousState := m.stats.ContainerState[container.Service]
		m.stats.ContainerState[container.Service] = container.State

		// Detect state changes
		if previousState != "" && previousState != container.State {
			m.logger.Info("Container %s changed state: %s -> %s", container.Service, previousState, container.State)
		}

		// Auto-restart exited containers if enabled
		if m.config.AutoRestart && container.State == "exited" {
			needsRestart = append(needsRestart, container.Service)
		}
	}
	m.mu.Unlock()

	// Restart containers outside of lock
	for _, service := range needsRestart {
		m.logger.Warn("Container %s is exited, attempting restart...", service)
		if err := m.restartContainer(service); err != nil {
			m.logger.Error("Failed to restart %s: %v", service, err)
		} else {
			m.incrementRestartCount()
			m.logger.Success("Successfully restarted %s", service)
		}
	}
}

// recordFailedCheck increments the failed check counter
func (m *Monitor) recordFailedCheck() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.FailedChecks++
}

// incrementRestartCount increments the restart counter
func (m *Monitor) incrementRestartCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats.RestartCount++
}

// restartContainer restarts a specific container
func (m *Monitor) restartContainer(service string) error {
	ctx := context.Background()
	return docker.Restart(ctx, m.paths.ComposeFile, service)
}

// GetStats returns current monitoring statistics
func (m *Monitor) GetStats() *MonitorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	stats := *m.stats
	stats.ContainerState = make(map[string]string)
	for k, v := range m.stats.ContainerState {
		stats.ContainerState[k] = v
	}

	return &stats
}
