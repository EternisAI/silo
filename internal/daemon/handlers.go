package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/internal/docker"
	"github.com/eternisai/silo/internal/installer"
	"github.com/eternisai/silo/internal/updater"
	"github.com/eternisai/silo/internal/version"
	"github.com/eternisai/silo/pkg/logger"
)

const (
	// API operation timeouts
	UpTimeout      = 10 * time.Minute
	DownTimeout    = 5 * time.Minute
	RestartTimeout = 5 * time.Minute
	UpgradeTimeout = 10 * time.Minute
	LogsTimeout    = 30 * time.Second
	VersionTimeout = 10 * time.Second
	MaxLogLines    = 10000
)

// handleUp handles POST /api/v1/up - start/install Silo
func (s *Server) handleUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	// Acquire operation lock
	s.daemon.opLock.Lock()
	defer s.daemon.opLock.Unlock()

	// Parse request
	var req UpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is OK, use defaults
		req = UpRequest{}
	}

	// Create API logger
	apiLog := NewAPILogger()

	// Check if already installed
	_, err := os.Stat(s.daemon.paths.ComposeFile)
	if err == nil {
		// Already installed, just start containers
		apiLog.Info("Starting containers...")

		ctx, cancel := context.WithTimeout(r.Context(), UpTimeout)
		defer cancel()

		if err := docker.Up(ctx, s.daemon.paths.ComposeFile); err != nil {
			s.respondWithLogs(w, http.StatusInternalServerError, false, "",
				fmt.Sprintf("Failed to start containers: %v", err), "", apiLog.GetLogs())
			return
		}

		apiLog.Success("Containers started successfully")
		s.respondWithLogs(w, http.StatusOK, true, "Silo started successfully", "", "", apiLog.GetLogs())
		return
	}

	// Not installed, run installation
	apiLog.Info("Installing Silo...")

	// Build config for installation
	cfg := config.NewDefaultConfig(s.daemon.paths)

	// Apply request parameters
	if req.ImageTag != "" {
		cfg.ImageTag = req.ImageTag
	}
	if req.Port > 0 {
		cfg.Port = req.Port
	}
	cfg.EnableInferenceEngine = req.EnableInferenceEngine
	cfg.EnableProxyAgent = req.EnableProxyAgent

	// Validate config
	if err := config.Validate(cfg); err != nil {
		apiLog.Error("Invalid configuration: %v", err)
		s.respondWithLogs(w, http.StatusBadRequest, false, "",
			fmt.Sprintf("Invalid configuration: %v", err), "", apiLog.GetLogs())
		return
	}

	// Run installer (using daemon logger for actual installation)
	ctx, cancel := context.WithTimeout(r.Context(), UpTimeout)
	defer cancel()

	// Use a quiet logger to avoid outputting to terminal
	quietLog := logger.New(true)
	inst := installer.New(cfg, s.daemon.paths, quietLog)
	if err := inst.Install(ctx); err != nil {
		apiLog.Error("Installation failed: %v", err)
		s.respondWithLogs(w, http.StatusInternalServerError, false, "",
			fmt.Sprintf("Installation failed: %v", err), "", apiLog.GetLogs())
		return
	}

	apiLog.Success("Installation completed successfully")
	s.respondWithLogs(w, http.StatusOK, true, "Silo installed and started successfully", "", "", apiLog.GetLogs())
}

// handleDown handles POST /api/v1/down - stop containers
func (s *Server) handleDown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	// Acquire operation lock
	s.daemon.opLock.Lock()
	defer s.daemon.opLock.Unlock()

	apiLog := NewAPILogger()
	apiLog.Info("Stopping containers...")

	ctx, cancel := context.WithTimeout(r.Context(), DownTimeout)
	defer cancel()

	if err := docker.Down(ctx, s.daemon.paths.ComposeFile, false); err != nil {
		apiLog.Error("Failed to stop containers: %v", err)
		s.respondWithLogs(w, http.StatusInternalServerError, false, "",
			fmt.Sprintf("Failed to stop containers: %v", err), "", apiLog.GetLogs())
		return
	}

	apiLog.Success("Containers stopped successfully")
	s.respondWithLogs(w, http.StatusOK, true, "Silo stopped successfully", "", "", apiLog.GetLogs())
}

// handleRestart handles POST /api/v1/restart - restart service(s)
func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	// Acquire operation lock
	s.daemon.opLock.Lock()
	defer s.daemon.opLock.Unlock()

	// Parse request
	var req RestartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = RestartRequest{}
	}

	apiLog := NewAPILogger()

	if req.Service != "" {
		apiLog.Info("Restarting service: %s", req.Service)
	} else {
		apiLog.Info("Restarting all services")
	}

	ctx, cancel := context.WithTimeout(r.Context(), RestartTimeout)
	defer cancel()

	if err := docker.Restart(ctx, s.daemon.paths.ComposeFile, req.Service); err != nil {
		apiLog.Error("Failed to restart: %v", err)
		s.respondWithLogs(w, http.StatusInternalServerError, false, "",
			fmt.Sprintf("Failed to restart: %v", err), "", apiLog.GetLogs())
		return
	}

	apiLog.Success("Restart completed successfully")
	s.respondWithLogs(w, http.StatusOK, true, "Service(s) restarted successfully", "", "", apiLog.GetLogs())
}

