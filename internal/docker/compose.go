package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Container struct {
	Name    string
	State   string
	Status  string
	Image   string
	Service string
}

type LogOptions struct {
	Follow bool
	Lines  int
}

func Up(ctx context.Context, composePath string) error {
	composeCmd := GetComposeCommand()
	args := append(composeCmd[1:], "-f", composePath, "up", "-d")
	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(composePath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}
	return nil
}

func Down(ctx context.Context, composePath string, removeVolumes bool) error {
	composeCmd := GetComposeCommand()
	args := append(composeCmd[1:], "-f", composePath, "down")
	if removeVolumes {
		args = append(args, "-v")
	}

	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(composePath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop containers: %w", err)
	}
	return nil
}

func Pull(ctx context.Context, composePath string) error {
	composeCmd := GetComposeCommand()
	services := []string{"backend", "frontend"}
	args := append(composeCmd[1:], "-f", composePath, "pull")
	args = append(args, services...)

	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(composePath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}
	return nil
}

func Ps(ctx context.Context, composePath string) ([]Container, error) {
	composeCmd := GetComposeCommand()
	args := append(composeCmd[1:], "-f", composePath, "ps", "-q")
	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Dir = filepath.Dir(composePath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := strings.TrimSpace(string(output))
	if containerIDs == "" {
		return []Container{}, nil
	}

	ids := strings.Split(containerIDs, "\n")

	inspectCmd := exec.CommandContext(ctx, "docker", "inspect", "--format",
		"{{.Name}}|{{.State.Status}}|{{.State.Status}}|{{.Config.Image}}|{{index .Config.Labels \"com.docker.compose.service\"}}")
	inspectCmd.Args = append(inspectCmd.Args, ids...)

	inspectOutput, err := inspectCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect containers: %w", err)
	}

	var containers []Container
	lines := strings.Split(strings.TrimSpace(string(inspectOutput)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 5 {
			continue
		}

		name := strings.TrimPrefix(parts[0], "/")
		containers = append(containers, Container{
			Name:    name,
			State:   parts[1],
			Status:  parts[2],
			Image:   parts[3],
			Service: parts[4],
		})
	}

	return containers, nil
}

func Logs(ctx context.Context, composePath string, service string, opts LogOptions) error {
	composeCmd := GetComposeCommand()
	args := append(composeCmd[1:], "-f", composePath, "logs")

	if opts.Follow {
		args = append(args, "-f")
	}
	if opts.Lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.Lines))
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(composePath)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		return fmt.Errorf("failed to fetch logs: %w", err)
	}
	return nil
}

func Exec(ctx context.Context, composePath string, service string, command []string) error {
	composeCmd := GetComposeCommand()
	args := append(composeCmd[1:], "-f", composePath, "exec", "-T", service)
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(composePath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command in container: %w", err)
	}
	return nil
}

func Restart(ctx context.Context, composePath string, service string) error {
	composeCmd := GetComposeCommand()
	args := append(composeCmd[1:], "-f", composePath, "restart")
	if service != "" {
		args = append(args, service)
	}

	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(composePath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart containers: %w", err)
	}
	return nil
}

func IsRunning(ctx context.Context, composePath string) (bool, error) {
	containers, err := Ps(ctx, composePath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return false, nil
		}
		return false, err
	}

	for _, c := range containers {
		if c.State == "running" {
			return true, nil
		}
	}
	return false, nil
}
