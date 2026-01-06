package agent

import (
	"context"
	"fmt"
	"sync"

	"ai-ops/internal/llm"
	"ai-ops/internal/llm/prompt"
	"ai-ops/internal/permission"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"
	"ai-ops/internal/tool/custom"
	"ai-ops/pkg/logger"

	"go.uber.org/zap"
)

// EnhancedAgent 增强版 Agent，集成所有安全和功能特性
type EnhancedAgent struct {
	*Agent

	// 安全组件
	permChecker  *permission.Checker
	doomDetector *DoomLoopDetector
	snapshots    *SnapshotManager

	// 功能组件
	compaction   *CompactionEngine
	promptLoader *prompt.PromptLoader
	customTools  *custom.CustomToolLoader

	// 配置
	providerID string
	mu         sync.RWMutex
}

// EnhancedConfig 增强版配置
type EnhancedConfig struct {
	Config // 嵌入基础配置

	// 安全配置
	PermissionConfigPath string // 权限配置文件路径
	DoomLoopWindowSize   int    // Doom Loop 检测窗口
	DoomLoopThreshold    int    // Doom Loop 阈值
	MaxSnapshots         int    // 最大快照数

	// 功能配置
	CompactionThreshold int    // 压缩阈值
	CompactionKeepRecent int   // 保留最近消息数
	PromptDir           string // Prompt 模板目录
	CustomToolDir       string // 自定义工具目录
	ProviderID          string // LLM Provider ID
}

// NewEnhancedAgent 创建增强版 Agent
func NewEnhancedAgent(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool, cfg EnhancedConfig) (*EnhancedAgent, error) {
	// 设置默认值
	if cfg.DoomLoopWindowSize == 0 {
		cfg.DoomLoopWindowSize = 10
	}
	if cfg.DoomLoopThreshold == 0 {
		cfg.DoomLoopThreshold = 3
	}
	if cfg.MaxSnapshots == 0 {
		cfg.MaxSnapshots = 50
	}
	if cfg.CompactionThreshold == 0 {
		cfg.CompactionThreshold = 20
	}
	if cfg.CompactionKeepRecent == 0 {
		cfg.CompactionKeepRecent = 10
	}
	if cfg.PromptDir == "" {
		cfg.PromptDir = "./prompts"
	}
	if cfg.ProviderID == "" {
		cfg.ProviderID = "anthropic"
	}

	// 创建基础 Agent
	baseAgent := NewAgent(llmClient, toolRegistry, sshPool, cfg.Config)

	ea := &EnhancedAgent{
		Agent:        baseAgent,
		doomDetector: NewDoomLoopDetector(cfg.DoomLoopWindowSize, cfg.DoomLoopThreshold),
		snapshots:    NewSnapshotManager(cfg.MaxSnapshots),
		compaction:   NewCompactionEngine(cfg.CompactionThreshold, cfg.CompactionKeepRecent),
		promptLoader: prompt.NewPromptLoader(cfg.PromptDir),
		providerID:   cfg.ProviderID,
	}

	// 加载权限配置（可选）
	if cfg.PermissionConfigPath != "" {
		checker, err := permission.NewChecker(cfg.PermissionConfigPath)
		if err != nil {
			logger.Warn("加载权限配置失败，使用默认策略", zap.Error(err))
		} else {
			ea.permChecker = checker
		}
	}

	// 加载自定义工具（可选）
	if cfg.CustomToolDir != "" {
		ea.customTools = custom.NewCustomToolLoader(cfg.CustomToolDir)
		if tools, err := ea.customTools.Load(); err == nil {
			logger.Info("加载自定义工具", zap.Int("count", len(tools)))
		}
	}

	return ea, nil
}

// Chat 增强版对话（集成所有功能）
func (ea *EnhancedAgent) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 重置 Doom Loop 检测器（新会话）
	if len(req.History) == 0 {
		ea.doomDetector.Reset()
	}

	// 上下文压缩
	messages := ea.buildMessagesWithCompaction(req)

	// 使用 Provider 特定的 Prompt
	systemPrompt, err := ea.getProviderPrompt(req)
	if err != nil {
		logger.Warn("加载 Provider Prompt 失败，使用默认", zap.Error(err))
	}

	if systemPrompt != "" {
		messages[0] = llm.NewSystemMessage(systemPrompt)
	}

	// 获取工具定义
	toolDefs := ea.buildToolDefinitions()

	// Agent 循环
	var toolCallRecords []ToolCallRecord
	var finalReply string

	for i := 0; i < ea.maxLoops; i++ {
		logger.Debug("Enhanced Agent 循环", zap.Int("loop", i+1))

		// 调用 LLM
		resp, err := ea.llmClient.ChatWithTools(ctx, messages, toolDefs)
		if err != nil {
			return nil, fmt.Errorf("LLM 调用失败: %w", err)
		}

		// 检查是否有工具调用
		if resp.HasToolCalls() {
			// 执行工具调用（带安全检查）
			toolResults, records, blocked := ea.executeToolCallsWithSafety(ctx, req.SessionID, resp.Message.ToolCalls, req.Hosts)
			toolCallRecords = append(toolCallRecords, records...)

			if blocked {
				// 有工具被阻止，返回提示
				return &ChatResponse{
					SessionID: req.SessionID,
					Reply:     "部分操作因安全策略被阻止，请确认后重试。",
					ToolCalls: toolCallRecords,
				}, nil
			}

			messages = append(messages, resp.Message)
			messages = append(messages, toolResults...)
			continue
		}

		finalReply = resp.Message.Content
		break
	}

	if finalReply == "" && len(toolCallRecords) > 0 {
		finalReply = "已执行相关操作，请查看工具调用结果。"
	}

	return &ChatResponse{
		SessionID: req.SessionID,
		Reply:     finalReply,
		ToolCalls: toolCallRecords,
	}, nil
}

