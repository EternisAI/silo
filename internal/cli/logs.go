package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/spf13/cobra"
)

var (
	follow bool
	lines  int
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View container logs",
	Long: `View logs from Silo containers.

If no service is specified, logs from all containers are shown.
Available services: silo, pgvector`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
			log.Error("Silo is not installed")
			return nil
		}

		service := ""
		if len(args) > 0 {
			service = args[0]
		}

		opts := docker.LogOptions{
			Follow: follow,
			Lines:  lines,
		}

		if err := docker.Logs(ctx, paths.ComposeFile, service, opts); err != nil {
			log.Error("Failed to fetch logs: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	logsCmd.Flags().IntVarP(&lines, "tail", "n", 100, "number of lines to show from the end of the logs")
}
