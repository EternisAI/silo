package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/pkg/logger"
)

type Installer struct {
	config *config.Config
	paths  *config.Paths
	logger *logger.Logger
}

func New(cfg *config.Config, paths *config.Paths, log *logger.Logger) *Installer {
	return &Installer{
		config: cfg,
		paths:  paths,
		logger: log,
	}
}

func (i *Installer) Install(ctx context.Context) error {
	i.logger.Info("Starting Silo installation...")

	if err := i.runPreflightChecks(); err != nil {
		return fmt.Errorf("preflight checks failed: %w", err)
	}

	if err := i.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if err := i.generateConfigs(); err != nil {
		return fmt.Errorf("failed to generate configs: %w", err)
	}

	if err := i.pullImages(ctx); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	if err := i.startContainers(ctx); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}

	if err := i.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	i.logger.Success("Silo installed successfully!")
	i.logger.Info("Application is running at http://localhost:%d", i.config.Port)

	return nil
}

func (i *Installer) runPreflightChecks() error {
	i.logger.Info("Running preflight checks...")

	i.logger.Debug("Checking system requirements...")
	if err := CheckSystemRequirements(); err != nil {
		return err
	}

	parentDir := filepath.Dir(i.paths.DataDir)
	i.logger.Debug("Checking disk space for %s...", parentDir)
	if err := CheckDiskSpace(parentDir, RequiredDiskSpaceGB); err != nil {
		return err
	}

	i.logger.Debug("Checking port availability...")
	if err := CheckPortAvailability(i.config.Port); err != nil {
		return err
	}

	i.logger.Success("Preflight checks passed")
	return nil
}

func (i *Installer) createDirectories() error {
	i.logger.Info("Creating directories...")

	dirs := []string{
		i.paths.ConfigDir,
		i.paths.DataDir,
		i.paths.AppDataDir,
	}

	for _, dir := range dirs {
		i.logger.Debug("Creating directory: %s", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	i.logger.Success("Directories created")
	return nil
}

func (i *Installer) generateConfigs() error {
	i.logger.Info("Generating configuration files...")

	i.logger.Debug("Generating docker-compose.yml...")
	if err := config.GenerateDockerCompose(i.config, i.paths.ComposeFile); err != nil {
		return err
	}

	i.logger.Debug("Generating config.yml...")
	if err := config.GenerateConfig(i.config, i.paths.ConfigFile); err != nil {
		return err
	}

	i.logger.Success("Configuration files generated")
	return nil
}

func (i *Installer) pullImages(ctx context.Context) error {
	i.logger.Info("Pulling Docker images...")

	// Determine which services to pull
	services := []string{"backend", "frontend"}
	if i.config.EnableDeepResearch {
		services = append(services, "deep-research")
	}

	// Pull each service, tracking failures
	results := docker.Pull(ctx, i.paths.ComposeFile, services...)

	var failed []string
	for _, r := range results {
		if r.Error != nil {
			i.logger.Warn("Failed to pull %s: %v", r.Service, r.Error)
			failed = append(failed, r.Service)
		} else {
			i.logger.Success("Pulled %s", r.Service)
		}
	}

	// Fail only if critical services (backend/frontend) failed
	for _, f := range failed {
		if f == "backend" || f == "frontend" {
			return fmt.Errorf("failed to pull critical service: %s", f)
		}
	}

	if len(failed) > 0 {
		i.logger.Warn("Some non-critical images failed to pull, continuing anyway")
	}

	return nil
}

func (i *Installer) startContainers(ctx context.Context) error {
	i.logger.Info("Starting containers...")

	if err := docker.Up(ctx, i.paths.ComposeFile); err != nil {
		return err
	}

	i.logger.Success("Containers started")
	return nil
}

func (i *Installer) saveState() error {
	i.logger.Debug("Saving installation state...")

	state := &config.State{
		Version:     i.config.Version,
		InstalledAt: time.Now().Format(time.RFC3339),
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	if err := config.SaveState(i.paths.StateFile, state); err != nil {
		return err
	}

	return nil
}
