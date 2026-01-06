package agent

import (
	"context"
	"fmt"
	"strings"

	"ai-ops/internal/llm"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"
)

// SkillExecutor Skill 技能执行器
type SkillExecutor struct {
	skills       map[string]SkillConfig
	experts      map[string]ExpertAgent
	defaultAgent *Agent
	llmClient    llm.Client
	toolRegistry *tool.Registry
	sshPool      *ssh.Pool
}

// NewSkillExecutor 创建 Skill 执行器
func NewSkillExecutor(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool, defaultAgent *Agent) *SkillExecutor {
	return &SkillExecutor{
		skills:       make(map[string]SkillConfig),
		experts:      make(map[string]ExpertAgent),
		defaultAgent: defaultAgent,
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sshPool:      sshPool,
	}
}

// RegisterSkill 注册 Skill
func (e *SkillExecutor) RegisterSkill(skill SkillConfig) {
	e.skills[skill.Name] = skill
}

// RegisterExpert 注册专家（用于 Skill 指定专家时使用）
func (e *SkillExecutor) RegisterExpert(expert ExpertAgent) {
	e.experts[string(expert.Type())] = expert
}

// GetSkills 获取所有 Skill 列表
func (e *SkillExecutor) GetSkills() []SkillConfig {
	skills := make([]SkillConfig, 0, len(e.skills))
	for _, s := range e.skills {
		skills = append(skills, s)
	}
	return skills
}

// GetSkill 获取指定 Skill
func (e *SkillExecutor) GetSkill(name string) (SkillConfig, bool) {
	skill, ok := e.skills[name]
	return skill, ok
}

// IsSkillCommand 检查消息是否是 Skill 命令（以 / 开头）
func (e *SkillExecutor) IsSkillCommand(message string) bool {
	message = strings.TrimSpace(message)
	if !strings.HasPrefix(message, "/") {
		return false
	}
	// 提取命令名
	parts := strings.SplitN(message[1:], " ", 2)
	if len(parts) == 0 {
		return false
	}
	_, ok := e.skills[parts[0]]
	return ok
}

// ParseSkillCommand 解析 Skill 命令
func (e *SkillExecutor) ParseSkillCommand(message string) (skillName string, args string, ok bool) {
	message = strings.TrimSpace(message)
	if !strings.HasPrefix(message, "/") {
		return "", "", false
	}
	parts := strings.SplitN(message[1:], " ", 2)
	if len(parts) == 0 {
		return "", "", false
	}
	skillName = parts[0]
	if _, exists := e.skills[skillName]; !exists {
		return "", "", false
	}
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}
	return skillName, args, true
}

// Execute 执行 Skill
func (e *SkillExecutor) Execute(ctx context.Context, skillName string, args string, req ChatRequest) (*ChatResponse, error) {
	skill, ok := e.skills[skillName]
	if !ok {
		return nil, fmt.Errorf("skill 不存在: %s", skillName)
	}

	// 构建完整的提示词
	fullPrompt := skill.Prompt
	if args != "" {
		fullPrompt = fmt.Sprintf("%s\n\n用户参数: %s", skill.Prompt, args)
	}

	// 创建新的请求
	skillReq := ChatRequest{
		SessionID: req.SessionID,
		Message:   fullPrompt,
		Hosts:     req.Hosts,
		History:   req.History,
	}

	// 如果指定了专家，使用专家执行
	if skill.Expert != "" {
		if expert, exists := e.experts[skill.Expert]; exists {
			return expert.Execute(ctx, skillReq)
		}
	}

	// 如果指定了工具列表，创建临时专家
	if len(skill.Tools) > 0 {
		tempExpert := NewDynamicExpert(ExpertConfig{
			Name:        skillName,
			Description: skill.Description,
			Tools:       skill.Tools,
			Prompt:      skill.Prompt,
			MaxLoops:    10,
			Timeout:     300,
		}, e.llmClient, e.toolRegistry, e.sshPool)
		return tempExpert.Execute(ctx, skillReq)
	}

	// 使用默认 Agent 执行
	return e.defaultAgent.Chat(ctx, skillReq)
}

// ExecuteWithMessage 执行 Skill（从消息中解析）
func (e *SkillExecutor) ExecuteWithMessage(ctx context.Context, message string, req ChatRequest) (*ChatResponse, error) {
	skillName, args, ok := e.ParseSkillCommand(message)
	if !ok {
		return nil, fmt.Errorf("无效的 skill 命令: %s", message)
	}
	return e.Execute(ctx, skillName, args, req)
}
