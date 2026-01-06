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

// DatabaseAgent 数据库专家 Agent
type DatabaseAgent struct {
	llmClient    llm.Client
	toolRegistry *tool.Registry
	sshPool      *ssh.Pool
	maxLoops     int
	timeout      time.Duration
}

// NewDatabaseAgent 创建数据库专家
func NewDatabaseAgent(llmClient llm.Client, toolRegistry *tool.Registry, sshPool *ssh.Pool) *DatabaseAgent {
	return &DatabaseAgent{
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sshPool:      sshPool,
		maxLoops:     10,
		timeout:      5 * time.Minute,
	}
}

func (a *DatabaseAgent) Type() ExpertType {
	return ExpertDatabase
}

func (a *DatabaseAgent) Description() string {
	return "数据库专家：擅长 MySQL/Redis/PostgreSQL 诊断、连接池分析、慢查询优化、数据库备份"
}

func (a *DatabaseAgent) GetTools() []string {
	return []string{"check_mysql_status", "check_redis_status", "backup_database", "run_command"}
}

func (a *DatabaseAgent) Execute(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
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
		logger.Debug("DatabaseAgent 循环", zap.Int("loop", i+1))

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
		finalReply = "已完成数据库诊断，请查看结果。"
	}

	return &ChatResponse{
		SessionID: req.SessionID,
		Reply:     finalReply,
		ToolCalls: toolCallRecords,
	}, nil
}

func (a *DatabaseAgent) buildPrompt(hosts []string) string {
	return fmt.Sprintf(`# 数据库专家

你是一名专业的数据库专家，擅长 MySQL、Redis、PostgreSQL 等数据库的诊断和优化。

## 核心能力
- MySQL 诊断：连接数、慢查询、锁等待、缓存命中率
- Redis 分析：内存使用、键空间、持久化状态
- PostgreSQL 优化：查询计划、索引建议、连接池
- 数据库备份：全量备份、增量备份、恢复验证

## 诊断流程
1. **状态检查**：获取数据库运行状态和关键指标
2. **问题识别**：分析连接数、慢查询、锁等待等问题
3. **性能分析**：评估缓存命中率、内存使用等
4. **优化建议**：提供具体的优化方案

## 可用工具
- check_mysql_status: 检查 MySQL 状态
- check_redis_status: 检查 Redis 状态
- backup_database: 数据库备份
- run_command: 执行数据库诊断命令

## 目标主机
%v

## 常用诊断命令
- MySQL 连接数: SHOW STATUS LIKE 'Threads_connected'
- MySQL 慢查询: SHOW VARIABLES LIKE 'slow_query_log'
- Redis 内存: redis-cli INFO memory
- Redis 键数量: redis-cli DBSIZE

## 注意事项
- 优先使用专用工具，避免直接执行危险 SQL
- 备份操作前确认磁盘空间充足
- 给出具体的优化参数建议
`, hosts)
}

func (a *DatabaseAgent) buildToolDefinitions() []llm.ToolDef {
	allowedTools := map[string]bool{
		"check_mysql_status": true,
		"check_redis_status": true,
		"backup_database":    true,
		"run_command":        true,
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

func (a *DatabaseAgent) executeToolCalls(ctx context.Context, toolCalls []llm.ToolCall, hosts []string) ([]llm.Message, []ToolCallRecord) {
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
