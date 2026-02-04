package cli

import (
	"context"
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/inference"
	"github.com/eternisai/silo/internal/installer"
	"github.com/spf13/cobra"
)

var upAll bool

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start Silo containers",
	Long: `Start Silo containers. On first run, this will install Silo.

This command will:
  - First run: perform full installation
  - Subsequent runs: start existing containers

By default, the inference engine is NOT started. Use --all to include it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir, "")

		if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
			log.Info("First run detected, installing Silo...")

			cfg := config.NewDefaultConfig(paths)

			if imageTag != "" {
				cfg.ImageTag = imageTag
			}
			if port > 0 {
				cfg.Port = port
			}
			cfg.EnableInferenceEngine = enableInferenceEngine
			cfg.EnableProxyAgent = enableProxyAgent

			if err := config.Validate(cfg); err != nil {
				log.Error("Invalid configuration: %v", err)
				return err
			}

			inst := installer.New(cfg, paths, log)
			if err := inst.Install(ctx); err != nil {
				log.Error("Installation failed: %v", err)
				return err
			}

			// Start inference engine if --all flag is set
			if upAll {
				if err := startInferenceEngine(ctx, cfg, paths); err != nil {
					return err
				}
			}

			return nil
		}

		log.Info("Starting Silo containers...")

		if err := os.MkdirAll(paths.DataDir, 0755); err != nil {
			log.Error("Failed to create data directory: %v", err)
			return err
		}

		cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}

		// Save merged config back to ensure any new fields are persisted
		if err := config.Save(paths.ConfigFile, cfg); err != nil {
			log.Warn("Failed to save merged config: %v", err)
		}

		log.Info("Regenerating docker-compose.yml from current configuration...")
		if err := config.GenerateDockerCompose(cfg, paths.ComposeFile); err != nil {
			log.Error("Failed to generate docker-compose: %v", err)
			return err
		}

		if err := docker.Up(ctx, paths.ComposeFile); err != nil {
			log.Error("Failed to start containers: %v", err)
			return err
		}

		log.Success("Silo is running")

		// Start inference engine if --all flag is set
		if upAll {
			if err := startInferenceEngine(ctx, cfg, paths); err != nil {
				return err
			}
		}

		return nil
	},
}

func startInferenceEngine(ctx context.Context, cfg *config.Config, paths *config.Paths) error {
	log.Info("Starting inference engine...")
	engine := inference.New(cfg, log)
	if err := engine.Up(ctx); err != nil {
		log.Error("Failed to start inference engine: %v", err)
		return err
	}

	// Update state to track that inference was running
	state, _ := config.LoadState(paths.StateFile)
	if state == nil {
		state = &config.State{}
	}
	state.InferenceWasRunning = true
	if err := config.SaveState(paths.StateFile, state); err != nil {
		log.Warn("Failed to save state: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().StringVar(&imageTag, "image-tag", config.DefaultImageTag, "Docker image tag (first install only)")
	upCmd.Flags().IntVar(&port, "port", config.DefaultPort, "Application port (first install only)")
	upCmd.Flags().BoolVar(&enableInferenceEngine, "enable-inference-engine", config.DefaultEnableInferenceEngine, "Enable local inference engine (first install only)")
	upCmd.Flags().BoolVar(&enableProxyAgent, "enable-proxy-agent", config.DefaultEnableProxyAgent, "Enable proxy agent (first install only)")
	upCmd.Flags().BoolVar(&upAll, "all", false, "Include inference engine")
}
