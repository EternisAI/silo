package updater

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
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
		return fmt.Errorf("silo is not installed. Run 'silo install' first")
	}

	if err := u.backupConfig(); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
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

	u.logger.Success("Silo updated successfully!")
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
	state.ImageTag = u.config.ImageTag

	if err := config.SaveState(u.paths.StateFile, state); err != nil {
		return err
	}

	return nil
}
