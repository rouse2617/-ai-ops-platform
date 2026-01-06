package custom

import "ai-ops/internal/tool"

// CustomTool 自定义工具定义
type CustomTool struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Parameters  []tool.Parameter    `yaml:"parameters"`
	Script      string              `yaml:"script"`
}

// ToolDefinition YAML 文件格式
type ToolDefinition struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Parameters  []ParamDef  `yaml:"parameters"`
	Script      string      `yaml:"script"`
}

// ParamDef 参数定义（用于 YAML 解析）
type ParamDef struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Required    bool        `yaml:"required"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default,omitempty"`
}
