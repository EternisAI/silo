package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DefaultVersion       = "0.1.0"
	DefaultImageRegistry = "ghcr.io/eternisai"
	DefaultImageTag      = "latest"
	DefaultPort          = 3000
)

type Config struct {
	Version       string `yaml:"version"`
	ImageRegistry string `yaml:"image_registry"`
	ImageTag      string `yaml:"image_tag"`
	Port          int    `yaml:"port"`
	ConfigFile    string `yaml:"-"`
	DataDir       string `yaml:"-"`
}

type State struct {
	Version       string `json:"version"`
	InstalledAt   string `json:"installed_at"`
	LastUpdated   string `json:"last_updated"`
	ImageRegistry string `json:"image_registry"`
	ImageTag      string `json:"image_tag"`
}

func NewDefaultConfig(paths *Paths) *Config {
	return &Config{
		Version:       DefaultVersion,
		ImageRegistry: DefaultImageRegistry,
		ImageTag:      DefaultImageTag,
		Port:          DefaultPort,
		ConfigFile:    paths.ConfigFile,
		DataDir:       paths.DataDir,
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func Save(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func Validate(config *Config) error {
	if config.ImageRegistry == "" {
		return fmt.Errorf("image_registry cannot be empty")
	}
	if config.ImageTag == "" {
		return fmt.Errorf("image_tag cannot be empty")
	}
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

func SaveState(path string, state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}
