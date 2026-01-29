package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/eternisai/silo/internal/config"
	versionpkg "github.com/eternisai/silo/internal/version"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show CLI and application versions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Silo CLI\n")
		fmt.Printf("  Version:    %s\n", version)
		fmt.Printf("  Commit:     %s\n", commit)
		fmt.Printf("  Build Date: %s\n", buildDate)
		fmt.Println()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		versionInfo, err := versionpkg.Check(ctx, version)
		if err != nil {
			log.Debug("Failed to check for CLI updates: %v", err)
		} else {
			if versionInfo.NeedsUpdate {
				log.Warn("New CLI version available: %s (current: %s)", versionInfo.Latest, versionInfo.Current)
				log.Info("Run 'silo upgrade' to update")
				log.Info("Release notes: %s", versionInfo.UpdateURL)
				fmt.Println()
			} else {
				log.Success("CLI is up to date")
				fmt.Println()
			}
		}

		paths := config.NewPaths(configDir)
		cfg, err := config.Load(paths.ConfigFile)
		if err != nil {
			log.Debug("Could not load config file, skipping image version check: %v", err)
			return
		}

		log.Info("Docker Images (configured tag: %s)", cfg.ImageTag)
		imageVersions, err := versionpkg.CheckImageVersions(ctx, cfg.ImageTag)
		if err != nil {
			log.Debug("Failed to check image versions: %v", err)
			return
		}

		anyUpdates := false
		for _, img := range imageVersions {
			if img.NeedsUpdate {
				log.Warn("  %s: %s → %s (update available)", img.ImageName, img.Current, img.Latest)
				anyUpdates = true
			} else {
				log.Success("  %s: %s (up to date)", img.ImageName, img.Current)
			}
		}

		if anyUpdates {
			fmt.Println()
			log.Info("To update images, edit ~/.config/silo/config.yml and run 'silo upgrade'")
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
