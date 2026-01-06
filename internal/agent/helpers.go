package agent

import (
	"encoding/json"
	"fmt"

	"ai-ops/internal/tool"
)

// parseToolParams 解析工具参数
func parseToolParams(arguments string) map[string]interface{} {
	params := make(map[string]interface{})
	if arguments == "" {
		return params
	}
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return params
	}
	return params
}

// formatToolResult 格式化工具结果
func formatToolResult(result *tool.Result) string {
	if result == nil {
		return ""
	}

	if !result.Success {
		if result.Error != "" {
			return fmt.Sprintf("错误: %s", result.Error)
		}
		return "执行失败"
	}

	// 尝试 JSON 格式化
	if result.Data != nil {
		if jsonBytes, err := json.MarshalIndent(result.Data, "", "  "); err == nil {
			return string(jsonBytes)
		}
	}

	if result.Message != "" {
		return result.Message
	}

	return "执行成功"
}
