package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicClient Anthropic Claude 客户端
type AnthropicClient struct {
	client    anthropic.Client
	model     string
	maxTokens int64
}

// AnthropicConfig 客户端配置
type AnthropicConfig struct {
	APIKey    string
	BaseURL   string
	Model     string
	MaxTokens int
}

// NewAnthropicClient 创建 Anthropic 客户端
func NewAnthropicClient(cfg AnthropicConfig) *AnthropicClient {
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-5-20250929"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}

	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	return &AnthropicClient{
		client:    anthropic.NewClient(opts...),
		model:     cfg.Model,
		maxTokens: int64(cfg.MaxTokens),
	}
}

// Chat 普通对话
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	return c.ChatWithTools(ctx, messages, nil)
}

// ChatWithTools 带工具调用的对话
func (c *AnthropicClient) ChatWithTools(ctx context.Context, messages []Message, tools []ToolDef) (*ChatResponse, error) {
	system, anthropicMsgs := c.convertMessages(messages)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens,
		Messages:  anthropicMsgs,
	}

	if system != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: system},
		}
	}

	if len(tools) > 0 {
		params.Tools = c.convertTools(tools)
	}

	resp, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Anthropic API 调用失败: %w", err)
	}

	return c.convertResponse(resp), nil
}

// ChatStream 流式对话
func (c *AnthropicClient) ChatStream(ctx context.Context, messages []Message, callback StreamCallback) error {
	return c.ChatStreamWithTools(ctx, messages, nil, callback)
}

// ChatStreamWithTools 带工具调用的流式对话
func (c *AnthropicClient) ChatStreamWithTools(ctx context.Context, messages []Message, tools []ToolDef, callback StreamCallback) error {
	system, anthropicMsgs := c.convertMessages(messages)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens,
		Messages:  anthropicMsgs,
	}

	if system != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: system},
		}
	}

	if len(tools) > 0 {
		params.Tools = c.convertTools(tools)
	}

	stream := c.client.Messages.NewStreaming(ctx, params)

	var currentToolUse *struct {
		id    string
		name  string
		input string
	}

	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		_ = message.Accumulate(event)

		switch eventVariant := event.AsAny().(type) {
		case anthropic.ContentBlockStartEvent:
			block := eventVariant.ContentBlock
			if block.Type == "tool_use" {
				currentToolUse = &struct {
					id    string
					name  string
					input string
				}{
					id:   block.ID,
					name: block.Name,
				}
			}

		case anthropic.ContentBlockDeltaEvent:
			switch deltaVariant := eventVariant.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				callback(StreamChunk{
					Type:    "content",
					Content: deltaVariant.Text,
				})
			case anthropic.InputJSONDelta:
				if currentToolUse != nil {
					currentToolUse.input += deltaVariant.PartialJSON
				}
			}

		case anthropic.ContentBlockStopEvent:
			if currentToolUse != nil {
				callback(StreamChunk{
					Type: "tool_call",
					ToolCall: &ToolCall{
						ID:   currentToolUse.id,
						Type: "function",
						Function: FunctionCall{
							Name:      currentToolUse.name,
							Arguments: currentToolUse.input,
						},
					},
				})
				currentToolUse = nil
			}

		case anthropic.MessageDeltaEvent:
			if eventVariant.Delta.StopReason != "" {
				callback(StreamChunk{
					Type:         "done",
					FinishReason: c.convertStopReason(string(eventVariant.Delta.StopReason)),
				})
			}
		}
	}

	if err := stream.Err(); err != nil {
		callback(StreamChunk{Type: "error", Error: err.Error()})
		return fmt.Errorf("流式调用失败: %w", err)
	}

	return nil
}

// convertMessages 转换消息格式，提取 system 消息
func (c *AnthropicClient) convertMessages(messages []Message) (string, []anthropic.MessageParam) {
	var system string
	var anthropicMsgs []anthropic.MessageParam

	for _, msg := range messages {
		if msg.Role == RoleSystem {
			system = msg.Content
			continue
		}

		// 转换为 Anthropic 格式
		if msg.Role == RoleTool {
			// tool 消息转换为 user 消息，包含 tool_result
			anthropicMsgs = append(anthropicMsgs, anthropic.NewUserMessage(
				anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false),
			))
		} else if len(msg.ToolCalls) > 0 {
			// assistant 消息包含 tool_use
			var content []anthropic.ContentBlockParamUnion
			if msg.Content != "" {
				content = append(content, anthropic.NewTextBlock(msg.Content))
			}
			for _, tc := range msg.ToolCalls {
				// 解析 Arguments JSON
				var input map[string]interface{}
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
				if input == nil {
					input = map[string]interface{}{}
				}
				content = append(content, anthropic.ContentBlockParamUnion{
					OfToolUse: &anthropic.ToolUseBlockParam{
						ID:    tc.ID,
						Name:  tc.Function.Name,
						Input: input,
					},
				})
			}
			anthropicMsgs = append(anthropicMsgs, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleAssistant,
				Content: content,
			})
		} else {
			// 普通文本消息
			role := anthropic.MessageParamRoleUser
			if msg.Role == RoleAssistant {
				role = anthropic.MessageParamRoleAssistant
			}
			anthropicMsgs = append(anthropicMsgs, anthropic.MessageParam{
				Role:    role,
				Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(msg.Content)},
			})
		}
	}

	return system, anthropicMsgs
}

// convertTools 转换工具定义
func (c *AnthropicClient) convertTools(tools []ToolDef) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(tools))
	for i, tool := range tools {
		toolParam := anthropic.ToolParam{
			Name:        tool.Function.Name,
			Description: anthropic.String(tool.Function.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: tool.Function.Parameters,
			},
		}
		result[i] = anthropic.ToolUnionParam{OfTool: &toolParam}
	}
	return result
}

// convertResponse 转换响应格式
func (c *AnthropicClient) convertResponse(resp *anthropic.Message) *ChatResponse {
	msg := Message{Role: string(resp.Role)}

	for _, block := range resp.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			msg.Content += variant.Text
		case anthropic.ToolUseBlock:
			inputJSON, _ := json.Marshal(variant.Input)
			msg.ToolCalls = append(msg.ToolCalls, ToolCall{
				ID:   variant.ID,
				Type: "function",
				Function: FunctionCall{
					Name:      variant.Name,
					Arguments: string(inputJSON),
				},
			})
		}
	}

	return &ChatResponse{
		ID:      resp.ID,
		Model:   string(resp.Model),
		Message: msg,
		Usage: Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
		FinishReason: c.convertStopReason(string(resp.StopReason)),
	}
}

// convertStopReason 转换结束原因
func (c *AnthropicClient) convertStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return FinishReasonStop
	case "tool_use":
		return FinishReasonToolCalls
	case "max_tokens":
		return FinishReasonLength
	default:
		return reason
	}
}

// GetModel 获取当前模型
func (c *AnthropicClient) GetModel() string {
	return c.model
}

// SetModel 设置模型
func (c *AnthropicClient) SetModel(model string) {
	c.model = model
}
