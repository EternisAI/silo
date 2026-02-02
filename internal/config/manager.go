package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

const (
	DefaultVersion    = "0.1.6"
	DefaultImageTag   = "0.1.8"
	DefaultPort       = 80
	DefaultLLMBaseURL = "http://host.docker.internal:30000/v1"
	DefaultModel      = "glm47-awq"

	// Inference engine defaults
	DefaultInferencePort        = 30000
	DefaultInferenceModelFile   = "GLM-4.7-Q4_K_M.gguf"
	DefaultInferenceShmSize     = "16g"
	DefaultInferenceContextSize = 8192
	DefaultInferenceBatchSize   = 256
	DefaultInferenceGPULayers   = 999
	DefaultInferenceTensorSplit = "1,1,1"
	DefaultInferenceMainGPU     = 0
	DefaultInferenceThreads     = 16
	DefaultInferenceHTTPThreads = 8
	DefaultInferenceFit         = "off"
	DefaultInferenceGPUDevices  = `"0", "1", "2"`

	// Service toggles
	DefaultEnableInferenceEngine = false
	DefaultEnableProxyAgent      = false
)

type Config struct {
	Version      string `yaml:"version"`
	ImageTag     string `yaml:"image_tag"`
	Port         int    `yaml:"port"`
	LLMBaseURL   string `yaml:"llm_base_url"`
	DefaultModel string `yaml:"default_model"`
	ConfigFile   string `yaml:"-"`
	DataDir      string `yaml:"-"`
	SocketFile   string `yaml:"-"`

	// Inference engine configuration
	InferencePort        int    `yaml:"inference_port"`
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

	// Service toggles
	EnableInferenceEngine bool `yaml:"enable_inference_engine"`
	EnableProxyAgent      bool `yaml:"enable_proxy_agent"`
}

type State struct {
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	LastUpdated string `json:"last_updated"`
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
		SocketFile:   paths.SocketFile,

		// Inference engine defaults
		InferencePort:        DefaultInferencePort,
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

		// Service toggles
		EnableInferenceEngine: DefaultEnableInferenceEngine,
		EnableProxyAgent:      DefaultEnableProxyAgent,
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

// LoadOrDefault loads config from file and fills missing fields with defaults.
// If the file doesn't exist, returns a new config with all defaults.
// If the file exists, preserves existing values and only fills in missing/zero fields.
func LoadOrDefault(path string, paths *Paths) (*Config, error) {
	existing, err := Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewDefaultConfig(paths), nil
		}
		return nil, err
	}

	defaults := NewDefaultConfig(paths)
	merged := mergeConfigs(existing, defaults)
	merged.ConfigFile = paths.ConfigFile
	merged.DataDir = paths.AppDataDir
	merged.SocketFile = paths.SocketFile

	return merged, nil
}

// mergeConfigs merges two configs, preferring non-zero values from existing config
// and filling in zero values with defaults. Uses reflection to handle all fields.
func mergeConfigs(existing, defaults *Config) *Config {
	result := &Config{}

	existingVal := reflect.ValueOf(existing).Elem()
	defaultsVal := reflect.ValueOf(defaults).Elem()
	resultVal := reflect.ValueOf(result).Elem()

	for i := 0; i < existingVal.NumField(); i++ {
		existingField := existingVal.Field(i)
		defaultField := defaultsVal.Field(i)
		resultField := resultVal.Field(i)

		if !resultField.CanSet() {
			continue
		}

		if isZeroValue(existingField) {
			resultField.Set(defaultField)
		} else {
			resultField.Set(existingField)
		}
	}

	return result
}

// isZeroValue checks if a reflect.Value is the zero value for its type
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	default:
		return false
	}
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

// UpdateImageTag updates the image_tag field in the config and saves it.
// This function uses LoadOrDefault to ensure any new fields are preserved.
func UpdateImageTag(cfg *Config, newTag string, configPath string) error {
	cfg.ImageTag = newTag

	if err := Validate(cfg); err != nil {
		return fmt.Errorf("config validation failed with new tag: %w", err)
	}

	if err := Save(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
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

	// Inference engine validation (only when enabled)
	if config.EnableInferenceEngine {
		if config.InferencePort < 1 || config.InferencePort > 65535 {
			return fmt.Errorf("inference_port must be between 1 and 65535")
		}
		if config.InferenceModelFile == "" {
			return fmt.Errorf("inference_model_file cannot be empty")
		}
		if config.InferenceContextSize < 1 {
			return fmt.Errorf("inference_context_size must be positive")
		}
		if config.InferenceBatchSize < 1 {
			return fmt.Errorf("inference_batch_size must be positive")
		}
		if config.InferenceThreads < 1 {
			return fmt.Errorf("inference_threads must be positive")
		}
		if config.InferenceHTTPThreads < 1 {
			return fmt.Errorf("inference_http_threads must be positive")
		}
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

// FindUnknownFields returns a list of field names in the config file that are not recognized.
func FindUnknownFields(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into a map to get all keys
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Get known fields from Config struct tags
	knownFields := make(map[string]bool)
	t := reflect.TypeOf(Config{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag != "" && tag != "-" {
			knownFields[tag] = true
		}
	}

	// Find unknown fields
	var unknown []string
	for key := range rawConfig {
		if !knownFields[key] {
			unknown = append(unknown, key)
		}
	}

	return unknown, nil
}

// FindMissingFields returns a list of field names that are missing from the config file
// but exist in the Config struct (useful for detecting when defaults will be applied).
func FindMissingFields(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into a map to get all keys
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Get all known fields from Config struct tags
	t := reflect.TypeOf(Config{})
	var missing []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag != "" && tag != "-" {
			if _, exists := rawConfig[tag]; !exists {
				missing = append(missing, tag)
			}
		}
	}

	return missing, nil
}
