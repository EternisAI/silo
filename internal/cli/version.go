package cli

import (
	"context"
	"fmt"
	"time"

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
			log.Debug("Failed to check for updates: %v", err)
			return
		}

		if versionInfo.NeedsUpdate {
			log.Warn("New version available: %s (current: %s)", versionInfo.Latest, versionInfo.Current)
			log.Info("Run 'silo upgrade' to update")
			log.Info("Release notes: %s", versionInfo.UpdateURL)
		} else {
			log.Success("You are running the latest version")
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
