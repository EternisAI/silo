package cli

import (
	"context"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/installer"
	"github.com/spf13/cobra"
)

var (
	imageRegistry string
	imageTag      string
	port          int
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and deploy the Silo application",
	Long: `Install and deploy the Silo application with Docker Compose.

This command will:
  - Check Docker dependencies
  - Create installation directories
  - Generate configuration files
  - Pull Docker images
  - Start containers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		paths := config.NewPaths(configDir)
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

		running, err := docker.IsRunning(ctx, paths.ComposeFile)
		if err == nil && running {
			log.Warn("Silo is already installed and running")
			log.Info("Use 'silo update' to update to a new version")
			return nil
		}

		inst := installer.New(cfg, paths, log)
		if err := inst.Install(ctx); err != nil {
			log.Error("Installation failed: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVar(&imageRegistry, "image-registry", config.DefaultImageRegistry, "Docker image registry")
	installCmd.Flags().StringVar(&imageTag, "image-tag", config.DefaultImageTag, "Docker image tag")
	installCmd.Flags().IntVar(&port, "port", config.DefaultPort, "Application port")
}
