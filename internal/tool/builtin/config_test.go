package builtin

import (
	"testing"

	"ai-ops/internal/tool"
)

func TestGetConfigTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"host": "test-host",
				"path": "/etc/nginx/nginx.conf",
			},
			expectError: true,
		},
		{
			name: "missing host",
			params: map[string]interface{}{
				"path": "/etc/nginx/nginx.conf",
			},
			expectError: true,
		},
		{
			name: "missing path",
			params: map[string]interface{}{
				"host": "test-host",
			},
			expectError: true,
		},
	}

	tool := NewGetConfigTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(nil, tt.params)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			if tt.expectError && result.Success {
				t.Errorf("Expected error but got success")
			}
			if !tt.expectError && !result.Success {
				t.Errorf("Expected success but got error: %s", result.Message)
			}
		})
	}
}

func TestUpdateConfigTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid parameters with backup",
			params: map[string]interface{}{
				"host":    "test-host",
				"path":    "/etc/nginx/nginx.conf",
				"content": "server { listen 80; }",
				"backup":  true,
			},
			expectError: true,
		},
		{
			name: "valid parameters without backup",
			params: map[string]interface{}{
				"host":    "test-host",
				"path":    "/etc/nginx/nginx.conf",
				"content": "server { listen 80; }",
				"backup":  false,
			},
			expectError: true,
		},
		{
			name: "missing host",
			params: map[string]interface{}{
				"path":    "/etc/nginx/nginx.conf",
				"content": "server { listen 80; }",
			},
			expectError: true,
		},
		{
			name: "missing path",
			params: map[string]interface{}{
				"host":    "test-host",
				"content": "server { listen 80; }",
			},
			expectError: true,
		},
		{
			name: "missing content",
			params: map[string]interface{}{
				"host": "test-host",
				"path": "/etc/nginx/nginx.conf",
			},
			expectError: true,
		},
	}

	tool := NewUpdateConfigTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(nil, tt.params)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			if tt.expectError && result.Success {
				t.Errorf("Expected error but got success")
			}
			if !tt.expectError && !result.Success {
				t.Errorf("Expected success but got error: %s", result.Message)
			}
		})
	}
}

func TestCheckSSLCertTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid parameters with default port",
			params: map[string]interface{}{
				"host":   "test-host",
				"domain": "example.com",
			},
			expectError: true,
		},
		{
			name: "valid parameters with custom port",
			params: map[string]interface{}{
				"host":   "test-host",
				"domain": "example.com",
				"port":   8443,
			},
			expectError: true,
		},
		{
			name: "missing host",
			params: map[string]interface{}{
				"domain": "example.com",
			},
			expectError: true,
		},
		{
			name: "missing domain",
			params: map[string]interface{}{
				"host": "test-host",
			},
			expectError: true,
		},
	}

	tool := NewCheckSSLCertTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(nil, tt.params)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			if tt.expectError && result.Success {
				t.Errorf("Expected error but got success")
			}
			if !tt.expectError && !result.Success {
				t.Errorf("Expected success but got error: %s", result.Message)
			}
		})
	}
}

func TestConfigToolsMetadata(t *testing.T) {
	tests := []struct {
		name string
		tool tool.Tool
	}{
		{"GetConfigTool", NewGetConfigTool()},
		{"UpdateConfigTool", NewUpdateConfigTool()},
		{"CheckSSLCertTool", NewCheckSSLCertTool()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tool.Name() == "" {
				t.Error("Tool name should not be empty")
			}
			if tt.tool.Description() == "" {
				t.Error("Tool description should not be empty")
			}
			if len(tt.tool.Parameters()) == 0 {
				t.Error("Tool should have at least one parameter")
			}
		})
	}
}
