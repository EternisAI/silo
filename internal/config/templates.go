package config

import (
	"fmt"
	"os"
	"text/template"

	"github.com/eternisai/silo/internal/assets"
)

func GenerateDockerCompose(config *Config, output string) error {
	tmpl, err := template.New("docker-compose").Parse(assets.DockerComposeTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose template: %w", err)
	}

	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create docker-compose file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, config); err != nil {
		return fmt.Errorf("failed to execute docker-compose template: %w", err)
	}

	return nil
}

func GenerateConfig(config *Config, output string) error {
	tmpl, err := template.New("config").Parse(assets.ConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, config); err != nil {
		return fmt.Errorf("failed to execute config template: %w", err)
	}

	return nil
}
