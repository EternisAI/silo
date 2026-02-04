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
	DefaultContainerName = "glm_model"
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
	return []string{
		"run", "-d",
		"--name", e.getContainerName(),
		"--restart", "unless-stopped",
		"--gpus", `"device=0,1,2"`,
		"--shm-size", "64g",
		"--ipc=host",
		"--ulimit", "memlock=-1:-1",
		"--ulimit", "nofile=1048576:1048576",
		"-p", "30000:30000",
		"-e", "CUDA_VISIBLE_DEVICES=0,1,2",
		"-e", "PYTORCH_ALLOC_CONF=expandable_segments:True",
		"-v", "/root/data/AWQ:/workspace/model",
		"-v", os.ExpandEnv("$HOME/.cache/huggingface") + ":/root/.cache/huggingface",
		"lmsysorg/sglang:latest",
		"python3", "-m", "sglang.launch_server",
		"--model-path", "/workspace/model",
		"--host", "0.0.0.0",
		"--port", "30000",
		"--dp-size", "3",
		"--tp-size", "1",
		"--max-running-requests", "32",
		"--max-total-tokens", "262144",
		"--context-length", "131072",
		"--mem-fraction-static", "0.88",
		"--chunked-prefill-size", "-1",
		"--schedule-policy", "fcfs",
		"--kv-cache-dtype", "fp8_e4m3",
		"--attention-backend", "flashinfer",
		"--disable-radix-cache",
		"--reasoning-parser", "glm45",
		"--tool-call-parser", "glm",
		"--trust-remote-code",
		"--log-level", "info",
	}
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
