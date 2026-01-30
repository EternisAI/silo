package daemon

import (
	"context"
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
	config    *config.Config
	state     *config.State
	paths     *config.Paths
	monitor   *Monitor
	scheduler *Scheduler
	server    *Server
	logger    *logger.Logger
	opLock    sync.Mutex // Prevents concurrent operations
	wg        sync.WaitGroup
}

// Config holds daemon configuration
type Config struct {
	MonitorInterval   time.Duration
	VersionCheckCron  string
	HealthCheckCron   string
	AutoRestart       bool
	ServerEnabled     bool
	ServerPort        int
	ServerBindAddress string
	LogFile           string
}

// DefaultConfig returns default daemon configuration
func DefaultConfig() *Config {
	return &Config{
		MonitorInterval:   30 * time.Second,
		VersionCheckCron:  "0 2 * * *",  // Daily at 2 AM
		HealthCheckCron:   "*/5 * * * *", // Every 5 minutes
		AutoRestart:       true,
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

	paths := config.NewPaths(configDir)

	// Load configuration
	cfg, err := config.Load(paths.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Load state
	state, err := config.LoadState(paths.StateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Check if system is installed
	if state.InstalledAt == "" {
		return nil, fmt.Errorf("silo is not installed, run 'silo up' first")
	}

	// Create daemon components
	daemonCfg := DefaultConfig()

	d := &Daemon{
		config:    cfg,
		state:     state,
		paths:     paths,
		logger:    log,
		monitor:   NewMonitor(paths, daemonCfg, log),
		scheduler: NewScheduler(paths, cfg, daemonCfg, log),
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

	// Start monitor
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.monitor.Start(ctx)
	}()

	// Start scheduler
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.scheduler.Start(ctx)
	}()

	// Start API server if enabled
	if d.server != nil {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			if err := d.server.Start(ctx); err != nil {
				d.logger.Error("Server error: %v", err)
			}
		}()
	}

	d.logger.Success("Daemon started successfully")

	// Wait for context cancellation
	<-ctx.Done()

	return nil
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
		MonitorStats:  d.monitor.GetStats(),
	}, nil
}

// Status represents daemon status
type Status struct {
	State         *config.State
	Config        *config.Config
	Containers    []docker.Container
	CLIVersion    *version.VersionInfo
	ImageVersions []version.ImageVersionInfo
	MonitorStats  *MonitorStats
}
