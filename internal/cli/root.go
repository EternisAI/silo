package cli

import (
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	verbose       bool
	configDir     string
	imageRegistry string
	imageTag      string
	port          int
	log           *logger.Logger
)

var rootCmd = &cobra.Command{
	Use:   "silo",
	Short: "Silo - Application deployment and management CLI",
	Long: `Silo is a CLI tool to deploy, manage, and upgrade the Silo application
deployed via Docker Compose with PostgreSQL and pgvector.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log = logger.New(verbose)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", config.DefaultInstallDir(), "configuration directory")
}
