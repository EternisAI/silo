package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eternisai/silo/pkg/logger"
)

// Server provides HTTP API for daemon
type Server struct {
	port   int
	daemon *Daemon
	logger *logger.Logger
	server *http.Server
}

// NewServer creates a new HTTP server
func NewServer(port int, daemon *Daemon, log *logger.Logger) *Server {
	return &Server{
		port:   port,
		daemon: daemon,
		logger: log,
	}
}

// Start begins serving HTTP requests
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/stats", s.handleStats)

	s.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: mux,
	}

	s.logger.Info("Starting API server on http://127.0.0.1:%d", s.port)

	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	return s.server.Shutdown(ctx)
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

// handleStats returns monitoring statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.daemon.monitor.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
