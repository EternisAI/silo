package config

import (
	"os"
	"path/filepath"
)

const (
	ConfigFileName  = "config.yml"
	ComposeFileName = "docker-compose.yml"
	StateFileName   = "state.json"
)

type Paths struct {
	ConfigDir   string
	DataDir     string
	ConfigFile  string
	ComposeFile string
	StateFile   string
	AppDataDir  string
}

func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "silo", "config")
	}
	return filepath.Join(home, ".config", "silo")
}

func DefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "silo", "data")
	}
	return filepath.Join(home, ".local", "share", "silo")
}

func NewPaths(configDir string) *Paths {
	if configDir == "" {
		configDir = DefaultConfigDir()
	}

	dataDir := DefaultDataDir()

	return &Paths{
		ConfigDir:   configDir,
		DataDir:     dataDir,
		ConfigFile:  filepath.Join(configDir, ConfigFileName),
		ComposeFile: filepath.Join(dataDir, ComposeFileName),
		StateFile:   filepath.Join(dataDir, StateFileName),
		AppDataDir:  filepath.Join(dataDir, "data"),
	}
}
