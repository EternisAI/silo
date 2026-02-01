package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/eternisai/silo/pkg/logger"
)

// Server provides HTTP API for daemon
type Server struct {
	bindAddr   string
	port       int
	socketPath string
	daemon     *Daemon
	logger     *logger.Logger
	server     *http.Server
}

// NewServer creates a new HTTP server
func NewServer(bindAddr string, port int, socketPath string, daemon *Daemon, log *logger.Logger) *Server {
	return &Server{
		bindAddr:   bindAddr,
		port:       port,
		socketPath: socketPath,
		daemon:     daemon,
		logger:     log,
	}
}

// Start begins serving HTTP requests
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Register existing handlers
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)

	// Register command API handlers
	mux.HandleFunc("/api/v1/up", s.handleUp)
	mux.HandleFunc("/api/v1/down", s.handleDown)
	mux.HandleFunc("/api/v1/restart", s.handleRestart)
	mux.HandleFunc("/api/v1/upgrade", s.handleUpgrade)
	mux.HandleFunc("/api/v1/logs", s.handleLogs)
	mux.HandleFunc("/api/v1/version", s.handleVersion)
	mux.HandleFunc("/api/v1/check", s.handleCheck)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.bindAddr, s.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	var listener net.Listener
	var err error

	if s.socketPath != "" {
		// Remove existing socket if it exists
		if _, err := os.Stat(s.socketPath); err == nil {
			if err := os.Remove(s.socketPath); err != nil {
				return fmt.Errorf("failed to remove existing socket: %w", err)
			}
		}

		listener, err = net.Listen("unix", s.socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen on unix socket: %w", err)
		}

		// Set permissions to allow container access
		if err := os.Chmod(s.socketPath, 0666); err != nil {
			s.logger.Warn("Failed to set socket permissions: %v", err)
		}

		s.logger.Info("Starting API server on unix://%s", s.socketPath)
	} else {
		listener, err = net.Listen("tcp", s.server.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen on tcp: %w", err)
		}
		s.logger.Info("Starting API server on http://%s:%d", s.bindAddr, s.port)
	}

	errChan := make(chan error, 1)
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		return s.Stop()
	}
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	// Remove socket file if using unix socket
	if s.socketPath != "" {
		if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
			s.logger.Warn("Failed to remove socket file: %v", err)
		}
	}

	return nil
}

// handleHealth returns basic health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleStatus returns detailed daemon status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.daemon.GetStatus()
	if err != nil {
		s.logger.Error("Failed to get status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
