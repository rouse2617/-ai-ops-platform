package builtin

import (
	"testing"

	"ai-ops/internal/tool"
)

func TestManageServiceTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "缺少 host 参数",
			params: map[string]interface{}{
				"service": "nginx",
				"action":  "start",
			},
			expectError: true,
			errorMsg:    "参数 host 不能为空",
		},
		{
			name: "缺少 service 参数",
			params: map[string]interface{}{
				"host":   "web1",
				"action": "start",
			},
			expectError: true,
			errorMsg:    "参数 service 不能为空",
		},
		{
			name: "缺少 action 参数",
			params: map[string]interface{}{
				"host":    "web1",
				"service": "nginx",
			},
			expectError: true,
			errorMsg:    "参数 action 不能为空",
		},
		{
			name: "SSH 上下文未初始化",
			params: map[string]interface{}{
				"host":    "web1",
				"service": "nginx",
				"action":  "start",
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
	}

	toolInstance := NewManageServiceTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toolInstance.Execute(nil, tt.params)
			if err != nil {
				t.Fatalf("Execute 返回错误: %v", err)
			}

			if tt.expectError {
				if result.Success {
					t.Errorf("期望失败但成功了")
				}
				if result.Error != tt.errorMsg {
					t.Errorf("错误消息不匹配，期望: %s, 实际: %s", tt.errorMsg, result.Error)
				}
			}
		})
	}
}

func TestCheckServiceHealthTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "缺少 host 参数",
			params: map[string]interface{}{
				"target":   "localhost",
				"port":     80,
				"protocol": "http",
			},
			expectError: true,
			errorMsg:    "参数 host 不能为空",
		},
		{
			name: "缺少 target 参数",
			params: map[string]interface{}{
				"host":     "web1",
				"port":     80,
				"protocol": "http",
			},
			expectError: true,
			errorMsg:    "参数 target 不能为空",
		},
		{
			name: "port 参数无效",
			params: map[string]interface{}{
				"host":     "web1",
				"target":   "localhost",
				"port":     0,
				"protocol": "http",
			},
			expectError: true,
			errorMsg:    "参数 port 必须大于 0",
		},
		{
			name: "缺少 protocol 参数",
			params: map[string]interface{}{
				"host":   "web1",
				"target": "localhost",
				"port":   80,
			},
			expectError: true,
			errorMsg:    "参数 protocol 不能为空",
		},
		{
			name: "不支持的协议",
			params: map[string]interface{}{
				"host":     "web1",
				"target":   "localhost",
				"port":     80,
				"protocol": "ftp",
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
		{
			name: "SSH 上下文未初始化",
			params: map[string]interface{}{
				"host":     "web1",
				"target":   "localhost",
				"port":     80,
				"protocol": "http",
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
	}

	toolInstance := NewCheckServiceHealthTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toolInstance.Execute(nil, tt.params)
			if err != nil {
				t.Fatalf("Execute 返回错误: %v", err)
			}

			if tt.expectError {
				if result.Success {
					t.Errorf("期望失败但成功了")
				}
				if result.Error != tt.errorMsg {
					t.Errorf("错误消息不匹配，期望: %s, 实际: %s", tt.errorMsg, result.Error)
				}
			}
		})
	}
}

func TestReloadNginxTool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "缺少 host 参数",
			params:      map[string]interface{}{},
			expectError: true,
			errorMsg:    "参数 host 不能为空",
		},
		{
			name: "SSH 上下文未初始化",
			params: map[string]interface{}{
				"host": "web1",
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
		{
			name: "仅验证模式",
			params: map[string]interface{}{
				"host":          "web1",
				"validate_only": true,
			},
			expectError: true,
			errorMsg:    "SSH 上下文未初始化",
		},
	}

	toolInstance := NewReloadNginxTool()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toolInstance.Execute(nil, tt.params)
			if err != nil {
				t.Fatalf("Execute 返回错误: %v", err)
			}

			if tt.expectError {
				if result.Success {
					t.Errorf("期望失败但成功了")
				}
				if result.Error != tt.errorMsg {
					t.Errorf("错误消息不匹配，期望: %s, 实际: %s", tt.errorMsg, result.Error)
				}
			}
		})
	}
}

func TestServiceToolsMetadata(t *testing.T) {
	tests := []struct {
		name         string
		tool         tool.Tool
		expectedName string
		minParams    int
	}{
		{
			name:         "ManageServiceTool",
			tool:         NewManageServiceTool(),
			expectedName: "manage_service",
			minParams:    3,
		},
		{
			name:         "CheckServiceHealthTool",
			tool:         NewCheckServiceHealthTool(),
			expectedName: "check_service_health",
			minParams:    4,
		},
		{
			name:         "ReloadNginxTool",
			tool:         NewReloadNginxTool(),
			expectedName: "reload_nginx",
			minParams:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tool.Name() != tt.expectedName {
				t.Errorf("工具名称不匹配，期望: %s, 实际: %s", tt.expectedName, tt.tool.Name())
			}

			if tt.tool.Description() == "" {
				t.Error("工具描述不能为空")
			}

			params := tt.tool.Parameters()
			if len(params) < tt.minParams {
				t.Errorf("参数数量不足，期望至少: %d, 实际: %d", tt.minParams, len(params))
			}

			for _, param := range params {
				if param.Name == "" {
					t.Error("参数名称不能为空")
				}
				if param.Type == "" {
					t.Error("参数类型不能为空")
				}
				if param.Description == "" {
					t.Error("参数描述不能为空")
				}
			}
		})
	}
}
