package daemon

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Details string      `json:"details,omitempty"`
	Logs    []LogEntry  `json:"logs,omitempty"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// UpRequest represents request body for /api/v1/up
type UpRequest struct {
	ImageTag              string `json:"image_tag,omitempty"`
	Port                  int    `json:"port,omitempty"`
	EnableInferenceEngine bool   `json:"enable_inference_engine,omitempty"`
	EnableProxyAgent      bool   `json:"enable_proxy_agent,omitempty"`
}

// RestartRequest represents request body for /api/v1/restart
type RestartRequest struct {
	Service string `json:"service,omitempty"`
}
