package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ExpertConfig 专家 Agent 配置
type ExpertConfig struct {
	Name        string   `yaml:"name"`        // 专家名称（唯一标识）
	Description string   `yaml:"description"` // 专家描述（用于路由匹配）
	Tools       []string `yaml:"tools"`       // 允许使用的工具列表
	Prompt      string   `yaml:"prompt"`      // 系统提示词（可以是内联或文件路径）
	PromptFile  string   `yaml:"prompt_file"` // 提示词文件路径
	MaxLoops    int      `yaml:"max_loops"`   // 最大循环次数
	Timeout     int      `yaml:"timeout"`     // 超时时间（秒）
}

// SkillConfig Skill 技能配置
type SkillConfig struct {
	Name        string   `yaml:"name"`        // 技能名称（唯一标识）
	Description string   `yaml:"description"` // 技能描述
	Prompt      string   `yaml:"prompt"`      // 执行提示词
	Expert      string   `yaml:"expert"`      // 指定使用的专家（可选）
	Tools       []string `yaml:"tools"`       // 允许使用的工具（可选）
}

// ExpertsConfig 专家和技能配置文件结构
type ExpertsConfig struct {
	Experts []ExpertConfig `yaml:"experts"` // 专家列表
	Skills  []SkillConfig  `yaml:"skills"`  // 技能列表
}

// LoadExpertsConfig 从文件加载专家和技能配置
func LoadExpertsConfig(path string) (*ExpertsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := &ExpertsConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	for i := range cfg.Experts {
		if cfg.Experts[i].MaxLoops == 0 {
			cfg.Experts[i].MaxLoops = 10
		}
		if cfg.Experts[i].Timeout == 0 {
			cfg.Experts[i].Timeout = 300 // 5 分钟
		}
	}

	// 加载提示词文件
	baseDir := filepath.Dir(path)
	for i := range cfg.Experts {
		if cfg.Experts[i].PromptFile != "" && cfg.Experts[i].Prompt == "" {
			promptPath := cfg.Experts[i].PromptFile
			if !filepath.IsAbs(promptPath) {
				promptPath = filepath.Join(baseDir, promptPath)
			}
			promptData, err := os.ReadFile(promptPath)
			if err != nil {
				return nil, fmt.Errorf("读取专家 %s 的提示词文件失败: %w", cfg.Experts[i].Name, err)
			}
			cfg.Experts[i].Prompt = string(promptData)
		}
	}

	return cfg, nil
}

// LoadExpertsConfigFromDir 从目录加载所有专家配置（支持多文件）
func LoadExpertsConfigFromDir(dir string) (*ExpertsConfig, error) {
	cfg := &ExpertsConfig{
		Experts: make([]ExpertConfig, 0),
		Skills:  make([]SkillConfig, 0),
	}

	// 查找所有 yaml 文件
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("查找配置文件失败: %w", err)
	}

	ymlFiles, _ := filepath.Glob(filepath.Join(dir, "*.yml"))
	files = append(files, ymlFiles...)

	for _, file := range files {
		fileCfg, err := LoadExpertsConfig(file)
		if err != nil {
			return nil, fmt.Errorf("加载配置文件 %s 失败: %w", file, err)
		}
		cfg.Experts = append(cfg.Experts, fileCfg.Experts...)
		cfg.Skills = append(cfg.Skills, fileCfg.Skills...)
	}

	// 也支持加载 Markdown 格式的 Skill 文件（类似 Claude Code）
	mdFiles, _ := filepath.Glob(filepath.Join(dir, "skills", "*.md"))
	for _, file := range mdFiles {
		skill, err := LoadSkillFromMarkdown(file)
		if err != nil {
			continue // 跳过无效的 skill 文件
		}
		cfg.Skills = append(cfg.Skills, *skill)
	}

	return cfg, nil
}

// LoadSkillFromMarkdown 从 Markdown 文件加载 Skill（类似 Claude Code 格式）
func LoadSkillFromMarkdown(path string) (*SkillConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// 解析 YAML 前置元数据
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("无效的 skill 文件格式：缺少 YAML 前置元数据")
	}

	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("无效的 skill 文件格式：YAML 前置元数据未闭合")
	}

	// 解析 YAML 元数据
	skill := &SkillConfig{}
	if err := yaml.Unmarshal([]byte(parts[0]), skill); err != nil {
		return nil, fmt.Errorf("解析 skill 元数据失败: %w", err)
	}

	// 使用 Markdown 内容作为提示词
	skill.Prompt = strings.TrimSpace(parts[1])

	return skill, nil
}
