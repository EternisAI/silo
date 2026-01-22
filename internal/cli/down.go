package cli

import (
	"context"
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop Silo containers",
	Long: `Stop Silo containers while preserving all data.

To completely remove Silo including data:
  silo down && sudo rm -rf /opt/silo && docker volume prune`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
			log.Warn("Silo is not installed")
			log.Info("Run 'silo up' to install")
			return nil
		}

		log.Info("Stopping Silo containers...")
		if err := docker.Down(ctx, paths.ComposeFile, false); err != nil {
			log.Error("Failed to stop containers: %v", err)
			return err
		}

		log.Success("Silo stopped successfully")
		log.Info("Configuration and data preserved at %s", paths.InstallDir)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
