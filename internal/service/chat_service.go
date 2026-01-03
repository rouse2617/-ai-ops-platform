package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai-ops/internal/llm"
	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ChatService 聊天服务
type ChatService struct {
	llmClient    llm.Client
	toolRegistry *tool.Registry
	sshPool      *ssh.Pool
	sessionRepo  repository.SessionRepository
	maxLoops     int
	timeout      time.Duration
}

// NewChatService 创建聊天服务
func NewChatService(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool, sessionRepo repository.SessionRepository) *ChatService {
	return &ChatService{
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sshPool:      sshPool,
		sessionRepo:  sessionRepo,
		maxLoops:     10,
		timeout:      5 * time.Minute,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string
	Message   string
	Hosts     []string
	Stream    bool
	History   []ChatMessage
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	SessionID string
	Reply     string
	ToolCalls []ToolCallRecord
	Thinking  string
}

// ToolCallRecord 工具调用记录
type ToolCallRecord struct {
	ID     string                 `json:"id,omitempty"`
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
	Result interface{}            `json:"result,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// Chat 执行聊天
func (s *ChatService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	req = s.normalizeRequest(req)

	if err := s.ensureSessionExists(req); err != nil {
		return nil, fmt.Errorf("确保会话存在失败: %w", err)
	}

	if err := s.saveUserMessage(req); err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	agentResp, err := s.callAgentService(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("调用 Agent 服务失败: %w", err)
	}

	if err := s.saveAssistantMessage(req, agentResp); err != nil {
		return nil, fmt.Errorf("保存助手消息失败: %w", err)
	}

	s.updateSessionTitleIfNeeded(req)

	return agentResp, nil
}

// ChatStream 流式聊天
func (s *ChatService) ChatStream(ctx context.Context, req ChatRequest, callback func(chunk StreamChunk)) error {
	req = s.normalizeRequest(req)

	if err := s.ensureSessionExists(req); err != nil {
		return fmt.Errorf("确保会话存在失败: %w", err)
	}

	if err := s.saveUserMessage(req); err != nil {
		return fmt.Errorf("保存用户消息失败: %w", err)
	}

	var assistantContent strings.Builder
	var toolCalls []model.ToolCall

	streamCallback := func(chunk StreamChunk) {
		switch chunk.Type {
		case "content", "text":
			assistantContent.WriteString(chunk.Content)
		case "tool_result":
			if chunk.ToolResult != nil {
				toolCalls = append(toolCalls, model.ToolCall{
					Tool:   chunk.ToolResult.Tool,
					Params: chunk.ToolResult.Params,
					Result: chunk.ToolResult.Result,
					Error:  chunk.ToolResult.Error,
				})
			}
		case "done":
			if assistantContent.Len() > 0 {
				assistantMsg := &model.Message{
					ID:        uuid.New().String(),
					SessionID: req.SessionID,
					Role:      model.RoleAssistant,
					Content:   assistantContent.String(),
					ToolCalls: toolCalls,
					CreatedAt: time.Now(),
				}
				if err := s.sessionRepo.AddMessage(req.SessionID, assistantMsg); err != nil {
					zap.L().Error("保存助手消息失败", zap.Error(err))
				}
			}
			s.updateSessionTitleIfNeeded(req)
		}
		callback(chunk)
	}

	return s.callAgentServiceStream(ctx, req, streamCallback)
}

// StreamChunk 流式数据块
type StreamChunk struct {
	Type       string          `json:"type"`
	Content    string          `json:"content,omitempty"`
	ToolCall   *ToolCallRecord `json:"tool_call,omitempty"`
	ToolResult *ToolCallRecord `json:"tool_result,omitempty"`
	Error      string          `json:"error,omitempty"`
	Step       int             `json:"step,omitempty"`
	Status     string          `json:"status,omitempty"`
}

// callAgentService 使用 llmClient 和 ToolRegistry 执行对话
func (s *ChatService) callAgentService(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 构建消息列表
	messages := s.buildMessages(req)

	// 获取工具定义
	toolDefs := s.buildToolDefinitions()

	// Agent 循环
	var toolCallRecords []ToolCallRecord
	var finalReply string

	for i := 0; i < s.maxLoops; i++ {
		// 调用 LLM
		resp, err := s.llmClient.ChatWithTools(ctx, messages, toolDefs)
		if err != nil {
			return nil, fmt.Errorf("LLM 调用失败: %w", err)
		}

		// 检查是否有工具调用
		if resp.HasToolCalls() {
			// 执行工具调用
			toolResults, records := s.executeToolCalls(ctx, resp.Message.ToolCalls, req.Hosts)
			toolCallRecords = append(toolCallRecords, records...)

			// 添加助手消息（包含工具调用）
			messages = append(messages, resp.Message)

			// 添加工具结果消息
			messages = append(messages, toolResults...)
			continue
		}

		// 没有工具调用，返回最终回复
		finalReply = resp.Message.Content
		break
	}

	// 如果达到最大循环次数但没有最终回复
	if finalReply == "" && len(toolCallRecords) > 0 {
		finalReply = "已执行相关操作，请查看工具调用结果。"
	}

	return &ChatResponse{
		SessionID: req.SessionID,
		Reply:     finalReply,
		ToolCalls: toolCallRecords,
	}, nil
}

// callAgentServiceStream 使用 llmClient 和 ToolRegistry 执行流式对话
func (s *ChatService) callAgentServiceStream(ctx context.Context, req ChatRequest, callback func(chunk StreamChunk)) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 构建消息列表
	messages := s.buildMessages(req)

	// 获取工具定义
	toolDefs := s.buildToolDefinitions()

	// Agent 循环
	for i := 0; i < s.maxLoops; i++ {
		var contentBuffer string
		var toolCalls []llm.ToolCall
		done := false

		// 流式调用 LLM
		err := s.llmClient.ChatStreamWithTools(ctx, messages, toolDefs, func(chunk llm.StreamChunk) {
			switch chunk.Type {
			case "content":
				contentBuffer += chunk.Content
				callback(StreamChunk{
					Type:    "content",
					Content: chunk.Content,
				})
			case "tool_call":
				if chunk.ToolCall != nil {
					toolCalls = append(toolCalls, *chunk.ToolCall)
				}
			case "done":
				done = true
			}
		})

		if err != nil {
			return fmt.Errorf("LLM 流式调用失败: %w", err)
		}

		// 检查是否有工具调用
		if len(toolCalls) > 0 {
			// 通知前端工具调用
			for _, tc := range toolCalls {
				callback(StreamChunk{
					Type: "tool_call",
					ToolCall: &ToolCallRecord{
						ID:     tc.ID,
						Tool:   tc.Function.Name,
						Params: s.parseToolParams(tc.Function.Arguments),
					},
				})
			}

			// 执行工具
			toolResults, records := s.executeToolCalls(ctx, toolCalls, req.Hosts)

			// 通知前端工具结果
			for _, r := range records {
				callback(StreamChunk{
					Type:       "tool_result",
					ToolResult: &r,
				})
			}

			// 添加消息继续对话
			messages = append(messages, llm.NewAssistantToolCallMessage(toolCalls))
			messages = append(messages, toolResults...)
			continue
		}

		// 没有工具调用，对话结束
		if done {
			callback(StreamChunk{Type: "done"})
			break
		}
	}

	return nil
}

// normalizeRequest 标准化请求
func (s *ChatService) normalizeRequest(req ChatRequest) ChatRequest {
	req.Message = strings.TrimSpace(req.Message)
	return req
}

// ensureSessionExists 确保会话存在
func (s *ChatService) ensureSessionExists(req ChatRequest) error {
	if req.SessionID == "" {
		return nil
	}

	_, err := s.sessionRepo.GetByID(req.SessionID)
	if err == nil {
		return nil
	}

	session := &model.Session{
		ID:        req.SessionID,
		Title:     s.generateSessionTitle(req.Message),
		Hosts:     req.Hosts,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.sessionRepo.Create(session)
}

// generateSessionTitle 生成会话标题
func (s *ChatService) generateSessionTitle(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 50 {
		return message[:50] + "..."
	}
	if message == "" {
		return fmt.Sprintf("会话 %s", time.Now().Format("2006-01-02 15:04:05"))
	}
	return message
}

// saveUserMessage 保存用户消息
func (s *ChatService) saveUserMessage(req ChatRequest) error {
	if req.SessionID == "" {
		return nil
	}

	userMsg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		Role:      model.RoleUser,
		Content:   req.Message,
		CreatedAt: time.Now(),
	}
	return s.sessionRepo.AddMessage(req.SessionID, userMsg)
}

// saveAssistantMessage 保存助手消息
func (s *ChatService) saveAssistantMessage(req ChatRequest, resp *ChatResponse) error {
	if req.SessionID == "" {
		return nil
	}

	toolCalls := make([]model.ToolCall, 0, len(resp.ToolCalls))
	for _, tc := range resp.ToolCalls {
		toolCalls = append(toolCalls, model.ToolCall{
			Tool:   tc.Tool,
			Params: tc.Params,
			Result: tc.Result,
			Error:  tc.Error,
		})
	}

	assistantMsg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		Role:      model.RoleAssistant,
		Content:   resp.Reply,
		ToolCalls: toolCalls,
		CreatedAt: time.Now(),
	}
	return s.sessionRepo.AddMessage(req.SessionID, assistantMsg)
}

// updateSessionTitleIfNeeded 更新会话标题（如果需要）
func (s *ChatService) updateSessionTitleIfNeeded(req ChatRequest) {
	if req.SessionID == "" {
		return
	}

	session, err := s.sessionRepo.GetByID(req.SessionID)
	if err != nil {
		return
	}

	if session.Title == "" || strings.HasPrefix(session.Title, "会话 ") {
		newTitle := s.generateSessionTitle(req.Message)
		if newTitle != "" {
			_ = s.sessionRepo.UpdateTitle(req.SessionID, newTitle)
		}
	}
}

// GetSessions 获取会话列表
func (s *ChatService) GetSessions(limit int) ([]*model.Session, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.sessionRepo.List(limit)
}

// CreateSession 创建新会话
func (s *ChatService) CreateSession(title string, hosts []string) (string, error) {
	sessionID := uuid.New().String()

	if title == "" {
		title = fmt.Sprintf("会话 %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	session := &model.Session{
		ID:        sessionID,
		Title:     title,
		Hosts:     hosts,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return "", err
	}

	return sessionID, nil
}

// GetHistory 获取对话历史
func (s *ChatService) GetHistory(sessionID string) ([]*model.Message, error) {
	_, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("会话不存在: %w", err)
	}

	return s.sessionRepo.GetMessages(sessionID)
}

// DeleteSession 删除会话
func (s *ChatService) DeleteSession(sessionID string) error {
	return s.sessionRepo.Delete(sessionID)
}

// buildMessages 构建消息列表
func (s *ChatService) buildMessages(req ChatRequest) []llm.Message {
	messages := make([]llm.Message, 0, len(req.History)+2)

	// 系统提示词
	systemPrompt := s.buildSystemPrompt(req.Hosts)
	messages = append(messages, llm.NewSystemMessage(systemPrompt))

	// 历史消息 - 转换类型
	for _, msg := range req.History {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 用户消息
	messages = append(messages, llm.NewUserMessage(req.Message))

	return messages
}

// buildSystemPrompt 构建系统提示词
func (s *ChatService) buildSystemPrompt(hosts []string) string {
	prompt := "你是一个专业的 AI 运维助手，可以帮助用户管理和操作服务器。\n\n"

	if len(hosts) > 0 {
		prompt += fmt.Sprintf("当前关联的主机: %s\n\n", strings.Join(hosts, ", "))
	}

	prompt += "你可以使用提供的工具来执行各种运维任务。请根据用户的需求选择合适的工具。\n"
	prompt += "在执行命令前，请确保理解用户的意图，必要时向用户确认。"

	return prompt
}

// buildToolDefinitions 构建工具定义
func (s *ChatService) buildToolDefinitions() []llm.ToolDef {
	if s.toolRegistry == nil {
		return nil
	}

	schemas := s.toolRegistry.GenerateJSONSchema()
	toolDefs := make([]llm.ToolDef, 0, len(schemas))

	for _, schema := range schemas {
		function, ok := schema["function"].(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := function["name"].(string)
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

// executeToolCalls 执行工具调用
func (s *ChatService) executeToolCalls(ctx context.Context, toolCalls []llm.ToolCall, hosts []string) ([]llm.Message, []ToolCallRecord) {
	toolMessages := make([]llm.Message, 0, len(toolCalls))
	records := make([]ToolCallRecord, 0, len(toolCalls))

	for _, tc := range toolCalls {
		toolName := tc.Function.Name
		params := s.parseToolParams(tc.Function.Arguments)

		// 执行工具
		toolCtx := &tool.Context{
			Hosts:   hosts,
			SSH:     s.sshPool,
			Timeout: 30 * time.Second,
		}

		result, err := s.toolRegistry.Execute(toolCtx, toolName, params)

		var resultStr string
		var errorStr string

		if err != nil {
			errorStr = err.Error()
			resultStr = fmt.Sprintf("工具执行失败: %s", err.Error())
		} else if result != nil {
			if result.Success {
				resultStr = result.Message
				if result.Data != nil {
					dataJSON, _ := json.Marshal(result.Data)
					resultStr = string(dataJSON)
				}
			} else {
				errorStr = result.Error
				resultStr = fmt.Sprintf("工具执行失败: %s", result.Error)
			}
		}

		// 添加工具结果消息
		toolMessages = append(toolMessages, llm.NewToolMessage(tc.ID, toolName, resultStr))

		// 记录工具调用
		records = append(records, ToolCallRecord{
			ID:     tc.ID,
			Tool:   toolName,
			Params: params,
			Result: resultStr,
			Error:  errorStr,
		})
	}

	return toolMessages, records
}

// parseToolParams 解析工具参数
func (s *ChatService) parseToolParams(arguments string) map[string]interface{} {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		zap.L().Error("解析工具参数失败", zap.Error(err), zap.String("arguments", arguments))
		return make(map[string]interface{})
	}
	return params
}
