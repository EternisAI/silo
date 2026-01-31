package cli

import (
	"os"

	"github.com/eternisai/silo/internal/config"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate configuration",
	Long: `Check if the existing configuration and installation files are valid.

This command will:
  - Verify the config file exists and can be parsed
  - Validate all configuration values
  - Check if installation files exist (docker-compose.yml, state.json)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.NewPaths(configDir, "")

		// Check config file exists
		if _, err := os.Stat(paths.ConfigFile); os.IsNotExist(err) {
			log.Error("Config file not found: %s", paths.ConfigFile)
			log.Info("Run 'silo up' to create a new configuration")
			return err
		}
		log.Success("Config file exists: %s", paths.ConfigFile)

		// Load and parse config
		cfg, err := config.Load(paths.ConfigFile)
		if err != nil {
			log.Error("Failed to load config: %v", err)
			return err
		}
		log.Success("Config file parsed successfully")

		// Check for unknown fields
		unknownFields, err := config.FindUnknownFields(paths.ConfigFile)
		if err != nil {
			log.Error("Failed to check for unknown fields: %v", err)
			return err
		}
		if len(unknownFields) > 0 {
			for _, field := range unknownFields {
				log.Warn("Unknown config field: %s", field)
			}
		}

		// Check for missing fields (will be filled with defaults)
		missingFields, err := config.FindMissingFields(paths.ConfigFile)
		if err != nil {
			log.Error("Failed to check for missing fields: %v", err)
			return err
		}
		if len(missingFields) > 0 {
			log.Info("Missing fields (will use defaults):")
			for _, field := range missingFields {
				log.Info("  - %s", field)
			}
		}

		// Validate config values
		if err := config.Validate(cfg); err != nil {
			log.Error("Invalid configuration: %v", err)
			return err
		}
		log.Success("Configuration values are valid")

		// Check generated files exist (indicates Silo was installed)
		composeExists := true
		if _, err := os.Stat(paths.ComposeFile); os.IsNotExist(err) {
			composeExists = false
		}

		stateExists := true
		if _, err := os.Stat(paths.StateFile); os.IsNotExist(err) {
			stateExists = false
		}

		if composeExists {
			log.Success("Docker Compose file exists: %s", paths.ComposeFile)
		} else {
			log.Warn("Docker Compose file not found: %s (run 'silo up' to install)", paths.ComposeFile)
		}

		if stateExists {
			log.Success("State file exists: %s", paths.StateFile)
		} else {
			log.Warn("State file not found: %s (run 'silo up' to install)", paths.StateFile)
		}

		log.Success("Configuration check passed")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
