package retriever

import (
	"fmt"
	"strings"
)

// ContextBuilder 上下文构建器
type ContextBuilder struct {
	maxTokens int
}

// NewContextBuilder 创建上下文构建器
func NewContextBuilder(maxTokens int) *ContextBuilder {
	return &ContextBuilder{
		maxTokens: maxTokens,
	}
}

// SearchResult 检索结果
type SearchResult struct {
	ChunkID    string                 `json:"chunk_id"`
	DocumentID string                 `json:"document_id"`
	Content    string                 `json:"content"`
	Score      float32                `json:"score"`
	Metadata   map[string]interface{} `json:"metadata"`
	Highlights []string               `json:"highlights"`
}

// BuildContext 构建上下文
func (cb *ContextBuilder) BuildContext(results []SearchResult) string {
	var context strings.Builder
	tokenCount := 0

	// 按类别分组
	grouped := cb.groupByCategory(results)

	// 按优先级添加内容
	priorities := []string{"incident", "sop", "alert_rule", "manual"}
	categoryTitles := map[string]string{
		"incident":   "历史故障案例",
		"sop":        "SOP 处理流程",
		"alert_rule": "相关告警规则",
		"manual":     "运维手册",
	}

	context.WriteString("# 相关知识库内容\n\n")

	for _, category := range priorities {
		if items, ok := grouped[category]; ok {
			section := cb.buildSection(categoryTitles[category], items, cb.maxTokens-tokenCount)
			context.WriteString(section)
			tokenCount += cb.estimateTokens(section)

			if tokenCount >= cb.maxTokens {
				break
			}
		}
	}

	return context.String()
}

// groupByCategory 按类别分组
func (cb *ContextBuilder) groupByCategory(results []SearchResult) map[string][]SearchResult {
	grouped := make(map[string][]SearchResult)

	for _, result := range results {
		if category, ok := result.Metadata["category"].(string); ok {
			grouped[category] = append(grouped[category], result)
		} else {
			grouped["manual"] = append(grouped["manual"], result)
		}
	}

	return grouped
}

// buildSection 构建分类章节
func (cb *ContextBuilder) buildSection(title string, results []SearchResult, remainingTokens int) string {
	var section strings.Builder
	section.WriteString(fmt.Sprintf("## %s\n\n", title))

	for i, result := range results {
		content := cb.formatResult(i+1, result)
		tokens := cb.estimateTokens(content)

		if tokens > remainingTokens {
			break
		}

		section.WriteString(content)
		section.WriteString("\n\n")
		remainingTokens -= tokens
	}

	return section.String()
}

// formatResult 格式化单个结果
func (cb *ContextBuilder) formatResult(index int, result SearchResult) string {
	var formatted strings.Builder

	formatted.WriteString(fmt.Sprintf("### 案例 %d (相似度: %.2f)\n\n", index, result.Score))
	formatted.WriteString(result.Content)

	// 添加元数据
	if service, ok := result.Metadata["service"].(string); ok {
		formatted.WriteString(fmt.Sprintf("\n\n**服务**: %s", service))
	}
	if severity, ok := result.Metadata["severity"].(string); ok {
		formatted.WriteString(fmt.Sprintf(" | **严重级别**: %s", severity))
	}

	return formatted.String()
}

// estimateTokens 估算 token 数量
func (cb *ContextBuilder) estimateTokens(text string) int {
	// 简单估算：1 token ≈ 4 字符
	return len(text) / 4
}
