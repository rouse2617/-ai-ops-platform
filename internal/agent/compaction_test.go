package agent

import (
	"strings"
	"testing"

	"ai-ops/internal/model"
)

func TestNewCompactionEngine(t *testing.T) {
	engine := NewCompactionEngine(20, 10)
	if engine.threshold != 20 {
		t.Errorf("threshold = %d; want 20", engine.threshold)
	}
	if engine.keepRecent != 10 {
		t.Errorf("keepRecent = %d; want 10", engine.keepRecent)
	}
}

func TestShouldCompact(t *testing.T) {
	engine := NewCompactionEngine(20, 10)

	tests := []struct {
		name     string
		msgCount int
		want     bool
	}{
		{"below threshold", 15, false},
		{"at threshold", 20, false},
		{"above threshold", 21, true},
		{"far above threshold", 50, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := make([]model.Message, tt.msgCount)
			got := engine.ShouldCompact(messages)
			if got != tt.want {
				t.Errorf("ShouldCompact() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestCompact(t *testing.T) {
	engine := NewCompactionEngine(20, 10)

	tests := []struct {
		name              string
		messages          []model.Message
		expectedCount     int
		expectedHasSummary bool
	}{
		{
			name: "no compaction needed - too few messages",
			messages: []model.Message{
				{Role: model.RoleSystem, Content: "system prompt"},
				{Role: model.RoleUser, Content: "hello"},
			},
			expectedCount:     2,
			expectedHasSummary: false,
		},
		{
			name: "compact with system prompt",
			messages: func() []model.Message {
				msgs := []model.Message{{Role: model.RoleSystem, Content: "system prompt"}}
				for i := 0; i < 25; i++ {
					msgs = append(msgs, model.Message{
						Role:    model.RoleUser,
						Content: "message " + string(rune(i)),
					})
				}
				return msgs
			}(),
			expectedCount:     12, // 1 system + 1 summary + 10 recent
			expectedHasSummary: true,
		},
		{
			name: "compact without system prompt",
			messages: func() []model.Message {
				msgs := []model.Message{}
				for i := 0; i < 25; i++ {
					msgs = append(msgs, model.Message{
						Role:    model.RoleUser,
						Content: "message " + string(rune(i)),
					})
				}
				return msgs
			}(),
			expectedCount:     11, // 1 summary + 10 recent
			expectedHasSummary: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, compacted := engine.Compact(tt.messages)

			if len(compacted) != tt.expectedCount {
				t.Errorf("compacted message count = %d; want %d", len(compacted), tt.expectedCount)
			}

			if result.OriginalCount != len(tt.messages) {
				t.Errorf("OriginalCount = %d; want %d", result.OriginalCount, len(tt.messages))
			}

			if result.CompactedCount != len(compacted) {
				t.Errorf("CompactedCount = %d; want %d", result.CompactedCount, len(compacted))
			}

			// 检查是否有摘要消息
			hasSummary := false
			for _, msg := range compacted {
				if msg.Role == model.RoleSystem && len(msg.Content) > 0 &&
					msg.Content != "system prompt" {
					hasSummary = true
					break
				}
			}

			if hasSummary != tt.expectedHasSummary {
				t.Errorf("has summary = %v; want %v", hasSummary, tt.expectedHasSummary)
			}
		})
	}
}

func TestCompact_PreservesSystemPrompt(t *testing.T) {
	engine := NewCompactionEngine(20, 10)

	messages := []model.Message{
		{Role: model.RoleSystem, Content: "you are a helpful assistant"},
	}
	for i := 0; i < 25; i++ {
		messages = append(messages, model.Message{
			Role:    model.RoleUser,
			Content: "test message",
		})
	}

	_, compacted := engine.Compact(messages)

	if compacted[0].Role != model.RoleSystem {
		t.Errorf("first message role = %s; want %s", compacted[0].Role, model.RoleSystem)
	}
	if compacted[0].Content != "you are a helpful assistant" {
		t.Errorf("system prompt not preserved")
	}
}

func TestCompact_PreservesRecentMessages(t *testing.T) {
	engine := NewCompactionEngine(20, 5)

	messages := []model.Message{{Role: model.RoleSystem, Content: "system"}}
	for i := 0; i < 20; i++ {
		messages = append(messages, model.Message{
			Role:    model.RoleUser,
			Content: "message " + string(rune('A'+i)),
		})
	}

	_, compacted := engine.Compact(messages)

	// 最后 5 条消息应该被保留
	recentStart := len(compacted) - 5
	for i := 0; i < 5; i++ {
		expected := "message " + string(rune('A'+15+i))
		if compacted[recentStart+i].Content != expected {
			t.Errorf("recent message[%d] = %s; want %s",
				i, compacted[recentStart+i].Content, expected)
		}
	}
}

func TestBuildSummary(t *testing.T) {
	engine := NewCompactionEngine(20, 10)

	messages := []model.Message{
		{Role: model.RoleUser, Content: "short message"},
		{Role: model.RoleAssistant, Content: "this is a very long message that should be truncated because it exceeds the 100 character limit for preview in the summary"},
	}

	summary := engine.buildSummary(messages)

	if summary == "" {
		t.Error("summary should not be empty")
	}

	// 检查摘要包含消息数量
	if !strings.Contains(summary, "2 messages") {
		t.Error("summary should contain message count")
	}

	// 检查长消息被截断（原始消息末尾的 "summary" 不应出现）
	if strings.Contains(summary, "in the summary") {
		t.Error("long message should be truncated")
	}

	// 检查截断标记存在
	if !strings.Contains(summary, "...") {
		t.Error("truncated message should have ellipsis")
	}
}
