package custom

import (
	"fmt"
	"os"
	"path/filepath"

	"ai-ops/internal/tool"

	"gopkg.in/yaml.v3"
)

// CustomToolLoader 自定义工具加载器
type CustomToolLoader struct {
	toolDir string
}

// NewCustomToolLoader 创建加载器
func NewCustomToolLoader(toolDir string) *CustomToolLoader {
	if toolDir == "" {
		home, _ := os.UserHomeDir()
		toolDir = filepath.Join(home, ".ai-ops", "tools")
	}
	return &CustomToolLoader{toolDir: toolDir}
}

// Load 加载所有工具
func (l *CustomToolLoader) Load() ([]*CustomTool, error) {
	if _, err := os.Stat(l.toolDir); os.IsNotExist(err) {
		return nil, nil
	}

	var tools []*CustomTool
	err := filepath.Walk(l.toolDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		t, err := l.LoadFile(path)
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		tools = append(tools, t)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk tool dir: %w", err)
	}
	return tools, nil
}

// LoadFile 加载单个工具文件
func (l *CustomToolLoader) LoadFile(path string) (*CustomTool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var def ToolDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if def.Name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if def.Script == "" {
		return nil, fmt.Errorf("tool script is required")
	}

	params := make([]tool.Parameter, len(def.Parameters))
	for i, p := range def.Parameters {
		params[i] = tool.Parameter{
			Name:        p.Name,
			Type:        p.Type,
			Required:    p.Required,
			Description: p.Description,
			Default:     p.Default,
		}
	}

	return &CustomTool{
		Name:        def.Name,
		Description: def.Description,
		Parameters:  params,
		Script:      def.Script,
	}, nil
}
