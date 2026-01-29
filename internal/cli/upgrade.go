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
			log.Warn("Could not check for latest CLI version: %v", err)
		} else if versionInfo.NeedsUpdate {
			log.Info("CLI: Upgrading from %s to %s", versionInfo.Current, versionInfo.Latest)
			log.Info("Release notes: %s", versionInfo.UpdateURL)
		} else {
			log.Info("CLI: Already running latest version %s", versionInfo.Current)
		}

		checkCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
		imageVersions, err := versionpkg.CheckImageVersions(checkCtx, cfg.ImageTag)
		cancel()

		if err != nil {
			log.Debug("Could not check image versions: %v", err)
		} else {
			log.Info("Docker Images (current tag: %s):", cfg.ImageTag)
			for _, img := range imageVersions {
				if img.NeedsUpdate {
					log.Info("  %s: %s → %s available", img.ImageName, img.Current, img.Latest)
				} else {
					log.Info("  %s: %s (up to date)", img.ImageName, img.Current)
				}
			}
			log.Info("")
			log.Info("Note: This upgrade pulls images with tag '%s'", cfg.ImageTag)
			log.Info("To use a different version, edit image_tag in config.yml first")
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