// handleUpgrade handles POST /api/v1/upgrade - upgrade to latest version
func (s *Server) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	// Acquire operation lock
	s.daemon.opLock.Lock()
	defer s.daemon.opLock.Unlock()

	apiLog := NewAPILogger()
	apiLog.Info("Starting upgrade process...")

	// Load current config with defaults for missing fields
	cfg, err := config.LoadOrDefault(s.daemon.paths.ConfigFile, s.daemon.paths)
	if err != nil {
		apiLog.Error("Failed to load config: %v", err)
		s.respondWithLogs(w, http.StatusInternalServerError, false, "",
			fmt.Sprintf("Failed to load config: %v", err), "", apiLog.GetLogs())
		return
	}

	// Run updater (using daemon logger for actual update)
	ctx, cancel := context.WithTimeout(r.Context(), UpgradeTimeout)
	defer cancel()

	quietLog := logger.New(true)
	upd := updater.New(cfg, s.daemon.paths, quietLog)
	if err := upd.Update(ctx); err != nil {
		apiLog.Error("Upgrade failed: %v", err)
		s.respondWithLogs(w, http.StatusInternalServerError, false, "",
			fmt.Sprintf("Upgrade failed: %v", err), "", apiLog.GetLogs())
		return
	}

	// Reload daemon config after upgrade with defaults for any new fields
	if newCfg, err := config.LoadOrDefault(s.daemon.paths.ConfigFile, s.daemon.paths); err == nil {
		s.daemon.config = newCfg
	} else {
		s.daemon.logger.Warn("Failed to reload config after upgrade: %v", err)
	}

	apiLog.Success("Upgrade completed successfully")
	s.respondWithLogs(w, http.StatusOK, true, "Upgrade completed successfully", "", "", apiLog.GetLogs())
}

// handleLogs handles GET /api/v1/logs - get container logs
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	// Parse query parameters
	service := r.URL.Query().Get("service")
	linesStr := r.URL.Query().Get("lines")

	lines := 100 // default
	if linesStr != "" {
		if n, err := strconv.Atoi(linesStr); err == nil && n > 0 {
			if n > MaxLogLines {
				lines = MaxLogLines
			} else {
				lines = n
			}
		}
	}

	apiLog := NewAPILogger()
	apiLog.Info("Fetching logs (service=%s, lines=%d)", service, lines)

	ctx, cancel := context.WithTimeout(r.Context(), LogsTimeout)
	defer cancel()

	opts := docker.LogOptions{
		Follow: false,
		Lines:  lines,
	}

	if err := docker.Logs(ctx, s.daemon.paths.ComposeFile, service, opts); err != nil {
		apiLog.Error("Failed to fetch logs: %v", err)
		s.respondWithLogs(w, http.StatusInternalServerError, false, "",
			fmt.Sprintf("Failed to fetch logs: %v", err), "", apiLog.GetLogs())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "Logs command executed (output sent to stdout)",
		Data:    map[string]interface{}{"note": "Logs are output to stdout, not captured in API response"},
	})
}

// handleVersion handles GET /api/v1/version - get version information
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	// Load config to get current version
	cfg, err := config.Load(s.daemon.paths.ConfigFile)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError,
			"Failed to load config", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), VersionTimeout)
	defer cancel()

	// Check CLI version
	cliVer, err := version.Check(ctx, cfg.Version)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError,
			"Failed to check CLI version", err.Error())
		return
	}

	// Check image versions
	imageVers, err := version.CheckImageVersions(ctx, cfg.ImageTag)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError,
			"Failed to check image versions", err.Error())
		return
	}

	data := map[string]interface{}{
		"cli":    cliVer,
		"images": imageVers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "Version information retrieved",
		Data:    data,
	})
}

// handleCheck handles GET /api/v1/check - validate configuration
func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	apiLog := NewAPILogger()
	apiLog.Info("Validating configuration...")

	// Load config with defaults for missing fields (ensures validation includes all required fields)
	cfg, err := config.LoadOrDefault(s.daemon.paths.ConfigFile, s.daemon.paths)
	if err != nil {
		apiLog.Error("Failed to load config: %v", err)
		s.respondWithLogs(w, http.StatusInternalServerError, false, "",
			fmt.Sprintf("Failed to load config: %v", err), "", apiLog.GetLogs())
		return
	}

	if err := config.Validate(cfg); err != nil {
		apiLog.Error("Config validation failed: %v", err)
		s.respondWithLogs(w, http.StatusBadRequest, false, "",
			fmt.Sprintf("Config validation failed: %v", err), "", apiLog.GetLogs())
		return
	}

	// Check required files exist
	files := []string{
		s.daemon.paths.ConfigFile,
		s.daemon.paths.StateFile,
		s.daemon.paths.ComposeFile,
	}

	for _, file := range files {
		if !fileExists(file) {
			apiLog.Error("Required file missing: %s", file)
			s.respondWithLogs(w, http.StatusInternalServerError, false, "",
				fmt.Sprintf("Required file missing: %s", file), "", apiLog.GetLogs())
			return
		}
	}

	apiLog.Success("Configuration is valid")
	s.respondWithLogs(w, http.StatusOK, true, "Configuration is valid", "", "", apiLog.GetLogs())
}

// respondError sends an error response
func (s *Server) respondError(w http.ResponseWriter, status int, error, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   error,
		Details: details,
	})
}

// respondWithLogs sends a response with log entries
func (s *Server) respondWithLogs(w http.ResponseWriter, status int, success bool, message, error, details string, logs []LogEntry) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: success,
		Message: message,
		Error:   error,
		Details: details,
		Logs:    logs,
	})
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
