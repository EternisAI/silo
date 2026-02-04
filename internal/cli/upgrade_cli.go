package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	versionpkg "github.com/eternisai/silo/internal/version"
	"github.com/spf13/cobra"
)

const installScriptURL = "https://raw.githubusercontent.com/EternisAI/silo/main/scripts/install.sh"

var upgradeCliCmd = &cobra.Command{
	Use:   "upgrade-cli",
	Short: "Upgrade the Silo CLI to the latest version",
	Long: `Upgrade the Silo CLI binary to the latest version.

This command will:
  - Check for the latest CLI version
  - Download and install the new binary
  - Restart the silod daemon service (Linux only)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info("Checking for updates...")

		versionInfo, err := versionpkg.Check(ctx, version)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !versionInfo.NeedsUpdate {
			log.Success("Already running the latest version (%s)", versionInfo.Current)
			return nil
		}

		log.Info("Upgrading CLI: %s → %s", versionInfo.Current, versionInfo.Latest)
		log.Info("Release notes: %s", versionInfo.UpdateURL)
		log.Info("")

		bashCmd := exec.Command("bash", "-c", fmt.Sprintf("curl -fsSL %s | bash", installScriptURL))
		bashCmd.Stdout = os.Stdout
		bashCmd.Stderr = os.Stderr

		if err := bashCmd.Run(); err != nil {
			return fmt.Errorf("failed to run install script: %w", err)
		}

		log.Success("CLI upgraded successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCliCmd)
}
