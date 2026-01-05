package builtin

import (
	"testing"
)

func TestListContainersTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name:        "missing host parameter",
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name: "valid parameters with all=false",
			params: map[string]interface{}{
				"host": "test-host",
				"all":  false,
			},
			expectError: true, // SSH 未初始化会报错
		},
		{
			name: "valid parameters with all=true",
			params: map[string]interface{}{
				"host": "test-host",
				"all":  true,
			},
			expectError: true, // SSH 未初始化会报错
		},
		{
			name: "valid parameters with filter",
			params: map[string]interface{}{
				"host":   "test-host",
				"filter": "name=nginx",
			},
			expectError: true, // SSH 未初始化会报错
		},
	}

	tool := NewListContainersTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(nil, tt.params)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if tt.expectError && result.Success {
				t.Errorf("Expected error but got success")
			}
		})
	}
}

func TestCheckContainerLogsTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name:        "missing host parameter",
			params:      map[string]interface{}{"container": "nginx"},
			expectError: true,
		},
		{
			name:        "missing container parameter",
			params:      map[string]interface{}{"host": "test-host"},
			expectError: true,
		},
		{
			name: "valid parameters with default lines",
			params: map[string]interface{}{
				"host":      "test-host",
				"container": "nginx",
			},
			expectError: true, // SSH 未初始化会报错
		},
		{
			name: "valid parameters with custom lines",
			params: map[string]interface{}{
				"host":      "test-host",
				"container": "nginx",
				"lines":     50,
			},
			expectError: true, // SSH 未初始化会报错
		},
		{
			name: "valid parameters with since",
			params: map[string]interface{}{
				"host":      "test-host",
				"container": "nginx",
				"since":     "1h",
			},
			expectError: true, // SSH 未初始化会报错
		},
	}

	tool := NewCheckContainerLogsTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(nil, tt.params)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			if tt.expectError && result.Success {
				t.Errorf("Expected error but got success")
			}
		})
	}
}

func TestListContainersToolMetadata(t *testing.T) {
	tool := NewListContainersTool()

	if tool.Name() != "list_containers" {
		t.Errorf("Name() = %v, want list_containers", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}

	params := tool.Parameters()
	if len(params) != 3 {
		t.Errorf("Parameters() length = %v, want 3", len(params))
	}

	// 验证必需参数
	hostParam := params[0]
	if hostParam.Name != "host" || !hostParam.Required {
		t.Error("First parameter should be required 'host'")
	}
}

func TestCheckContainerLogsToolMetadata(t *testing.T) {
	tool := NewCheckContainerLogsTool()

	if tool.Name() != "check_container_logs" {
		t.Errorf("Name() = %v, want check_container_logs", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}

	params := tool.Parameters()
	if len(params) != 4 {
		t.Errorf("Parameters() length = %v, want 4", len(params))
	}

	// 验证必需参数
	requiredParams := 0
	for _, p := range params {
		if p.Required {
			requiredParams++
		}
	}
	if requiredParams != 2 {
		t.Errorf("Required parameters = %v, want 2", requiredParams)
	}
}
