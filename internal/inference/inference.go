package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/pkg/logger"
)

const (
	DefaultContainerName = "silo-inference"
	DefaultImage         = "lmsysorg/sglang:latest"
)

// Engine manages the inference engine container
type Engine struct {
	cfg    *config.Config
	logger *logger.Logger
}

// ContainerInfo holds information about the inference container
type ContainerInfo struct {
	Name    string
	State   string
	Status  string
	Image   string
	Running bool
}

// New creates a new inference engine manager
func New(cfg *config.Config, log *logger.Logger) *Engine {
	return &Engine{
		cfg:    cfg,
		logger: log,
	}
}

// Up starts the inference engine container
func (e *Engine) Up(ctx context.Context) error {
	// Check if already running
	running, err := e.IsRunning(ctx)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	if running {
		e.logger.Info("Inference engine is already running")
		return nil
	}

	// Remove existing stopped container if present
	if exists, _ := e.containerExists(ctx); exists {
		e.logger.Info("Removing existing stopped container...")
		if err := e.removeContainer(ctx); err != nil {
			return fmt.Errorf("failed to remove existing container: %w", err)
		}
	}

	e.logger.Info("Starting inference engine...")

	args := e.buildDockerRunArgs()
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start inference engine: %w", err)
	}

	e.logger.Success("Inference engine started")
	return nil
}

// Down stops the inference engine container
func (e *Engine) Down(ctx context.Context) error {
	running, err := e.IsRunning(ctx)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	if !running {
		e.logger.Info("Inference engine is not running")
		return nil
	}

	e.logger.Info("Stopping inference engine...")

	containerName := e.getContainerName()

	// Stop container
	stopCmd := exec.CommandContext(ctx, "docker", "stop", containerName)
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Remove container
	rmCmd := exec.CommandContext(ctx, "docker", "rm", containerName)
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	e.logger.Success("Inference engine stopped")
	return nil
}

// Status returns the current status of the inference engine
func (e *Engine) Status(ctx context.Context) (*ContainerInfo, error) {
	containerName := e.getContainerName()

	cmd := exec.CommandContext(ctx, "docker", "inspect",
		"--format", "{{.State.Status}}|{{.State.Running}}|{{.Config.Image}}",
		containerName)

	output, err := cmd.Output()
	if err != nil {
		// Container doesn't exist
		return &ContainerInfo{
			Name:    containerName,
			State:   "not found",
			Running: false,
		}, nil
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected docker inspect output")
	}

	return &ContainerInfo{
		Name:    containerName,
		State:   parts[0],
		Status:  parts[0],
		Image:   parts[2],
		Running: parts[1] == "true",
	}, nil
}

// Logs streams logs from the inference engine container
func (e *Engine) Logs(ctx context.Context, follow bool, lines int) error {
	containerName := e.getContainerName()

	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}
	args = append(args, containerName)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		return fmt.Errorf("failed to fetch logs: %w", err)
	}
	return nil
}

// IsRunning checks if the inference engine container is running
func (e *Engine) IsRunning(ctx context.Context) (bool, error) {
	info, err := e.Status(ctx)
	if err != nil {
		return false, err
	}
	return info.Running, nil
}

// containerExists checks if the container exists (running or stopped)
func (e *Engine) containerExists(ctx context.Context) (bool, error) {
	containerName := e.getContainerName()
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerName)
	err := cmd.Run()
	return err == nil, nil
}

// removeContainer removes the container
func (e *Engine) removeContainer(ctx context.Context) error {
	containerName := e.getContainerName()
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerName)
	return cmd.Run()
}

// getContainerName returns the container name from config or default
func (e *Engine) getContainerName() string {
	if e.cfg.SGLang.ContainerName != "" {
		return e.cfg.SGLang.ContainerName
	}
	return DefaultContainerName
}

// getImage returns the image from config or default
func (e *Engine) getImage() string {
	if e.cfg.SGLang.Image != "" {
		return e.cfg.SGLang.Image
	}
	return DefaultImage
}

// buildDockerRunArgs builds the docker run command arguments
func (e *Engine) buildDockerRunArgs() []string {
	cfg := e.cfg.SGLang

	args := []string{
		"run", "-d",
		"--name", e.getContainerName(),
		"--restart", "unless-stopped",
	}

	// GPU configuration
	if len(cfg.GPUDevices) > 0 {
		gpuDevices := strings.Join(cfg.GPUDevices, ",")
		args = append(args, "--gpus", fmt.Sprintf(`"device=%s"`, gpuDevices))
	}

	// Memory and IPC settings
	if cfg.ShmSize != "" {
		args = append(args, "--shm-size", cfg.ShmSize)
	}
	args = append(args, "--ipc=host")

	// Ulimits
	args = append(args,
		"--ulimit", "memlock=-1:-1",
		"--ulimit", "nofile=1048576:1048576",
	)

	// Port
	port := cfg.Port
	if port == 0 {
		port = 30000
	}
	args = append(args, "-p", fmt.Sprintf("%d:%d", port, port))

	// Environment variables
	if len(cfg.GPUDevices) > 0 {
		args = append(args, "-e", fmt.Sprintf("CUDA_VISIBLE_DEVICES=%s", strings.Join(cfg.GPUDevices, ",")))
	}
	args = append(args, "-e", "PYTORCH_ALLOC_CONF=expandable_segments:True")

	// Volume mounts
	if cfg.ModelPath != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/workspace/model", cfg.ModelPath))
	}
	if cfg.HuggingFaceCache != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/root/.cache/huggingface", cfg.HuggingFaceCache))
	}

	// Image
	args = append(args, e.getImage())

	// SGLang command
	sglangArgs := e.buildSGLangArgs()
	args = append(args, sglangArgs...)

	return args
}

