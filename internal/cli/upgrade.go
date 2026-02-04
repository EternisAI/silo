package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/updater"
	versionpkg "github.com/eternisai/silo/internal/version"
	"github.com/eternisai/silo/pkg/logger"
	"github.com/spf13/cobra"
)

type UpgradeOutput struct {
	PreCheck PreCheckInfo `json:"pre_check"`
	Upgrade  UpgradeInfo  `json:"upgrade"`
	Error    *ErrorInfo   `json:"error,omitempty"`
	Success  bool         `json:"success"`
}

type PreCheckInfo struct {
	CLI    *CLICheckInfo    `json:"cli,omitempty"`
	Images []ImageCheckInfo `json:"images,omitempty"`
}

type CLICheckInfo struct {
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	NeedsUpdate bool   `json:"needs_update"`
	UpdateURL   string `json:"update_url,omitempty"`
}

type ImageCheckInfo struct {
	Name        string `json:"name"`
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	NeedsUpdate bool   `json:"needs_update"`
}

type UpgradeInfo struct {
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type ErrorInfo struct {
	Message string `json:"message"`
	Step    string `json:"step,omitempty"`
}

var upgradeJSONOutput bool

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
		paths := config.NewPaths(configDir, "")

		startTime := time.Now()
		var output UpgradeOutput
		output.Upgrade.StartedAt = startTime.Format(time.RFC3339)
		output.Success = false

		cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			log.Info("Run 'silo up' first")
			return err
		}

		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		versionInfo, err := versionpkg.Check(checkCtx, version)
		cancel()

		if err != nil {
			if !upgradeJSONOutput {
				log.Warn("Could not check for latest CLI version: %v", err)
			}
		} else {
			output.PreCheck.CLI = &CLICheckInfo{
				Current:     versionInfo.Current,
				Latest:      versionInfo.Latest,
				NeedsUpdate: versionInfo.NeedsUpdate,
				UpdateURL:   versionInfo.UpdateURL,
			}
			if !upgradeJSONOutput {
				if versionInfo.NeedsUpdate {
					log.Info("CLI: Upgrading from %s to %s", versionInfo.Current, versionInfo.Latest)
					log.Info("Release notes: %s", versionInfo.UpdateURL)
				} else {
					log.Info("CLI: Already running latest version %s", versionInfo.Current)
				}
			}
		}

		checkCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
		imageVersions, err := versionpkg.CheckImageVersions(checkCtx, cfg.ImageTag)
		cancel()

		anyImageUpdates := false
		if err != nil {
			if !upgradeJSONOutput {
				log.Debug("Could not check image versions: %v", err)
			}
		} else {
			for _, img := range imageVersions {
				output.PreCheck.Images = append(output.PreCheck.Images, ImageCheckInfo{
					Name:        img.ImageName,
					Current:     img.Current,
					Latest:      img.Latest,
					NeedsUpdate: img.NeedsUpdate,
				})
				if img.NeedsUpdate {
					anyImageUpdates = true
				}
			}
			if !upgradeJSONOutput {
				log.Info("Docker Images (current tag: %s):", cfg.ImageTag)
				for _, img := range imageVersions {
					if img.NeedsUpdate {
						log.Info("  %s: %s → %s available", img.ImageName, img.Current, img.Latest)
					} else {
						log.Info("  %s: %s (up to date)", img.ImageName, img.Current)
					}
				}
				if anyImageUpdates {
					log.Info("")
					log.Info("This upgrade will automatically update to the latest versions")
				}
			}
		}

		// Check if deep research image needs update
		deepResearchNeedsUpdate := cfg.DeepResearchImage != config.DefaultDeepResearchImage

		if !anyImageUpdates && !deepResearchNeedsUpdate {
			if !upgradeJSONOutput {
				log.Info("")
				log.Success("All services are already up to date. No upgrade needed.")
			}
			output.Upgrade.CompletedAt = time.Now().Format(time.RFC3339)
			output.Success = true

			if upgradeJSONOutput {
				jsonData, jsonErr := json.MarshalIndent(output, "", "  ")
				if jsonErr != nil {
					log.Error("Failed to marshal JSON: %v", jsonErr)
					return jsonErr
				}
				fmt.Println(string(jsonData))
			}
			return nil
		}

		if upgradeJSONOutput {
			log = logger.NewSilent()
		}

		upd := updater.New(cfg, paths, log)
		err = upd.Update(ctx)

		output.Upgrade.CompletedAt = time.Now().Format(time.RFC3339)

		if err != nil {
			output.Error = &ErrorInfo{
				Message: err.Error(),
			}
			output.Success = false
		} else {
			output.Success = true
		}

		if upgradeJSONOutput {
			jsonData, jsonErr := json.MarshalIndent(output, "", "  ")
			if jsonErr != nil {
				log.Error("Failed to marshal JSON: %v", jsonErr)
				return jsonErr
			}
			fmt.Println(string(jsonData))
		} else if err != nil {
			log.Error("Upgrade failed: %v", err)
		}

		return err
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeJSONOutput, "json", false, "Output in JSON format")
	rootCmd.AddCommand(upgradeCmd)
}
