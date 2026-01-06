package agent

import (
	"context"
	"fmt"
	"time"

	"ai-ops/internal/llm"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"
	"ai-ops/pkg/logger"

	"go.uber.org/zap"
)

// TroubleshootAgent 问题定位专家 Agent
type TroubleshootAgent struct {
	llmClient    llm.Client
	toolRegistry *tool.Registry
	sshPool      *ssh.Pool
	maxLoops     int
	timeout      time.Duration
}

// NewTroubleshootAgent 创建问题定位专家
func NewTroubleshootAgent(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool) *TroubleshootAgent {
	return &TroubleshootAgent{
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sshPool:      sshPool,
		maxLoops:     10,
		timeout:      5 * time.Minute,
	}
}

func (a *TroubleshootAgent) Type() ExpertType {
	return ExpertTroubleshoot
}

func (a *TroubleshootAgent) Description() string {
	return "问题定位专家：擅长日志分析、错误追踪、根因定位、故障排查"
}

func (a *TroubleshootAgent) GetTools() []string {
	return []string{"query_log", "check_process", "root_cause_diagnosis", "anomaly_detection", "run_command"}
}

func (a *TroubleshootAgent) Execute(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	systemPrompt := a.buildPrompt(req.Hosts)
	messages := []llm.Message{
		llm.NewSystemMessage(systemPrompt),
	}
	messages = append(messages, req.History...)
	messages = append(messages, llm.NewUserMessage(req.Message))

	toolDefs := a.buildToolDefinitions()
	var toolCallRecords []ToolCallRecord
	var finalReply string

	for i := 0; i < a.maxLoops; i++ {
		logger.Debug("TroubleshootAgent 循环", zap.Int("loop", i+1))

		resp, err := a.llmClient.ChatWithTools(ctx, messages, toolDefs)
		if err != nil {
			return nil, fmt.Errorf("LLM 调用失败: %w", err)
		}

		if resp.HasToolCalls() {
			toolResults, records := a.executeToolCalls(ctx, resp.Message.ToolCalls, req.Hosts)
			toolCallRecords = append(toolCallRecords, records...)
			messages = append(messages, resp.Message)
			messages = append(messages, toolResults...)
			continue
		}

		finalReply = resp.Message.Content
		break
	}

	if finalReply == "" && len(toolCallRecords) > 0 {
		finalReply = "已完成问题排查，请查看工具调用结果。"
	}

	return &ChatResponse{
		SessionID: req.SessionID,
		Reply:     finalReply,
		ToolCalls: toolCallRecords,
	}, nil
}

func (a *TroubleshootAgent) buildPrompt(hosts []string) string {
	return fmt.Sprintf(`# 问题定位专家

你是一名专业的故障排查专家，擅长日志分析、错误追踪和根因定位。

## 核心能力
- 日志分析：快速定位错误日志、异常模式
- 进程诊断：检查进程状态、资源占用
- 根因分析：追踪问题根源，提供修复建议
- 异常检测：识别系统异常行为

## 排查流程
1. **收集信息**：查看相关日志、进程状态
2. **分析问题**：识别错误模式、异常指标
3. **定位根因**：追踪问题源头
4. **提供建议**：给出具体的修复方案

## 可用工具
- query_log: 查询系统日志
- check_process: 检查进程状态
- root_cause_diagnosis: 根因诊断
- anomaly_detection: 异常检测
- run_command: 执行诊断命令

## 目标主机
%v

## 注意事项
- 优先查看日志，日志是问题定位的关键
- 关注错误时间线，找出问题发生的顺序
- 给出具体可执行的修复建议
`, hosts)
}

func (a *TroubleshootAgent) buildToolDefinitions() []llm.ToolDef {
	allowedTools := map[string]bool{
		"query_log":            true,
		"check_process":        true,
		"root_cause_diagnosis": true,
		"anomaly_detection":    true,
		"run_command":          true,
	}

	tools := a.toolRegistry.GenerateJSONSchema()
	toolDefs := make([]llm.ToolDef, 0)

	for _, t := range tools {
		function, ok := t["function"].(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := function["name"].(string)
		if !allowedTools[name] {
			continue
		}
		description, _ := function["description"].(string)
		parameters, _ := function["parameters"].(map[string]interface{})

		toolDefs = append(toolDefs, llm.ToolDef{
			Type: "function",
			Function: llm.FunctionDef{
				Name:        name,
				Description: description,
				Parameters:  parameters,
			},
		})
	}
	return toolDefs
}

func (a *TroubleshootAgent) executeToolCalls(ctx context.Context, toolCalls []llm.ToolCall, hosts []string) ([]llm.Message, []ToolCallRecord) {
	var messages []llm.Message
	var records []ToolCallRecord

	toolCtx := &tool.Context{
		Hosts: hosts,
		SSH:   a.sshPool,
	}

	for _, tc := range toolCalls {
		params := parseToolParams(tc.Function.Arguments)
		if len(hosts) > 0 && params["host"] == nil {
			params["host"] = hosts[0]
		}

		result, err := a.toolRegistry.Execute(toolCtx, tc.Function.Name, params)

		record := ToolCallRecord{
			ID:     tc.ID,
			Tool:   tc.Function.Name,
			Params: params,
		}

		if err != nil {
			record.Error = err.Error()
			messages = append(messages, llm.NewToolMessage(tc.ID, tc.Function.Name, fmt.Sprintf("错误: %v", err)))
		} else if result != nil {
			record.Result = formatToolResult(result)
			messages = append(messages, llm.NewToolMessage(tc.ID, tc.Function.Name, record.Result))
		}
		records = append(records, record)
	}
	return messages, records
}
