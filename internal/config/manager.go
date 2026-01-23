package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DefaultVersion    = "0.1.0"
	DefaultImageTag   = "0.1.1"
	DefaultPort       = 3000
	DefaultLLMBaseURL = "http://inference-engine:30000/v1"
	DefaultModel      = "GLM-4.7-Q4_K_M.gguf"

	// Inference engine defaults
	DefaultInferencePort         = 30000
	DefaultInferenceModelsDir    = "/data/models"
	DefaultInferenceModelFile    = "GLM-4.7-Q4_K_M.gguf"
	DefaultInferenceShmSize      = "16g"
	DefaultInferenceContextSize  = 32768
	DefaultInferenceBatchSize    = 2048
	DefaultInferenceGPULayers    = 999
	DefaultInferenceTensorSplit  = "1,1,1"
	DefaultInferenceMainGPU      = 0
	DefaultInferenceThreads      = 16
	DefaultInferenceHTTPThreads  = 8
	DefaultInferenceFit          = "off"
	DefaultInferenceGPUDevices   = "'0', '1', '2'"
)

type Config struct {
	Version      string `yaml:"version"`
	ImageTag     string `yaml:"image_tag"`
	Port         int    `yaml:"port"`
	LLMBaseURL   string `yaml:"llm_base_url"`
	DefaultModel string `yaml:"default_model"`
	ConfigFile   string `yaml:"-"`
	DataDir      string `yaml:"-"`

	// Inference engine configuration
	InferencePort        int    `yaml:"inference_port"`
	InferenceModelsDir   string `yaml:"inference_models_dir"`
	InferenceModelFile   string `yaml:"inference_model_file"`
	InferenceShmSize     string `yaml:"inference_shm_size"`
	InferenceContextSize int    `yaml:"inference_context_size"`
	InferenceBatchSize   int    `yaml:"inference_batch_size"`
	InferenceGPULayers   int    `yaml:"inference_gpu_layers"`
	InferenceTensorSplit string `yaml:"inference_tensor_split"`
	InferenceMainGPU     int    `yaml:"inference_main_gpu"`
	InferenceThreads     int    `yaml:"inference_threads"`
	InferenceHTTPThreads int    `yaml:"inference_http_threads"`
	InferenceFit         string `yaml:"inference_fit"`
	InferenceGPUDevices  string `yaml:"inference_gpu_devices"`
}

type State struct {
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	LastUpdated string `json:"last_updated"`
	ImageTag    string `json:"image_tag"`
}

func NewDefaultConfig(paths *Paths) *Config {
	return &Config{
		Version:      DefaultVersion,
		ImageTag:     DefaultImageTag,
		Port:         DefaultPort,
		LLMBaseURL:   DefaultLLMBaseURL,
		DefaultModel: DefaultModel,
		ConfigFile:   paths.ConfigFile,
		DataDir:      paths.AppDataDir,

		// Inference engine defaults
		InferencePort:        DefaultInferencePort,
		InferenceModelsDir:   DefaultInferenceModelsDir,
		InferenceModelFile:   DefaultInferenceModelFile,
		InferenceShmSize:     DefaultInferenceShmSize,
		InferenceContextSize: DefaultInferenceContextSize,
		InferenceBatchSize:   DefaultInferenceBatchSize,
		InferenceGPULayers:   DefaultInferenceGPULayers,
		InferenceTensorSplit: DefaultInferenceTensorSplit,
		InferenceMainGPU:     DefaultInferenceMainGPU,
		InferenceThreads:     DefaultInferenceThreads,
		InferenceHTTPThreads: DefaultInferenceHTTPThreads,
		InferenceFit:         DefaultInferenceFit,
		InferenceGPUDevices:  DefaultInferenceGPUDevices,
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func Save(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func Validate(config *Config) error {
	if config.ImageTag == "" {
		return fmt.Errorf("image_tag cannot be empty")
	}
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if config.LLMBaseURL == "" {
		return fmt.Errorf("llm_base_url cannot be empty")
	}
	if config.DefaultModel == "" {
		return fmt.Errorf("default_model cannot be empty")
	}
	return nil
}

func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

func SaveState(path string, state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}
