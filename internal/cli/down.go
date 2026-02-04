package cli

import (
	"context"
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/inference"
	"github.com/spf13/cobra"
)

var downAll bool

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop Silo containers",
	Long: `Stop Silo containers while preserving all data.

By default, the inference engine is NOT stopped (it's slow to restart).
Use --all to also stop the inference engine.

To completely remove Silo including data:
  silo down --all && rm -rf ~/.config/silo ~/.local/share/silo && docker volume prune`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir, "")

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

		// Stop inference engine if --all flag is set
		if downAll {
			cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
			if err != nil {
				log.Warn("Failed to load config: %v", err)
			} else {
				log.Info("Stopping inference engine...")
				engine := inference.New(cfg, log)
				if err := engine.Down(ctx); err != nil {
					log.Warn("Failed to stop inference engine: %v", err)
				}

				// Update state to track that inference was stopped
				state, _ := config.LoadState(paths.StateFile)
				if state == nil {
					state = &config.State{}
				}
				state.InferenceWasRunning = false
				if err := config.SaveState(paths.StateFile, state); err != nil {
					log.Warn("Failed to save state: %v", err)
				}
			}
		} else {
			log.Info("Inference engine not stopped (use --all to include)")
		}

		log.Info("Configuration preserved at %s", paths.ConfigDir)
		log.Info("Data preserved at %s", paths.DataDir)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
	downCmd.Flags().BoolVar(&downAll, "all", false, "Also stop the inference engine")
}
