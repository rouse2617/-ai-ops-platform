package permission_test

import (
	"fmt"
	"os"
	"path/filepath"

	"ai-ops/internal/permission"
)

// 示例：在工具执行前检查权限
func ExampleChecker_Check() {
	// 创建临时配置
	tmpDir, _ := os.MkdirTemp("", "perm-test")
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "permissions.yaml")
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
	os.WriteFile(configPath, []byte(configContent), 0644)

	// 初始化权限检查器
	checker, _ := permission.NewChecker(configPath)

	// 检查工具权限
	toolName := "check_disk"
	result := checker.Check(toolName)

	switch result.Action {
	case permission.ActionAllow:
		fmt.Printf("工具 %s: 自动允许执行\n", toolName)
	case permission.ActionDeny:
		fmt.Printf("工具 %s: 拒绝执行\n", toolName)
	case permission.ActionAsk:
		fmt.Printf("工具 %s: 需要用户确认", toolName)
		if result.Message != "" {
			fmt.Printf(" - %s", result.Message)
		}
		fmt.Println()
	}

	// Output:
	// 工具 check_disk: 自动允许执行
}

// 示例：批量检查多个工具
func ExampleChecker_Check_batch() {
	tmpDir, _ := os.MkdirTemp("", "perm-test")
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "permissions.yaml")
	configContent := `default: ask
rules:
  - pattern: "check_*"
    action: allow
  - pattern: "exec_command"
    action: ask
  - pattern: "rm_*"
    action: deny
`
	os.WriteFile(configPath, []byte(configContent), 0644)

	checker, _ := permission.NewChecker(configPath)

	tools := []string{"check_disk", "exec_command", "rm_file", "unknown_tool"}

	for _, tool := range tools {
		result := checker.Check(tool)
		fmt.Printf("%s: %s\n", tool, result.Action)
	}

	// Output:
	// check_disk: allow
	// exec_command: ask
	// rm_file: deny
	// unknown_tool: ask
}

// 示例：模式匹配
func ExampleMatch() {
	patterns := []string{"check_*", "exec_command", "rm_*"}
	toolName := "check_disk"

	for _, pattern := range patterns {
		if permission.Match(pattern, toolName) {
			fmt.Printf("工具 %s 匹配模式 %s\n", toolName, pattern)
		}
	}

	// Output:
	// 工具 check_disk 匹配模式 check_*
}