// buildSGLangArgs builds the sglang launch server arguments
func (e *Engine) buildSGLangArgs() []string {
	cfg := e.cfg.SGLang

	args := []string{
		"python3", "-m", "sglang.launch_server",
		"--model-path", "/workspace/model",
		"--host", "0.0.0.0",
		"--port", fmt.Sprintf("%d", cfg.Port),
	}

	// Parallelism settings
	if cfg.DPSize > 0 {
		args = append(args, "--dp-size", fmt.Sprintf("%d", cfg.DPSize))
	}
	if cfg.TPSize > 0 {
		args = append(args, "--tp-size", fmt.Sprintf("%d", cfg.TPSize))
	}

	// Request limits
	if cfg.MaxRunningRequests > 0 {
		args = append(args, "--max-running-requests", fmt.Sprintf("%d", cfg.MaxRunningRequests))
	}
	if cfg.MaxTotalTokens > 0 {
		args = append(args, "--max-total-tokens", fmt.Sprintf("%d", cfg.MaxTotalTokens))
	}
	if cfg.ContextLength > 0 {
		args = append(args, "--context-length", fmt.Sprintf("%d", cfg.ContextLength))
	}

	// Memory settings
	if cfg.MemFractionStatic > 0 {
		args = append(args, "--mem-fraction-static", fmt.Sprintf("%.2f", cfg.MemFractionStatic))
	}
	if cfg.ChunkedPrefillSize > 0 {
		args = append(args, "--chunked-prefill-size", fmt.Sprintf("%d", cfg.ChunkedPrefillSize))
	}

	// Scheduling
	if cfg.SchedulePolicy != "" {
		args = append(args, "--schedule-policy", cfg.SchedulePolicy)
	}

	// KV cache
	if cfg.KVCacheDtype != "" {
		args = append(args, "--kv-cache-dtype", cfg.KVCacheDtype)
	}

	// Attention backend
	if cfg.AttentionBackend != "" {
		args = append(args, "--attention-backend", cfg.AttentionBackend)
	}

	// Flags
	if cfg.DisableRadixCache {
		args = append(args, "--disable-radix-cache")
	}

	// Reasoning parser
	if cfg.ReasoningParser != "" {
		args = append(args, "--reasoning-parser", cfg.ReasoningParser)
	}

	// Trust remote code
	if cfg.TrustRemoteCode {
		args = append(args, "--trust-remote-code")
	}

	// Log level
	logLevel := cfg.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}
	args = append(args, "--log-level", logLevel)

	return args
}

// GetDockerRunCommand returns the full docker run command for debugging
func (e *Engine) GetDockerRunCommand() string {
	args := e.buildDockerRunArgs()
	return "docker " + strings.Join(args, " ")
}

// HealthCheck checks if the inference engine is healthy by calling its API
func (e *Engine) HealthCheck(ctx context.Context) error {
	port := e.cfg.SGLang.Port
	if port == 0 {
		port = 30000
	}

	cmd := exec.CommandContext(ctx, "curl", "-sf",
		fmt.Sprintf("http://localhost:%d/health", port))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}

// InspectRaw returns raw docker inspect output as JSON
func (e *Engine) InspectRaw(ctx context.Context) (map[string]interface{}, error) {
	containerName := e.getContainerName()

	cmd := exec.CommandContext(ctx, "docker", "inspect", containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("container not found: %w", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no container data returned")
	}

	return result[0], nil
}

// WaitForHealthy waits for the inference engine to become healthy
func (e *Engine) WaitForHealthy(ctx context.Context) error {
	e.logger.Info("Waiting for inference engine to become healthy...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := e.HealthCheck(ctx); err == nil {
				e.logger.Success("Inference engine is healthy")
				return nil
			}
			// Wait before retrying (context-aware sleep)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ctx.Done():
			}
		}
	}
}

// LogsBuffer returns logs as a string buffer (for API responses)
func (e *Engine) LogsBuffer(ctx context.Context, lines int) (string, error) {
	containerName := e.getContainerName()

	args := []string{"logs"}
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}
	args = append(args, containerName)

	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to fetch logs: %w", err)
	}

	// Docker logs go to stderr for container stderr
	combined := stdout.String() + stderr.String()
	return combined, nil
}
