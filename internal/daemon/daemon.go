package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/version"
	"github.com/eternisai/silo/pkg/logger"
)

// Daemon represents the background service
type Daemon struct {
	config *config.Config
	state  *config.State
	paths  *config.Paths
	server *Server
	logger *logger.Logger
	opLock sync.Mutex // Prevents concurrent operations
	wg     sync.WaitGroup
}

// Config holds daemon configuration
type Config struct {
	ServerEnabled     bool
	ServerPort        int
	ServerBindAddress string
	LogFile           string
}

// DefaultConfig returns default daemon configuration
func DefaultConfig() *Config {
	return &Config{
		ServerEnabled:     true,
		ServerPort:        9999,
		ServerBindAddress: "0.0.0.0", // Allow access from host and Docker containers
		LogFile:           "",
	}
}

// New creates a new daemon instance
func New() (*Daemon, error) {
	log := logger.New(false)

	// Determine config directory
	configDir := os.Getenv("SILO_CONFIG_DIR")
	if configDir == "" {
		configDir = config.DefaultConfigDir()
	}

	// Determine data directory
	dataDir := os.Getenv("SILO_DATA_DIR")
	if dataDir == "" {
		dataDir = config.DefaultDataDir()
	}

	paths := config.NewPaths(configDir, dataDir)

	// Load configuration (create default if not found)
	cfg, err := config.Load(paths.ConfigFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("Config file not found, creating default config...")
			cfg = config.NewDefaultConfig(paths)

			// Create config directory if it doesn't exist
			if err := os.MkdirAll(paths.ConfigDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create config directory: %w", err)
			}

			// Save default config
			if err := config.GenerateConfig(cfg, paths.ConfigFile); err != nil {
				return nil, fmt.Errorf("failed to generate config file: %w", err)
			}

			log.Success("Created default config at %s", paths.ConfigFile)
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Load state (use default empty state if not found)
	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("State file not found, using default empty state")
			state = &config.State{}
		} else {
			return nil, fmt.Errorf("failed to load state: %w", err)
		}
	}

	// Create daemon components
	daemonCfg := DefaultConfig()

	d := &Daemon{
		config: cfg,
		state:  state,
		paths:  paths,
		logger: log,
	}

	// Create API server if enabled
	if daemonCfg.ServerEnabled {
		// Allow override via environment variable
		bindAddr := os.Getenv("SILO_DAEMON_BIND_ADDRESS")
		if bindAddr == "" {
			bindAddr = daemonCfg.ServerBindAddress
		}
		d.server = NewServer(bindAddr, daemonCfg.ServerPort, d, log)
	}

	return d, nil
}

// Start begins daemon operations
func (d *Daemon) Start(ctx context.Context) error {
	d.logger.Info("Starting daemon services...")

	// Error channel for critical failures
	errChan := make(chan error, 1)

	// Start API server if enabled
	if d.server != nil {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			if err := d.server.Start(ctx); err != nil {
				d.logger.Error("Server error: %v", err)
				select {
				case errChan <- err:
				default:
				}
			}
		}()
	}

	d.logger.Success("Daemon started successfully")

	// Wait for context cancellation or critical error
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

// Stop gracefully stops the daemon
func (d *Daemon) Stop() error {
	d.logger.Info("Stopping daemon services...")

	// Stop server first
	if d.server != nil {
		if err := d.server.Stop(); err != nil {
			d.logger.Warn("Error stopping server: %v", err)
		}
	}

	// Wait for all goroutines to finish
	d.wg.Wait()

	d.logger.Success("Daemon stopped successfully")
	return nil
}

// GetStatus returns current daemon status
func (d *Daemon) GetStatus() (*Status, error) {
	ctx := context.Background()
	containers, err := docker.Ps(ctx, d.paths.ComposeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	// Get version info
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cliVer, err := version.Check(ctx, d.config.Version)
	if err != nil {
		d.logger.Warn("Failed to check CLI version: %v", err)
	}

	imageVers, err := version.CheckImageVersions(ctx, d.config.ImageTag)
	if err != nil {
		d.logger.Warn("Failed to check image versions: %v", err)
	}

	return &Status{
		State:         d.state,
		Config:        d.config,
		Containers:    containers,
		CLIVersion:    cliVer,
		ImageVersions: imageVers,
	}, nil
}

// Status represents daemon status
type Status struct {
	State         *config.State
	Config        *config.Config
	Containers    []docker.Container
	CLIVersion    *version.VersionInfo
	ImageVersions []version.ImageVersionInfo
}
