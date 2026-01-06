package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-ops/internal/llm"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"
	"ai-ops/pkg/logger"

	"go.uber.org/zap"
)

// DynamicExpert 动态配置的专家 Agent
type DynamicExpert struct {
	config       ExpertConfig
	llmClient    llm.Client
	toolRegistry *tool.Registry
	sshPool      *ssh.Pool
}

// NewDynamicExpert 根据配置创建动态专家
func NewDynamicExpert(cfg ExpertConfig, llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool) *DynamicExpert {
	return &DynamicExpert{
		config:       cfg,
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sshPool:      sshPool,
	}
}

func (e *DynamicExpert) Type() ExpertType {
	return ExpertType(e.config.Name)
}

func (e *DynamicExpert) Description() string {
	return e.config.Description
}

func (e *DynamicExpert) GetTools() []string {
	return e.config.Tools
}

func (e *DynamicExpert) Execute(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	timeout := time.Duration(e.config.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	systemPrompt := e.buildPrompt(req.Hosts)
	messages := []llm.Message{
		llm.NewSystemMessage(systemPrompt),
	}
	messages = append(messages, req.History...)
	messages = append(messages, llm.NewUserMessage(req.Message))

	toolDefs := e.buildToolDefinitions()
	var toolCallRecords []ToolCallRecord
	var finalReply string

	for i := 0; i < e.config.MaxLoops; i++ {
		logger.Debug("DynamicExpert 循环", zap.String("expert", e.config.Name), zap.Int("loop", i+1))

		resp, err := e.llmClient.ChatWithTools(ctx, messages, toolDefs)
		if err != nil {
			return nil, fmt.Errorf("LLM 调用失败: %w", err)
		}

		if resp.HasToolCalls() {
			toolResults, records := e.executeToolCalls(ctx, resp.Message.ToolCalls, req.Hosts)
			toolCallRecords = append(toolCallRecords, records...)
			messages = append(messages, resp.Message)
			messages = append(messages, toolResults...)
			continue
		}

		finalReply = resp.Message.Content
		break
	}

	if finalReply == "" && len(toolCallRecords) > 0 {
		finalReply = fmt.Sprintf("已完成 %s 任务，请查看结果。", e.config.Name)
	}

	return &ChatResponse{
		SessionID: req.SessionID,
		Reply:     finalReply,
		ToolCalls: toolCallRecords,
	}, nil
}

func (e *DynamicExpert) buildPrompt(hosts []string) string {
	// 如果配置了自定义提示词，使用它
	if e.config.Prompt != "" {
		// 替换变量
		prompt := e.config.Prompt
		prompt = strings.ReplaceAll(prompt, "{{hosts}}", fmt.Sprintf("%v", hosts))
		prompt = strings.ReplaceAll(prompt, "{{tools}}", strings.Join(e.config.Tools, ", "))
		return prompt
	}

	// 默认提示词模板
	return fmt.Sprintf(`# %s

%s

## 可用工具
%s

## 目标主机
%v

## 注意事项
- 优先使用专用工具完成任务
- 给出具体可执行的建议
- 如果遇到问题，说明原因并提供替代方案
`, e.config.Name, e.config.Description, strings.Join(e.config.Tools, "\n- "), hosts)
}

func (e *DynamicExpert) buildToolDefinitions() []llm.ToolDef {
	allowedTools := make(map[string]bool)
	for _, t := range e.config.Tools {
		allowedTools[t] = true
	}

	tools := e.toolRegistry.GenerateJSONSchema()
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

func (e *DynamicExpert) executeToolCalls(ctx context.Context, toolCalls []llm.ToolCall, hosts []string) ([]llm.Message, []ToolCallRecord) {
	var messages []llm.Message
	var records []ToolCallRecord

	toolCtx := &tool.Context{
		Hosts: hosts,
		SSH:   e.sshPool,
	}

	for _, tc := range toolCalls {
		params := parseToolParams(tc.Function.Arguments)
		if len(hosts) > 0 && params["host"] == nil {
			params["host"] = hosts[0]
		}

		result, err := e.toolRegistry.Execute(toolCtx, tc.Function.Name, params)

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
