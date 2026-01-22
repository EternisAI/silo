package cli

import (
	"context"
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/installer"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start Silo containers",
	Long: `Start Silo containers. On first run, this will install Silo.

This command will:
  - First run: perform full installation
  - Subsequent runs: start existing containers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
			log.Info("First run detected, installing Silo...")

			cfg := config.NewDefaultConfig(paths)

			if imageRegistry != "" {
				cfg.ImageRegistry = imageRegistry
			}
			if imageTag != "" {
				cfg.ImageTag = imageTag
			}
			if port > 0 {
				cfg.Port = port
			}

			if err := config.Validate(cfg); err != nil {
				log.Error("Invalid configuration: %v", err)
				return err
			}

			inst := installer.New(cfg, paths, log)
			if err := inst.Install(ctx); err != nil {
				log.Error("Installation failed: %v", err)
				return err
			}

			return nil
		}

		log.Info("Starting Silo containers...")

		if err := os.MkdirAll(paths.DataDir, 0755); err != nil {
			log.Error("Failed to create data directory: %v", err)
			return err
		}

		if err := docker.Up(ctx, paths.ComposeFile); err != nil {
			log.Error("Failed to start containers: %v", err)
			return err
		}

		log.Success("Silo is running")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().StringVar(&imageRegistry, "image-registry", config.DefaultImageRegistry, "Docker image registry (first install only)")
	upCmd.Flags().StringVar(&imageTag, "image-tag", config.DefaultImageTag, "Docker image tag (first install only)")
	upCmd.Flags().IntVar(&port, "port", config.DefaultPort, "Application port (first install only)")
}
