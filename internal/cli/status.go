package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show deployment status",
	Long:  "Display the status of Silo containers, versions, and health information.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
			log.Warn("Silo is not installed")
			log.Info("Run 'silo up' to install")
			return nil
		}

		cfg, err := config.Load(paths.ConfigFile)
		if err != nil {
			log.Debug("Could not load config file: %v", err)
		}

		state, err := config.LoadState(paths.StateFile)
		if err != nil {
			log.Debug("Could not load state file: %v", err)
		}

		if cfg != nil || state != nil {
			log.Info("Installation Details:")
			if cfg != nil {
				log.Info("  Image Tag: %s", cfg.ImageTag)
			}
			if state != nil {
				log.Info("  CLI Version: %s", state.Version)
				log.Info("  Installed: %s", state.InstalledAt)
				log.Info("  Last Updated: %s", state.LastUpdated)
			}
			fmt.Println()
		}

		containers, err := docker.Ps(ctx, paths.ComposeFile)
		if err != nil {
			log.Error("Failed to get container status: %v", err)
			return err
		}

		if len(containers) == 0 {
			log.Warn("No containers running")
			log.Info("Run 'silo up' to start")
			return nil
		}

		log.Info("Containers:")
		for _, c := range containers {
			status := "✓"
			if c.State != "running" {
				status = "✗"
			}

			log.Info("  %s %s (%s)", status, c.Service, c.State)
			log.Info("    Name:   %s", c.Name)
			log.Info("    Image:  %s", c.Image)
			log.Info("    Status: %s", c.Status)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
