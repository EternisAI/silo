package docker

import (
	"fmt"
	"os/exec"
)

func CheckDockerInstalled() error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("docker is not installed. Please install Docker: https://docs.docker.com/get-docker/")
	}
	return nil
}

func CheckDockerRunning() error {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not running. Please start the Docker daemon")
	}
	return nil
}

func CheckDockerComposeInstalled() error {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return nil
	}

	_, err := exec.LookPath("docker-compose")
	if err != nil {
		return fmt.Errorf("docker compose is not installed. Please install Docker Compose: https://docs.docker.com/compose/install/")
	}
	return nil
}

func GetComposeCommand() []string {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return []string{"docker", "compose"}
	}
	return []string{"docker-compose"}
}

func ValidateRequirements() error {
	if err := CheckDockerInstalled(); err != nil {
		return err
	}
	if err := CheckDockerRunning(); err != nil {
		return err
	}
	if err := CheckDockerComposeInstalled(); err != nil {
		return err
	}
	return nil
}
