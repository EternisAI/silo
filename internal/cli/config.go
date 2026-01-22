package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/eternisai/silo/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Silo configuration",
	Long:  "View and edit Silo configuration files.",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ConfigFile); os.IsNotExist(err) {
			log.Error("Configuration file not found. Is Silo installed?")
			return nil
		}

		data, err := os.ReadFile(paths.ConfigFile)
		if err != nil {
			log.Error("Failed to read config file: %v", err)
			return err
		}

		fmt.Print(string(data))
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  "Open the configuration file in the default editor ($EDITOR or vi).",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ConfigFile); os.IsNotExist(err) {
			log.Error("Configuration file not found. Is Silo installed?")
			return nil
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editorCmd := exec.Command(editor, paths.ConfigFile)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			log.Error("Failed to edit config: %v", err)
			return err
		}

		cfg, err := config.Load(paths.ConfigFile)
		if err != nil {
			log.Error("Failed to parse edited config: %v", err)
			log.Warn("Configuration may be invalid")
			return err
		}

		if err := config.Validate(cfg); err != nil {
			log.Error("Invalid configuration: %v", err)
			log.Warn("Please fix the configuration file")
			return err
		}

		log.Success("Configuration updated")
		log.Info("Restart containers with 'silo update' to apply changes")
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.NewPaths(configDir)

		if _, err := os.Stat(paths.ConfigFile); os.IsNotExist(err) {
			log.Error("Configuration file not found. Is Silo installed?")
			return nil
		}

		data, err := os.ReadFile(paths.ConfigFile)
		if err != nil {
			log.Error("Failed to read config file: %v", err)
			return err
		}

		var configData map[string]interface{}
		if err := yaml.Unmarshal(data, &configData); err != nil {
			log.Error("Failed to parse config file: %v", err)
			return err
		}

		key := args[0]
		value := args[1]
		configData[key] = value

		newData, err := yaml.Marshal(configData)
		if err != nil {
			log.Error("Failed to marshal config: %v", err)
			return err
		}

		if err := os.WriteFile(paths.ConfigFile, newData, 0644); err != nil {
			log.Error("Failed to write config file: %v", err)
			return err
		}

		log.Success("Configuration updated: %s = %s", key, value)
		log.Info("Restart containers with 'silo update' to apply changes")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configSetCmd)
}
