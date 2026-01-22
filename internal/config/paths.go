package config

import (
	"path/filepath"
)

const (
	DefaultInstallDir = "/opt/silo"
	ConfigFileName    = "config.yml"
	ComposeFileName   = "docker-compose.yml"
	StateFileName     = "state.json"
)

type Paths struct {
	InstallDir  string
	ConfigFile  string
	ComposeFile string
	DataDir     string
	StateFile   string
}

func NewPaths(installDir string) *Paths {
	if installDir == "" {
		installDir = DefaultInstallDir
	}

	return &Paths{
		InstallDir:  installDir,
		ConfigFile:  filepath.Join(installDir, ConfigFileName),
		ComposeFile: filepath.Join(installDir, ComposeFileName),
		DataDir:     filepath.Join(installDir, "data"),
		StateFile:   filepath.Join(installDir, StateFileName),
	}
}
