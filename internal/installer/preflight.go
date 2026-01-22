package installer

import (
	"fmt"
	"net"
	"syscall"

	"github.com/eternisai/silo/internal/docker"
)

const RequiredDiskSpaceGB = 5

func CheckSystemRequirements() error {
	if err := docker.ValidateRequirements(); err != nil {
		return err
	}
	return nil
}

func CheckDiskSpace(path string, requiredGB int) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	availableGB := (stat.Bavail * uint64(stat.Bsize)) / (1024 * 1024 * 1024)
	if availableGB < uint64(requiredGB) {
		return fmt.Errorf("insufficient disk space: %d GB available, %d GB required", availableGB, requiredGB)
	}

	return nil
}

func CheckPortAvailability(port int) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use", port)
	}
	listener.Close()
	return nil
}
