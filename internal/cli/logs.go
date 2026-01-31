package cli

import (
	"os"
	"strconv"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View Silo container logs",
	Long:  `View logs from Silo containers. Use -f to follow logs in real-time.`,
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVar(&logsTail, "tail", "100", "Number of lines to show from the end of the logs")
	rootCmd.AddCommand(logsCmd)
}

func runLogs(cmd *cobra.Command, args []string) error {
	paths := config.NewPaths(configDir, "")

	if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
		log.Error("Silo is not installed")
		log.Info("Run 'silo up' to install")
		return err
	}

	tail := 100
	if logsTail != "" {
		if n, err := strconv.Atoi(logsTail); err == nil {
			tail = n
		}
	}

	opts := docker.LogOptions{
		Follow: logsFollow,
		Lines:  tail,
	}

	service := ""
	if len(args) > 0 {
		service = args[0]
	}

	return docker.Logs(cmd.Context(), paths.ComposeFile, service, opts)
}
