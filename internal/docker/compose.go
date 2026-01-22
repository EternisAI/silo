package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	args := append(composeCmd[1:], "-f", composePath, "pull")
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
	args := append(composeCmd[1:], "-f", composePath, "ps", "--format", "json")
	cmd := exec.CommandContext(ctx, composeCmd[0], args...)
	cmd.Dir = filepath.Dir(composePath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var containers []Container
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var c struct {
			Name    string `json:"Name"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Image   string `json:"Image"`
			Service string `json:"Service"`
		}
		if err := decoder.Decode(&c); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to parse container info: %w", err)
		}

		containers = append(containers, Container{
			Name:    c.Name,
			State:   c.State,
			Status:  c.Status,
			Image:   c.Image,
			Service: c.Service,
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
