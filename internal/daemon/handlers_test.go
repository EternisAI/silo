package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondError(t *testing.T) {
	s := &Server{}
	w := httptest.NewRecorder()

	s.respondError(w, http.StatusBadRequest, "test error", "test details")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("Expected success to be false")
	}
	if resp.Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", resp.Error)
	}
	if resp.Details != "test details" {
		t.Errorf("Expected details 'test details', got '%s'", resp.Details)
	}
}

func TestRespondWithLogs(t *testing.T) {
	s := &Server{}
	w := httptest.NewRecorder()

	logs := []LogEntry{
		{Timestamp: "2026-01-30T00:00:00Z", Level: "info", Message: "test message"},
	}

	s.respondWithLogs(w, http.StatusOK, true, "success", "", "", logs)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", resp.Message)
	}
	if len(resp.Logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(resp.Logs))
	}
}

func TestHandleHealthEndpoint(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", resp["status"])
	}
}

func TestMethodNotAllowed(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name     string
		handler  http.HandlerFunc
		method   string
		path     string
		wantCode int
	}{
		{
			name:     "UP with GET",
			handler:  s.handleUp,
			method:   http.MethodGet,
			path:     "/api/v1/up",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "DOWN with GET",
			handler:  s.handleDown,
			method:   http.MethodGet,
			path:     "/api/v1/down",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "RESTART with GET",
			handler:  s.handleRestart,
			method:   http.MethodGet,
			path:     "/api/v1/restart",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "UPGRADE with GET",
			handler:  s.handleUpgrade,
			method:   http.MethodGet,
			path:     "/api/v1/upgrade",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "LOGS with POST",
			handler:  s.handleLogs,
			method:   http.MethodPost,
			path:     "/api/v1/logs",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "VERSION with POST",
			handler:  s.handleVersion,
			method:   http.MethodPost,
			path:     "/api/v1/version",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "CHECK with POST",
			handler:  s.handleCheck,
			method:   http.MethodPost,
			path:     "/api/v1/check",
			wantCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestUpRequestParsing(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			body:    `{"image_tag":"0.1.2","port":8080}`,
			wantErr: false,
		},
		{
			name:    "empty body",
			body:    "",
			wantErr: false, // Empty body should use defaults
		},
		{
			name:    "partial JSON",
			body:    `{"image_tag":"0.1.2"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req UpRequest
			body := bytes.NewBufferString(tt.body)
			err := json.NewDecoder(body).Decode(&req)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil && tt.body != "" {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRestartRequestParsing(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{
			name:    "with service",
			body:    `{"service":"backend"}`,
			want:    "backend",
			wantErr: false,
		},
		{
			name:    "empty service",
			body:    `{"service":""}`,
			want:    "",
			wantErr: false,
		},
		{
			name:    "no service field",
			body:    `{}`,
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req RestartRequest
			body := bytes.NewBufferString(tt.body)
			err := json.NewDecoder(body).Decode(&req)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if req.Service != tt.want {
				t.Errorf("Expected service '%s', got '%s'", tt.want, req.Service)
			}
		})
	}
}
