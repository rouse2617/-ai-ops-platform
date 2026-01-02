package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-ops/internal/model"
	"ai-ops/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ChatService 聊天服务
type ChatService struct {
	agentServiceURL string
	sessionRepo     repository.SessionRepository
	httpClient      *http.Client
}

// NewChatService 创建聊天服务
func NewChatService(agentServiceURL string, sessionRepo repository.SessionRepository) *ChatService {
	return &ChatService{
		agentServiceURL: agentServiceURL,
		sessionRepo:     sessionRepo,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
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

// callAgentService 调用 Node.js Agent Service
func (s *ChatService) callAgentService(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	reqBody := map[string]interface{}{
		"sessionId": req.SessionID,
		"message":   req.Message,
		"hosts":     req.Hosts,
	}
	if len(req.History) > 0 {
		reqBody["history"] = req.History
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.agentServiceURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Agent 服务返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var result struct {
		SessionID string           `json:"sessionId"`
		Reply     string           `json:"reply"`
		Response  string           `json:"response"`
		ToolCalls []ToolCallRecord `json:"toolCalls"`
		Thinking  string           `json:"thinking"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 兼容 response 和 reply 两种字段名
	reply := result.Reply
	if reply == "" {
		reply = result.Response
	}

	return &ChatResponse{
		SessionID: result.SessionID,
		Reply:     reply,
		ToolCalls: result.ToolCalls,
		Thinking:  result.Thinking,
	}, nil
}

// callAgentServiceStream 调用 Node.js Agent Service 流式接口
func (s *ChatService) callAgentServiceStream(ctx context.Context, req ChatRequest, callback func(chunk StreamChunk)) error {
	reqBody := map[string]interface{}{
		"sessionId": req.SessionID,
		"message":   req.Message,
		"hosts":     req.Hosts,
		"stream":    true,
	}
	if len(req.History) > 0 {
		reqBody["history"] = req.History
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.agentServiceURL+"/chat/stream", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Agent 服务返回错误: %d - %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if bytes.HasPrefix(line, []byte("data: ")) {
			data := bytes.TrimPrefix(line, []byte("data: "))
			var chunk StreamChunk
			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}
			callback(chunk)
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
