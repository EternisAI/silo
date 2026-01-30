package daemon

import (
	"context"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/version"
	"github.com/eternisai/silo/pkg/logger"
)

// Scheduler handles scheduled background tasks
type Scheduler struct {
	paths  *config.Paths
	config *config.Config
	daemon *Config
	logger *logger.Logger
}

// NewScheduler creates a new task scheduler
func NewScheduler(paths *config.Paths, cfg *config.Config, daemonCfg *Config, log *logger.Logger) *Scheduler {
	return &Scheduler{
		paths:  paths,
		config: cfg,
		daemon: daemonCfg,
		logger: log,
	}
}

// Start begins running scheduled tasks
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("Starting task scheduler")

	// Version check ticker (daily)
	versionTicker := time.NewTicker(24 * time.Hour)
	defer versionTicker.Stop()

	// Health check ticker (every 5 minutes)
	healthTicker := time.NewTicker(5 * time.Minute)
	defer healthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Task scheduler stopped")
			return
		case <-versionTicker.C:
			s.checkVersions()
		case <-healthTicker.C:
			s.performHealthCheck()
		}
	}
}

// checkVersions checks for available updates
func (s *Scheduler) checkVersions() {
	s.logger.Debug("Running scheduled version check...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cliVer, err := version.Check(ctx, s.config.Version)
	if err != nil {
		s.logger.Warn("Failed to check CLI version: %v", err)
	} else if cliVer.NeedsUpdate {
		s.logger.Info("CLI update available: %s -> %s", cliVer.Current, cliVer.Latest)
	}

	imageVers, err := version.CheckImageVersions(ctx, s.config.ImageTag)
	if err != nil {
		s.logger.Warn("Failed to check image versions: %v", err)
		return
	}

	// Log if updates are available
	for _, img := range imageVers {
		if img.NeedsUpdate {
			s.logger.Info("%s update available: %s -> %s", img.ImageName, img.Current, img.Latest)
		}
	}
}

// performHealthCheck performs a general health check
func (s *Scheduler) performHealthCheck() {
	s.logger.Debug("Running scheduled health check...")

	ctx := context.Background()
	containers, err := docker.Ps(ctx, s.paths.ComposeFile)
	if err != nil {
		s.logger.Error("Health check failed: %v", err)
		return
	}

	running := 0
	for _, c := range containers {
		if c.State == "running" {
			running++
		}
	}

	s.logger.Debug("Health check: %d/%d containers running", running, len(containers))
}
