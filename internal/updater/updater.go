package updater

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/version"
	"github.com/eternisai/silo/pkg/logger"
)

type Updater struct {
	config *config.Config
	paths  *config.Paths
	logger *logger.Logger
}

func New(cfg *config.Config, paths *config.Paths, log *logger.Logger) *Updater {
	return &Updater{
		config: cfg,
		paths:  paths,
		logger: log,
	}
}

func (u *Updater) Update(ctx context.Context) error {
	u.logger.Info("Starting Silo update...")

	if _, err := os.Stat(u.paths.ComposeFile); os.IsNotExist(err) {
		return fmt.Errorf("silo is not installed, run 'silo up' first")
	}

	if err := u.backupConfig(); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	newTag, err := u.updateConfigWithLatestVersion(checkCtx)
	cancel()

	if err != nil {
		u.logger.Warn("Could not auto-update to latest version: %v", err)
		u.logger.Info("Continuing with current version %s", u.config.ImageTag)
	}

	// Update deep research image to latest default if outdated
	if updated, info := config.UpdateDeepResearchImage(u.config, u.paths.ConfigFile); updated {
		u.logger.Info("Deep research image %s", info)
		// Regenerate docker-compose with new image
		if err := config.GenerateDockerCompose(u.config, u.paths.ComposeFile); err != nil {
			u.logger.Warn("Failed to regenerate docker-compose for deep research: %v", err)
		}
	}

	if err := u.pullImages(ctx); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	if err := u.recreateContainers(ctx); err != nil {
		return fmt.Errorf("failed to recreate containers: %w", err)
	}

	if err := u.updateState(); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	if newTag != "" {
		u.logger.Success("Silo upgraded to version %s successfully!", newTag)
	} else {
		u.logger.Success("Silo updated successfully!")
	}
	return nil
}

func (u *Updater) backupConfig() error {
	u.logger.Info("Backing up configuration...")

	backupPath := u.paths.ConfigFile + ".backup"
	data, err := os.ReadFile(u.paths.ConfigFile)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return err
	}

	u.logger.Success("Configuration backed up to %s", backupPath)
	return nil
}

func (u *Updater) updateConfigWithLatestVersion(ctx context.Context) (string, error) {
	u.logger.Info("Checking for latest Docker image versions...")

	imageVersions, err := version.CheckImageVersions(ctx, u.config.ImageTag)
	if err != nil {
		return "", fmt.Errorf("failed to check image versions: %w", err)
	}

	if len(imageVersions) == 0 {
		return "", fmt.Errorf("no image version information available")
	}

	latestTag := imageVersions[0].Latest
	currentTag := u.config.ImageTag

	if !imageVersions[0].NeedsUpdate {
		u.logger.Info("Already running latest version %s", currentTag)
		return currentTag, nil
	}

	u.logger.Info("Updating image tag: %s → %s", currentTag, latestTag)

	if err := config.UpdateImageTag(u.config, latestTag, u.paths.ConfigFile); err != nil {
		return "", fmt.Errorf("failed to update config: %w", err)
	}
	u.logger.Success("Config updated with new image tag")

	u.logger.Info("Regenerating docker-compose.yml...")
	if err := config.GenerateDockerCompose(u.config, u.paths.ComposeFile); err != nil {
		return "", fmt.Errorf("failed to regenerate docker-compose: %w", err)
	}
	u.logger.Success("Docker-compose regenerated with new image versions")

	return latestTag, nil
}

func (u *Updater) pullImages(ctx context.Context) error {
	u.logger.Info("Pulling latest Docker images...")

	if err := docker.Pull(ctx, u.paths.ComposeFile); err != nil {
		return err
	}

	u.logger.Success("Docker images pulled")
	return nil
}

func (u *Updater) recreateContainers(ctx context.Context) error {
	u.logger.Info("Recreating containers...")

	if err := docker.Down(ctx, u.paths.ComposeFile, false); err != nil {
		return err
	}

	if err := docker.Up(ctx, u.paths.ComposeFile); err != nil {
		return err
	}

	u.logger.Success("Containers recreated")
	return nil
}

func (u *Updater) updateState() error {
	u.logger.Debug("Updating installation state...")

	state, err := config.LoadState(u.paths.StateFile)
	if err != nil {
		state = &config.State{}
	}

	state.LastUpdated = time.Now().Format(time.RFC3339)

	if err := config.SaveState(u.paths.StateFile, state); err != nil {
		return err
	}

	return nil
}
