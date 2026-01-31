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
	if cfg.EnableInferenceEngine != DefaultEnableInferenceEngine {
		t.Errorf("Expected EnableInferenceEngine=%v, got %v", DefaultEnableInferenceEngine, cfg.EnableInferenceEngine)
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
default_model: "custom-model.gguf"
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
	if cfg.InferencePort != DefaultInferencePort {
		t.Errorf("Expected default InferencePort=%d, got %d", DefaultInferencePort, cfg.InferencePort)
	}
	if cfg.EnableInferenceEngine != DefaultEnableInferenceEngine {
		t.Errorf("Expected default EnableInferenceEngine=%v, got %v", DefaultEnableInferenceEngine, cfg.EnableInferenceEngine)
	}
	if cfg.EnableProxyAgent != DefaultEnableProxyAgent {
		t.Errorf("Expected default EnableProxyAgent=%v, got %v", DefaultEnableProxyAgent, cfg.EnableProxyAgent)
	}
	if cfg.InferenceGPULayers != DefaultInferenceGPULayers {
		t.Errorf("Expected default InferenceGPULayers=%d, got %d", DefaultInferenceGPULayers, cfg.InferenceGPULayers)
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
default_model: "my-model.gguf"
inference_port: 40000
inference_model_file: "my-inference.gguf"
inference_shm_size: "32g"
inference_context_size: 16384
inference_batch_size: 512
inference_gpu_layers: 50
inference_tensor_split: "2,2,2"
inference_main_gpu: 1
inference_threads: 32
inference_http_threads: 16
inference_fit: "on"
inference_gpu_devices: "\"1\", \"2\""
enable_inference_engine: true
enable_proxy_agent: true
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
	if cfg.InferencePort != 40000 {
		t.Errorf("Expected preserved InferencePort=40000, got %d", cfg.InferencePort)
	}
	if cfg.InferenceGPULayers != 50 {
		t.Errorf("Expected preserved InferenceGPULayers=50, got %d", cfg.InferenceGPULayers)
	}
	if !cfg.EnableInferenceEngine {
		t.Errorf("Expected preserved EnableInferenceEngine=true, got false")
	}
	if !cfg.EnableProxyAgent {
		t.Errorf("Expected preserved EnableProxyAgent=true, got false")
	}
}

func TestMergeConfigs(t *testing.T) {
	existing := &Config{
		Version:  "0.1.1",
		ImageTag: "0.1.1",
		Port:     8080,
		// Missing: LLMBaseURL, DefaultModel, InferencePort, etc.
	}

	defaults := &Config{
		Version:               DefaultVersion,
		ImageTag:              DefaultImageTag,
		Port:                  DefaultPort,
		LLMBaseURL:            DefaultLLMBaseURL,
		DefaultModel:          DefaultModel,
		InferencePort:         DefaultInferencePort,
		EnableInferenceEngine: DefaultEnableInferenceEngine,
		EnableProxyAgent:      DefaultEnableProxyAgent,
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
	if merged.InferencePort != DefaultInferencePort {
		t.Errorf("Expected merged.InferencePort=%d, got %d", DefaultInferencePort, merged.InferencePort)
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
		"llm_base_url":            true,
		"default_model":           true,
		"enable_inference_engine": true,
		"enable_proxy_agent":      true,
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
