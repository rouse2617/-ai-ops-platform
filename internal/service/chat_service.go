package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai-ops/internal/agent"
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
	agent       *agent.Agent
	router      *agent.AgentRouter // 智能路由器
	sessionRepo repository.SessionRepository
}

// NewChatService 创建聊天服务
func NewChatService(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool, sessionRepo repository.SessionRepository) *ChatService {
	// 创建 Agent
	agentInstance := agent.NewAgent(llmClient, toolRegistry, sshPool, agent.Config{
		MaxLoops:      10,
		Timeout:       5 * time.Minute,
		PromptVersion: "enhanced",
		CacheTTL:      30 * time.Second,
		MaxConcurrent: 10,
	})

	// 创建智能路由器
	router := agent.NewAgentRouter(llmClient, toolRegistry, sshPool, agentInstance)

	// 注册专家 Agent
	router.RegisterExpert(agent.NewTroubleshootAgent(llmClient, toolRegistry, sshPool))
	router.RegisterExpert(agent.NewMonitorAgent(llmClient, toolRegistry, sshPool))
	router.RegisterExpert(agent.NewDatabaseAgent(llmClient, toolRegistry, sshPool))

	return &ChatService{
		agent:       agentInstance,
		router:      router,
		sessionRepo: sessionRepo,
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
				hostID := s.getCurrentHostID(req.Hosts)
				assistantMsg := &model.Message{
					ID:        uuid.New().String(),
					SessionID: req.SessionID,
					HostID:    hostID,
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

// callAgentService 使用 Agent 执行对话
func (s *ChatService) callAgentService(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 从数据库加载历史消息
	history := s.loadHistoryFromDB(req.SessionID)

	// 如果前端也传了历史，合并（优先使用数据库的）
	if len(history) == 0 && len(req.History) > 0 {
		for _, msg := range req.History {
			history = append(history, llm.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 调用智能路由器（自动选择专家 Agent）
	agentResp, err := s.router.Route(ctx, agent.ChatRequest{
		SessionID: req.SessionID,
		Message:   req.Message,
		Hosts:     req.Hosts,
		History:   history,
	})
	if err != nil {
		return nil, err
	}

	// 转换响应
	toolCalls := make([]ToolCallRecord, 0, len(agentResp.ToolCalls))
	for _, tc := range agentResp.ToolCalls {
		toolCalls = append(toolCalls, ToolCallRecord{
			ID:     tc.ID,
			Tool:   tc.Tool,
			Params: tc.Params,
			Result: tc.Result,
			Error:  tc.Error,
		})
	}

	return &ChatResponse{
		SessionID: agentResp.SessionID,
		Reply:     agentResp.Reply,
		ToolCalls: toolCalls,
		Thinking:  agentResp.Thinking,
	}, nil
}

// callAgentServiceStream 使用 Agent 执行流式对话
func (s *ChatService) callAgentServiceStream(ctx context.Context, req ChatRequest, callback func(chunk StreamChunk)) error {
	// 从数据库加载历史消息
	history := s.loadHistoryFromDB(req.SessionID)

	// 如果前端也传了历史，合并（优先使用数据库的）
	if len(history) == 0 && len(req.History) > 0 {
		for _, msg := range req.History {
			history = append(history, llm.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 调用 Agent 流式接口
	return s.agent.ChatStream(ctx, agent.ChatRequest{
		SessionID: req.SessionID,
		Message:   req.Message,
		Hosts:     req.Hosts,
		History:   history,
	}, func(chunk agent.StreamChunk) {
		// 转换 StreamChunk
		serviceChunk := StreamChunk{
			Type:    chunk.Type,
			Content: chunk.Content,
			Error:   chunk.Error,
			Step:    chunk.Step,
			Status:  chunk.Status,
		}

		if chunk.ToolCall != nil {
			serviceChunk.ToolCall = &ToolCallRecord{
				ID:     chunk.ToolCall.ID,
				Tool:   chunk.ToolCall.Function.Name,
				Params: s.parseToolParams(chunk.ToolCall.Function.Arguments),
			}
		}

		if chunk.ToolResult != nil {
			serviceChunk.ToolResult = &ToolCallRecord{
				ID:     chunk.ToolResult.ID,
				Tool:   chunk.ToolResult.Tool,
				Params: chunk.ToolResult.Params,
				Result: chunk.ToolResult.Result,
				Error:  chunk.ToolResult.Error,
			}
		}

		callback(serviceChunk)
	})
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

	hostID := s.getCurrentHostID(req.Hosts)
	userMsg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		HostID:    hostID,
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

	hostID := s.getCurrentHostID(req.Hosts)
	assistantMsg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		HostID:    hostID,
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

// UpdateSessionHosts 更新会话关联的主机列表
func (s *ChatService) UpdateSessionHosts(sessionID string, hosts []string) error {
	return s.sessionRepo.UpdateHosts(sessionID, hosts)
}

// loadHistoryFromDB 从数据库加载历史消息
func (s *ChatService) loadHistoryFromDB(sessionID string) []llm.Message {
	if sessionID == "" {
		return nil
	}

	messages, err := s.sessionRepo.GetMessages(sessionID)
	if err != nil {
		return nil
	}

	// 只加载纯文本消息（user 和 assistant），跳过工具调用相关消息
	history := make([]llm.Message, 0, len(messages))
	for _, msg := range messages {
		// 只保留有内容的 user 和 assistant 消息
		if msg.Content == "" {
			continue
		}
		role := string(msg.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		history = append(history, llm.Message{
			Role:    role,
			Content: msg.Content,
		})
	}
	return history
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

// getCurrentHostID 获取当前活动的主机 ID
func (s *ChatService) getCurrentHostID(hosts []string) string {
	if len(hosts) == 0 {
		return ""
	}
	return hosts[0]
}
