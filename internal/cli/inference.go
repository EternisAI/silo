package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/inference"
	"github.com/spf13/cobra"
)

var (
	inferenceFollow bool
	inferenceLines  int
)

var inferenceCmd = &cobra.Command{
	Use:   "inference",
	Short: "Manage the inference engine",
	Long: `Manage the SGLang inference engine container.

The inference engine runs separately from the main Silo services and is
not affected by 'silo up', 'silo down', or 'silo upgrade' commands.

Use 'silo up --all' or 'silo down --all' to include the inference engine.`,
}

var inferenceUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the inference engine",
	Long: `Start the SGLang inference engine container.

This will start the inference engine with the configuration from config.yml.
The container runs with --restart unless-stopped, so it will automatically
restart after system reboots.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir, "")

		cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}

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
	},
}

var inferenceDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the inference engine",
	Long:  `Stop and remove the SGLang inference engine container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir, "")

		cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}

		engine := inference.New(cfg, log)
		if err := engine.Down(ctx); err != nil {
			log.Error("Failed to stop inference engine: %v", err)
			return err
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

		return nil
	},
}

var inferenceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show inference engine status",
	Long:  `Display the current status of the inference engine container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		paths := config.NewPaths(configDir, "")

		cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}

		engine := inference.New(cfg, log)
		info, err := engine.Status(ctx)
		if err != nil {
			log.Error("Failed to get status: %v", err)
			return err
		}

		log.Info("Inference Engine:")
		if info.Running {
			log.Info("  ✓ %s (running)", info.Name)
			log.Info("    Image:  %s", info.Image)
			log.Info("    Status: %s", info.Status)

			// Try health check
			if err := engine.HealthCheck(ctx); err == nil {
				log.Success("  Health: OK")
			} else {
				log.Warn("  Health: Not responding (may still be starting)")
			}
		} else {
			log.Warn("  ✗ %s (%s)", info.Name, info.State)
		}

		return nil
	},
}

var inferenceLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View inference engine logs",
	Long:  `View logs from the inference engine container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle interrupt signal for log following
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		paths := config.NewPaths(configDir, "")

		cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}

		engine := inference.New(cfg, log)
		if err := engine.Logs(ctx, inferenceFollow, inferenceLines); err != nil {
			if ctx.Err() == context.Canceled {
				return nil
			}
			log.Error("Failed to get logs: %v", err)
			return err
		}

		return nil
	},
}

var inferenceShowConfigCmd = &cobra.Command{
	Use:   "show-config",
	Short: "Show the docker run command",
	Long:  `Display the full docker run command that would be used to start the inference engine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.NewPaths(configDir, "")

		cfg, err := config.LoadOrDefault(paths.ConfigFile, paths)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}

		engine := inference.New(cfg, log)
		log.Info("Docker run command:")
		log.Info(engine.GetDockerRunCommand())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(inferenceCmd)

	inferenceCmd.AddCommand(inferenceUpCmd)
	inferenceCmd.AddCommand(inferenceDownCmd)
	inferenceCmd.AddCommand(inferenceStatusCmd)
	inferenceCmd.AddCommand(inferenceLogsCmd)
	inferenceCmd.AddCommand(inferenceShowConfigCmd)

	inferenceLogsCmd.Flags().BoolVarP(&inferenceFollow, "follow", "f", false, "Follow log output")
	inferenceLogsCmd.Flags().IntVarP(&inferenceLines, "tail", "n", 100, "Number of lines to show")
}
