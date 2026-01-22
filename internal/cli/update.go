package cli

import (
	"context"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Silo to the latest version",
	Long: `Update Silo to the latest version by pulling new images and recreating containers.

This command will:
  - Backup current configuration
  - Pull latest Docker images
  - Recreate containers with new images
  - Preserve all data`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir)

		cfg, err := config.Load(paths.ConfigFile)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}

		cfg.ConfigFile = paths.ConfigFile
		cfg.DataDir = paths.DataDir

		upd := updater.New(cfg, paths, log)
		if err := upd.Update(ctx); err != nil {
			log.Error("Update failed: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
