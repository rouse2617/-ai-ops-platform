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

// MonitorAgent 监控查询专家 Agent
type MonitorAgent struct {
	llmClient    llm.Client
	toolRegistry *tool.Registry
	sshPool      *ssh.Pool
	maxLoops     int
	timeout      time.Duration
}

// NewMonitorAgent 创建监控查询专家
func NewMonitorAgent(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool) *MonitorAgent {
	return &MonitorAgent{
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sshPool:      sshPool,
		maxLoops:     10,
		timeout:      5 * time.Minute,
	}
}

func (a *MonitorAgent) Type() ExpertType {
	return ExpertMonitor
}

func (a *MonitorAgent) Description() string {
	return "监控查询专家：擅长 CPU/内存/磁盘监控、指标查询、趋势分析、告警管理"
}

func (a *MonitorAgent) GetTools() []string {
	return []string{"check_cpu", "check_memory", "check_disk", "trend_analysis", "alert_manager", "performance_analysis"}
}

func (a *MonitorAgent) Execute(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
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
		logger.Debug("MonitorAgent 循环", zap.Int("loop", i+1))

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
		finalReply = "已完成监控数据查询，请查看结果。"
	}

	return &ChatResponse{
		SessionID: req.SessionID,
		Reply:     finalReply,
		ToolCalls: toolCallRecords,
	}, nil
}

func (a *MonitorAgent) buildPrompt(hosts []string) string {
	return fmt.Sprintf(`# 监控查询专家

你是一名专业的系统监控专家，擅长指标查询、趋势分析和告警管理。

## 核心能力
- 指标查询：CPU、内存、磁盘、网络等系统指标
- 趋势分析：分析指标变化趋势，预测潜在问题
- 告警管理：查看和管理系统告警
- 性能分析：综合分析系统性能状况

## 分析流程
1. **收集指标**：获取相关系统指标数据
2. **趋势分析**：分析指标变化趋势
3. **异常识别**：识别异常指标和潜在风险
4. **建议优化**：提供性能优化建议

## 可用工具
- check_cpu: 检查 CPU 使用率
- check_memory: 检查内存使用情况
- check_disk: 检查磁盘使用情况
- trend_analysis: 趋势分析
- alert_manager: 告警管理
- performance_analysis: 性能分析

## 目标主机
%v

## 注意事项
- 提供具体的数值和百分比
- 对比历史数据，分析趋势
- 给出明确的健康状态评估
`, hosts)
}

func (a *MonitorAgent) buildToolDefinitions() []llm.ToolDef {
	allowedTools := map[string]bool{
		"check_cpu":            true,
		"check_memory":         true,
		"check_disk":           true,
		"trend_analysis":       true,
		"alert_manager":        true,
		"performance_analysis": true,
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

func (a *MonitorAgent) executeToolCalls(ctx context.Context, toolCalls []llm.ToolCall, hosts []string) ([]llm.Message, []ToolCallRecord) {
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