// buildMessagesWithCompaction 构建消息列表（带压缩）
func (ea *EnhancedAgent) buildMessagesWithCompaction(req ChatRequest) []llm.Message {
	messages := make([]llm.Message, 0, len(req.History)+2)

	// 系统提示词
	systemPrompt := SelectSystemPrompt(ea.promptVersion, ea.toolRegistry, req.Hosts)
	messages = append(messages, llm.NewSystemMessage(systemPrompt))

	// 历史消息（可能需要压缩）
	history := req.History
	if len(history) > ea.compaction.Threshold() {
		// 简化压缩：保留最近 N 条消息
		keepRecent := ea.compaction.KeepRecent()
		if len(history) > keepRecent {
			// 生成摘要
			compactCount := len(history) - keepRecent
			summary := fmt.Sprintf("[Earlier conversation: %d messages compressed]", compactCount)
			messages = append(messages, llm.NewSystemMessage(summary))
			history = history[len(history)-keepRecent:]
			logger.Info("上下文压缩", zap.Int("original", len(req.History)), zap.Int("kept", keepRecent))
		}
	}
	messages = append(messages, history...)

	// 用户消息
	messages = append(messages, llm.NewUserMessage(req.Message))

	return messages
}

// getProviderPrompt 获取 Provider 特定的 Prompt
func (ea *EnhancedAgent) getProviderPrompt(req ChatRequest) (string, error) {
	if ea.promptLoader == nil {
		return "", nil
	}

	data := map[string]any{
		"Tools":       ea.toolRegistry.GeneratePrompt(),
		"Context":     fmt.Sprintf("Hosts: %v", req.Hosts),
		"UserMessage": req.Message,
	}

	return ea.promptLoader.Render(ea.providerID, data)
}

// executeToolCallsWithSafety 执行工具调用（带安全检查）
func (ea *EnhancedAgent) executeToolCallsWithSafety(ctx context.Context, sessionID string, toolCalls []llm.ToolCall, hosts []string) ([]llm.Message, []ToolCallRecord, bool) {
	var toolMessages []llm.Message
	var records []ToolCallRecord
	blocked := false

	for _, tc := range toolCalls {
		// 1. 权限检查
		if ea.permChecker != nil {
			result := ea.permChecker.Check(tc.Function.Name)
			switch result.Action {
			case permission.ActionDeny:
				logger.Warn("工具被拒绝", zap.String("tool", tc.Function.Name))
				records = append(records, ToolCallRecord{
					ID:    tc.ID,
					Tool:  tc.Function.Name,
					Error: fmt.Sprintf("权限拒绝: %s", result.Message),
				})
				toolMessages = append(toolMessages, llm.NewToolMessage(tc.ID, tc.Function.Name, "权限拒绝"))
				blocked = true
				continue
			case permission.ActionAsk:
				logger.Info("工具需要确认", zap.String("tool", tc.Function.Name), zap.String("message", result.Message))
				// 这里可以添加用户确认逻辑，暂时允许执行
			}
		}

		// 2. Doom Loop 检测
		ea.doomDetector.Record(tc.Function.Name, nil) // 简化：不传参数
		if isLoop, reason := ea.doomDetector.Check(); isLoop {
			logger.Warn("检测到 Doom Loop", zap.String("reason", reason))
			records = append(records, ToolCallRecord{
				ID:    tc.ID,
				Tool:  tc.Function.Name,
				Error: fmt.Sprintf("检测到死循环: %s", reason),
			})
			toolMessages = append(toolMessages, llm.NewToolMessage(tc.ID, tc.Function.Name, "检测到死循环，已中止"))
			blocked = true
			continue
		}

		// 3. 创建快照（危险操作）
		if isDangerousTool(tc.Function.Name) {
			_, err := ea.snapshots.Create(sessionID, tc.Function.Name, nil, nil)
			if err != nil {
				logger.Warn("创建快照失败", zap.Error(err))
			}
		}

		// 4. 执行工具
		msgs, recs := ea.parallelExecutor.ExecuteParallel(ctx, []llm.ToolCall{tc}, hosts)
		toolMessages = append(toolMessages, msgs...)
		records = append(records, recs...)
	}

	return toolMessages, records, blocked
}

// isDangerousTool 判断是否为危险工具
func isDangerousTool(name string) bool {
	dangerous := map[string]bool{
		"exec_command": true,
		"run_script":   true,
		"deploy":       true,
		"restart":      true,
		"stop":         true,
		"kill":         true,
	}
	return dangerous[name]
}

// GetSnapshots 获取会话快照列表
func (ea *EnhancedAgent) GetSnapshots(sessionID string) []*Snapshot {
	return ea.snapshots.List(sessionID)
}

// ResetDoomDetector 重置 Doom Loop 检测器
func (ea *EnhancedAgent) ResetDoomDetector() {
	ea.doomDetector.Reset()
}
