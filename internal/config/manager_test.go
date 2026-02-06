package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrDefault_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &Paths{
		ConfigDir:   tmpDir,
		ConfigFile:  filepath.Join(tmpDir, "config.yml"),
		DataDir:     filepath.Join(tmpDir, "data"),
		AppDataDir:  filepath.Join(tmpDir, "data", "app"),
		StateFile:   filepath.Join(tmpDir, "state.json"),
		ComposeFile: filepath.Join(tmpDir, "docker-compose.yml"),
	}

	cfg, err := LoadOrDefault(paths.ConfigFile, paths)
	if err != nil {
		t.Fatalf("LoadOrDefault failed: %v", err)
	}

	if cfg.Version != DefaultVersion {
		t.Errorf("Expected Version=%s, got %s", DefaultVersion, cfg.Version)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("Expected Port=%d, got %d", DefaultPort, cfg.Port)
	}
	if cfg.EnableProxyAgent != DefaultEnableProxyAgent {
		t.Errorf("Expected EnableProxyAgent=%v, got %v", DefaultEnableProxyAgent, cfg.EnableProxyAgent)
	}
}

func TestLoadOrDefault_ExistingFileWithMissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &Paths{
		ConfigDir:   tmpDir,
		ConfigFile:  filepath.Join(tmpDir, "config.yml"),
		DataDir:     filepath.Join(tmpDir, "data"),
		AppDataDir:  filepath.Join(tmpDir, "data", "app"),
		StateFile:   filepath.Join(tmpDir, "state.json"),
		ComposeFile: filepath.Join(tmpDir, "docker-compose.yml"),
	}

	// Create a config file with only some fields (simulating old config)
	partialConfig := `version: "0.1.2"
image_tag: "0.1.2"
port: 8080
llm_base_url: "http://custom-url:9000/v1"
default_model: "custom-model"
`
	if err := os.WriteFile(paths.ConfigFile, []byte(partialConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadOrDefault(paths.ConfigFile, paths)
	if err != nil {
		t.Fatalf("LoadOrDefault failed: %v", err)
	}

	// Check that existing values are preserved
	if cfg.Version != "0.1.2" {
		t.Errorf("Expected preserved Version=0.1.2, got %s", cfg.Version)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected preserved Port=8080, got %d", cfg.Port)
	}
	if cfg.LLMBaseURL != "http://custom-url:9000/v1" {
		t.Errorf("Expected preserved LLMBaseURL=http://custom-url:9000/v1, got %s", cfg.LLMBaseURL)
	}

	// Check that missing fields are filled with defaults
	if cfg.EnableProxyAgent != DefaultEnableProxyAgent {
		t.Errorf("Expected default EnableProxyAgent=%v, got %v", DefaultEnableProxyAgent, cfg.EnableProxyAgent)
	}
	if cfg.EnableDeepResearch != DefaultEnableDeepResearch {
		t.Errorf("Expected default EnableDeepResearch=%v, got %v", DefaultEnableDeepResearch, cfg.EnableDeepResearch)
	}
}

func TestLoadOrDefault_ExistingFileWithAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &Paths{
		ConfigDir:   tmpDir,
		ConfigFile:  filepath.Join(tmpDir, "config.yml"),
		DataDir:     filepath.Join(tmpDir, "data"),
		AppDataDir:  filepath.Join(tmpDir, "data", "app"),
		StateFile:   filepath.Join(tmpDir, "state.json"),
		ComposeFile: filepath.Join(tmpDir, "docker-compose.yml"),
	}

	// Create a complete config file
	completeConfig := `version: "0.1.2"
image_tag: "0.1.2"
port: 9999
llm_base_url: "http://my-llm:5000/v1"
default_model: "my-model"
enable_proxy_agent: true
enable_deep_research: true
`
	if err := os.WriteFile(paths.ConfigFile, []byte(completeConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadOrDefault(paths.ConfigFile, paths)
	if err != nil {
		t.Fatalf("LoadOrDefault failed: %v", err)
	}

	// All custom values should be preserved
	if cfg.Port != 9999 {
		t.Errorf("Expected preserved Port=9999, got %d", cfg.Port)
	}
	if !cfg.EnableProxyAgent {
		t.Errorf("Expected preserved EnableProxyAgent=true, got false")
	}
	if !cfg.EnableDeepResearch {
		t.Errorf("Expected preserved EnableDeepResearch=true, got false")
	}
}

func TestMergeConfigs(t *testing.T) {
	existing := &Config{
		Version:  "0.1.1",
		ImageTag: "0.1.1",
		Port:     8080,
		// Missing: LLMBaseURL, DefaultModel, etc.
	}

	defaults := &Config{
		Version:          DefaultVersion,
		ImageTag:         DefaultImageTag,
		Port:             DefaultPort,
		LLMBaseURL:       DefaultLLMBaseURL,
		DefaultModel:     DefaultModel,
		EnableProxyAgent: DefaultEnableProxyAgent,
	}

	merged := mergeConfigs(existing, defaults)

	// Existing non-zero values should be preserved
	if merged.Version != "0.1.1" {
		t.Errorf("Expected merged.Version=0.1.1, got %s", merged.Version)
	}
	if merged.Port != 8080 {
		t.Errorf("Expected merged.Port=8080, got %d", merged.Port)
	}

	// Zero values should be filled from defaults
	if merged.LLMBaseURL != DefaultLLMBaseURL {
		t.Errorf("Expected merged.LLMBaseURL=%s, got %s", DefaultLLMBaseURL, merged.LLMBaseURL)
	}
	if merged.DefaultModel != DefaultModel {
		t.Errorf("Expected merged.DefaultModel=%s, got %s", DefaultModel, merged.DefaultModel)
	}
}

func TestFindMissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Create a config file with only some fields
	partialConfig := `version: "0.1.2"
image_tag: "0.1.2"
port: 8080
`
	if err := os.WriteFile(configPath, []byte(partialConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	missing, err := FindMissingFields(configPath)
	if err != nil {
		t.Fatalf("FindMissingFields failed: %v", err)
	}

	// Check that we detected some missing fields
	if len(missing) == 0 {
		t.Error("Expected to find missing fields, but got none")
	}

	// Check that specific fields are in the missing list
	expectedMissing := map[string]bool{
		"llm_base_url":       true,
		"default_model":      true,
		"enable_proxy_agent": true,
	}

	for _, field := range missing {
		if expectedMissing[field] {
			delete(expectedMissing, field)
		}
	}

	if len(expectedMissing) > 0 {
		for field := range expectedMissing {
			t.Errorf("Expected field %s to be in missing list", field)
		}
	}
}
