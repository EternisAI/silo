package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eternisai/silo/internal/config"
	versionpkg "github.com/eternisai/silo/internal/version"
	"github.com/eternisai/silo/pkg/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var jsonOutput bool

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

type VersionOutput struct {
	CLI    CLIInfo                 `json:"cli"`
	Latest *versionpkg.VersionInfo `json:"latest,omitempty"`
	Images []ImageOutput           `json:"images,omitempty"`
}

type CLIInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

type ImageOutput struct {
	Name        string `json:"name"`
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	NeedsUpdate bool   `json:"needs_update"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show CLI and application versions",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if jsonOutput {
			log = logger.NewSilent()
		}

		var output VersionOutput
		output.CLI = CLIInfo{
			Version:   version,
			Commit:    commit,
			BuildDate: buildDate,
		}

		versionInfo, err := versionpkg.Check(ctx, version)
		if err != nil {
			log.Debug("Failed to check for CLI updates: %v", err)
		} else {
			output.Latest = versionInfo
		}

		paths := config.NewPaths(configDir, "")
		cfg, err := config.Load(paths.ConfigFile)
		var imageTag string
		if err != nil {
			log.Warn("Failed to load config: %v", err)
		} else {
			imageTag = cfg.ImageTag
			imageVersions, err := versionpkg.CheckImageVersions(ctx, cfg.ImageTag)
			if err != nil {
				log.Warn("Failed to check Docker image versions: %v", err)
			} else {
				for _, img := range imageVersions {
					output.Images = append(output.Images, ImageOutput{
						Name:        img.ImageName,
						Current:     img.Current,
						Latest:      img.Latest,
						NeedsUpdate: img.NeedsUpdate,
					})
				}
			}
		}

		if jsonOutput {
			jsonData, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				log.Error("Failed to marshal JSON: %v", err)
				return
			}
			fmt.Println(string(jsonData))
		} else {
			fmt.Printf("Silo CLI\n")
			fmt.Printf("  Version:    %s\n", output.CLI.Version)
			fmt.Printf("  Commit:     %s\n", output.CLI.Commit)
			fmt.Printf("  Build Date: %s\n", output.CLI.BuildDate)
			fmt.Println()

			if output.Latest != nil {
				if output.Latest.NeedsUpdate {
					fmt.Printf("⚠ New CLI version available: %s (current: %s)\n", output.Latest.Latest, output.Latest.Current)
					fmt.Printf("Run 'silo upgrade-cli' to update\n")
					fmt.Printf("Release notes: %s\n", output.Latest.UpdateURL)
					fmt.Println()
				} else {
					green := color.New(color.FgGreen).SprintFunc()
					fmt.Printf("%s CLI is up to date\n", green("✓"))
					fmt.Println()
				}
			}

			if len(output.Images) > 0 {
				tag := imageTag
				if tag == "" {
					tag = "unknown"
				}
				fmt.Printf("Docker Images (configured tag: %s)\n", tag)
				anyUpdates := false
				yellow := color.New(color.FgYellow).SprintFunc()
				green := color.New(color.FgGreen).SprintFunc()
				for _, img := range output.Images {
					if img.NeedsUpdate {
						fmt.Printf("%s   %s: %s → %s (update available)\n", yellow("⚠"), img.Name, img.Current, img.Latest)
						anyUpdates = true
					} else {
						fmt.Printf("%s   %s: %s (up to date)\n", green("✓"), img.Name, img.Current)
					}
				}
				if anyUpdates {
					fmt.Println()
					fmt.Printf("Run 'silo upgrade' to update to the latest versions\n")
				}
			}
		}
	},
}

func init() {
	versionCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.AddCommand(versionCmd)
}
