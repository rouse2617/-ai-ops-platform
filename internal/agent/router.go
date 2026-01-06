package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-ops/internal/llm"
	"ai-ops/internal/ssh"
	"ai-ops/internal/tool"
	"ai-ops/pkg/logger"

	"go.uber.org/zap"
)

// AgentRouter 智能路由器，根据用户意图选择合适的专家 Agent
type AgentRouter struct {
	experts      map[ExpertType]ExpertAgent
	llmClient    llm.Client
	toolRegistry *tool.Registry
	sshPool      *ssh.Pool
	defaultAgent *Agent // 通用 Agent 作为兜底
}

// NewAgentRouter 创建智能路由器
func NewAgentRouter(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool, defaultAgent *Agent) *AgentRouter {
	return &AgentRouter{
		experts:      make(map[ExpertType]ExpertAgent),
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sshPool:      sshPool,
		defaultAgent: defaultAgent,
	}
}

// RegisterExpert 注册专家 Agent
func (r *AgentRouter) RegisterExpert(expert ExpertAgent) {
	r.experts[expert.Type()] = expert
	logger.Info("注册专家 Agent", zap.String("type", string(expert.Type())))
}

// RouteResult 路由结果
type RouteResult struct {
	Expert     ExpertType `json:"expert"`
	Confidence float64    `json:"confidence"`
	Reason     string     `json:"reason"`
}

// Route 智能路由：分析用户意图，选择合适的专家 Agent
func (r *AgentRouter) Route(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 1. 使用 LLM 分析意图
	routeResult, err := r.analyzeIntent(ctx, req.Message)
	if err != nil {
		logger.Warn("意图分析失败，使用默认 Agent", zap.Error(err))
		return r.defaultAgent.Chat(ctx, req)
	}

	logger.Info("路由决策",
		zap.String("expert", string(routeResult.Expert)),
		zap.Float64("confidence", routeResult.Confidence),
		zap.String("reason", routeResult.Reason),
	)

	// 2. 如果置信度高且有对应专家，使用专家 Agent
	if routeResult.Confidence >= 0.7 {
		if expert, ok := r.experts[routeResult.Expert]; ok {
			resp, err := expert.Execute(ctx, req)
			if err == nil {
				// 在响应中添加路由信息
				resp.Thinking = fmt.Sprintf("[路由到 %s 专家，置信度 %.0f%%]\n%s",
					routeResult.Expert, routeResult.Confidence*100, routeResult.Reason)
			}
			return resp, err
		}
	}

	// 3. 兜底：使用通用 Agent
	return r.defaultAgent.Chat(ctx, req)
}

// analyzeIntent 使用 LLM 分析用户意图
func (r *AgentRouter) analyzeIntent(ctx context.Context, message string) (*RouteResult, error) {
	// 构建专家信息
	expertInfos := GetAllExpertInfo()
	var expertDescriptions strings.Builder
	for _, info := range expertInfos {
		expertDescriptions.WriteString(fmt.Sprintf("- %s: %s\n", info.Type, info.Description))
	}

	prompt := fmt.Sprintf(`你是一个智能路由器，负责分析用户问题并选择最合适的专家来处理。

## 可用专家
%s

## 用户问题
%s

## 任务
分析用户问题，选择最合适的专家类型。返回 JSON 格式：
{"expert": "专家类型", "confidence": 0.0-1.0, "reason": "选择原因"}

注意：
- confidence 表示匹配程度，0.7 以上才会使用专家
- 如果问题不明确或涉及多个领域，选择 general
- 只返回 JSON，不要其他内容`, expertDescriptions.String(), message)

	resp, err := r.llmClient.Chat(ctx, []llm.Message{
		llm.NewUserMessage(prompt),
	})
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 解析 JSON 响应
	content := strings.TrimSpace(resp.Message.Content)
	// 处理可能的 markdown 代码块
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result RouteResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析路由结果失败: %w, content: %s", err, content)
	}

	return &result, nil
}

// GetRegisteredExperts 获取已注册的专家列表
func (r *AgentRouter) GetRegisteredExperts() []ExpertType {
	experts := make([]ExpertType, 0, len(r.experts))
	for t := range r.experts {
		experts = append(experts, t)
	}
	return experts
}
