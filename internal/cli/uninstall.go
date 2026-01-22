package cli

import (
	"context"
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/spf13/cobra"
)

var purgeData bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Silo application",
	Long: `Stop and remove Silo containers.

Use --purge-data to also remove all data volumes and configuration files.
WARNING: This will permanently delete all data!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
			log.Warn("Silo is not installed")
			return nil
		}

		log.Info("Uninstalling Silo...")

		log.Info("Stopping containers...")
		if err := docker.Down(ctx, paths.ComposeFile, purgeData); err != nil {
			log.Error("Failed to stop containers: %v", err)
			return err
		}
		log.Success("Containers stopped")

		if purgeData {
			log.Info("Removing installation directory...")
			if err := os.RemoveAll(paths.InstallDir); err != nil {
				log.Error("Failed to remove installation directory: %v", err)
				return err
			}
			log.Success("Installation directory removed")
		}

		log.Success("Silo uninstalled successfully")

		if !purgeData {
			log.Info("Configuration and data preserved at %s", paths.InstallDir)
			log.Info("Use 'silo uninstall --purge-data' to remove all data")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVar(&purgeData, "purge-data", false, "remove all data volumes and configuration files")
}
