package retriever

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-ops/internal/llm"
)

// QueryRewriter 查询改写器
type QueryRewriter struct {
	llmClient llm.Client
}

// NewQueryRewriter 创建查询改写器
func NewQueryRewriter(llmClient llm.Client) *QueryRewriter {
	return &QueryRewriter{
		llmClient: llmClient,
	}
}

// Rewrite 改写查询
func (qr *QueryRewriter) Rewrite(ctx context.Context, query string, context map[string]interface{}) ([]string, error) {
	prompt := fmt.Sprintf(`你是运维知识库检索专家。请将以下查询改写为更适合检索的形式：

原始查询: %s
上下文: %v

要求：
1. 添加相关同义词和术语
2. 如果是复杂查询，拆分为多个子查询
3. 保留关键技术术语
4. 返回 1-3 个改写后的查询

输出格式（JSON）：
{
  "queries": ["改写查询1", "改写查询2"]
}`, query, context)

	resp, err := qr.llmClient.Chat(ctx, []llm.Message{
		llm.NewUserMessage(prompt),
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Queries []string `json:"queries"`
	}

	if err := json.Unmarshal([]byte(resp.Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result.Queries, nil
}
