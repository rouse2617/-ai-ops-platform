package permission

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChecker_Check(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_permissions.yaml")

	configContent := `default: ask
rules:
  - pattern: "check_*"
    action: allow
  - pattern: "exec_command"
    action: ask
    message: "执行远程命令需要确认"
  - pattern: "rm_*"
    action: deny
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	checker, err := NewChecker(configPath)
	if err != nil {
		t.Fatalf("NewChecker() error = %v", err)
	}

	tests := []struct {
		name        string
		toolName    string
		wantAction  Action
		wantMessage string
	}{
		{
			name:       "allow check tool",
			toolName:   "check_disk",
			wantAction: ActionAllow,
		},
		{
			name:        "ask exec command",
			toolName:    "exec_command",
			wantAction:  ActionAsk,
			wantMessage: "执行远程命令需要确认",
		},
		{
			name:       "deny rm tool",
			toolName:   "rm_file",
			wantAction: ActionDeny,
		},
		{
			name:       "default ask",
			toolName:   "unknown_tool",
			wantAction: ActionAsk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.Check(tt.toolName)
			if result.Action != tt.wantAction {
				t.Errorf("Check(%q).Action = %v; want %v", tt.toolName, result.Action, tt.wantAction)
			}
			if tt.wantMessage != "" && result.Message != tt.wantMessage {
				t.Errorf("Check(%q).Message = %q; want %q", tt.toolName, result.Message, tt.wantMessage)
			}
		})
	}
}

func TestNewChecker_InvalidConfig(t *testing.T) {
	_, err := NewChecker("nonexistent.yaml")
	if err == nil {
		t.Error("NewChecker() with invalid path should return error")
	}
}
