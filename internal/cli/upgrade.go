package cli

import (
	"context"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/updater"
	versionpkg "github.com/eternisai/silo/internal/version"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Silo to the latest version",
	Long: `Upgrade Silo to the latest version by pulling new images and recreating containers.

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
			log.Info("Run 'silo up' first")
			return err
		}

		cfg.ConfigFile = paths.ConfigFile
		cfg.DataDir = paths.AppDataDir

		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		versionInfo, err := versionpkg.Check(checkCtx, version)
		cancel()

		if err != nil {
			log.Warn("Could not check for latest version: %v", err)
			log.Info("Proceeding with upgrade anyway...")
		} else if versionInfo.NeedsUpdate {
			log.Info("Upgrading from %s to %s", versionInfo.Current, versionInfo.Latest)
			log.Info("Release notes: %s", versionInfo.UpdateURL)
		} else {
			log.Info("Already running latest version %s", versionInfo.Current)
		}

		upd := updater.New(cfg, paths, log)
		if err := upd.Update(ctx); err != nil {
			log.Error("Upgrade failed: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
