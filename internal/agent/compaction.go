package agent

import (
	"fmt"
	"strings"

	"ai-ops/internal/model"
)

// CompactionEngine 上下文压缩引擎
type CompactionEngine struct {
	threshold  int // 触发压缩的消息数阈值
	keepRecent int // 保留最近 N 条消息
}

// CompactionResult 压缩结果
type CompactionResult struct {
	OriginalCount  int
	CompactedCount int
	Summary        string
}

// NewCompactionEngine 创建压缩引擎
func NewCompactionEngine(threshold, keepRecent int) *CompactionEngine {
	return &CompactionEngine{
		threshold:  threshold,
		keepRecent: keepRecent,
	}
}

// ShouldCompact 判断是否需要压缩
func (c *CompactionEngine) ShouldCompact(messages []model.Message) bool {
	return len(messages) > c.threshold
}

// Compact 执行压缩
func (c *CompactionEngine) Compact(messages []model.Message) (*CompactionResult, []model.Message) {
	originalCount := len(messages)

	if len(messages) <= c.keepRecent+1 {
		return &CompactionResult{
			OriginalCount:  originalCount,
			CompactedCount: len(messages),
			Summary:        "no compaction needed",
		}, messages
	}

	var systemMsg *model.Message
	var compactMessages []model.Message
	var recentMessages []model.Message

	// 分离 system prompt
	if len(messages) > 0 && messages[0].Role == model.RoleSystem {
		systemMsg = &messages[0]
		messages = messages[1:]
	}

	// 分离最近的消息
	if len(messages) > c.keepRecent {
		compactMessages = messages[:len(messages)-c.keepRecent]
		recentMessages = messages[len(messages)-c.keepRecent:]
	} else {
		recentMessages = messages
	}

	// 构建结果
	result := []model.Message{}
	if systemMsg != nil {
		result = append(result, *systemMsg)
	}

	// 合并早期消息为摘要
	if len(compactMessages) > 0 {
		summary := c.buildSummary(compactMessages)
		result = append(result, model.Message{
			Role:    model.RoleSystem,
			Content: summary,
		})
	}

	// 添加最近消息
	result = append(result, recentMessages...)

	return &CompactionResult{
		OriginalCount:  originalCount,
		CompactedCount: len(result),
		Summary:        fmt.Sprintf("compressed %d messages into summary", len(compactMessages)),
	}, result
}

// buildSummary 构建消息摘要
func (c *CompactionEngine) buildSummary(messages []model.Message) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[Earlier conversation summary: %d messages]", len(messages)))

	for _, msg := range messages {
		preview := msg.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		parts = append(parts, fmt.Sprintf("- %s: %s", msg.Role, preview))
	}

	return strings.Join(parts, "\n")
}
