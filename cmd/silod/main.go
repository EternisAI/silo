package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eternisai/silo/internal/daemon"
	"github.com/eternisai/silo/pkg/logger"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	log := logger.New(false)

	log.Info("Starting Silo Daemon v%s (commit: %s, built: %s)", version, commit, buildDate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create and start daemon
	d, err := daemon.New()
	if err != nil {
		log.Error("Failed to create daemon: %v", err)
		os.Exit(1)
	}

	// Start daemon in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := d.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Info("Received shutdown signal, stopping daemon...")
		cancel()
	case err := <-errChan:
		log.Error("Daemon error: %v", err)
		cancel()
		os.Exit(1)
	}

	// Graceful shutdown
	if err := d.Stop(); err != nil {
		log.Error("Error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Success("Silo Daemon stopped")
}
