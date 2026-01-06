package agent

import (
	"context"
	"testing"
	"time"

	"ai-ops/internal/llm"
	"ai-ops/internal/tool"
)

// mockLLMClient 模拟 LLM 客户端
type mockLLMClient struct{}

func (m *mockLLMClient) Chat(ctx context.Context, messages []llm.Message) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{
		Message: llm.Message{Role: "assistant", Content: "Test response"},
	}, nil
}

func (m *mockLLMClient) ChatWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDef) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{
		Message: llm.Message{Role: "assistant", Content: "Test response with tools"},
	}, nil
}

func (m *mockLLMClient) ChatStream(ctx context.Context, messages []llm.Message, callback llm.StreamCallback) error {
	callback(llm.StreamChunk{Type: "content", Content: "Test"})
	callback(llm.StreamChunk{Type: "done"})
	return nil
}

func (m *mockLLMClient) ChatStreamWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDef, callback llm.StreamCallback) error {
	callback(llm.StreamChunk{Type: "content", Content: "Test"})
	callback(llm.StreamChunk{Type: "done"})
	return nil
}

func TestNewEnhancedAgent(t *testing.T) {
	registry := tool.NewRegistry()
	client := &mockLLMClient{}

	cfg := EnhancedConfig{
		Config: Config{
			MaxLoops: 5,
			Timeout:  time.Minute,
		},
		DoomLoopWindowSize: 10,
		DoomLoopThreshold:  3,
		MaxSnapshots:       50,
	}

	agent, err := NewEnhancedAgent(client, registry, nil, cfg)
	if err != nil {
		t.Fatalf("NewEnhancedAgent() error = %v", err)
	}

	if agent == nil {
		t.Fatal("agent should not be nil")
	}

	if agent.doomDetector == nil {
		t.Error("doomDetector should be initialized")
	}

	if agent.snapshots == nil {
		t.Error("snapshots should be initialized")
	}

	if agent.compaction == nil {
		t.Error("compaction should be initialized")
	}
}

func TestEnhancedAgent_Chat(t *testing.T) {
	registry := tool.NewRegistry()
	client := &mockLLMClient{}

	cfg := EnhancedConfig{
		Config: Config{
			MaxLoops: 5,
			Timeout:  time.Minute,
		},
	}

	agent, _ := NewEnhancedAgent(client, registry, nil, cfg)

	req := ChatRequest{
		SessionID: "test-session",
		Message:   "Hello",
	}

	resp, err := agent.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if resp.SessionID != "test-session" {
		t.Errorf("SessionID = %v, want test-session", resp.SessionID)
	}

	if resp.Reply == "" {
		t.Error("Reply should not be empty")
	}
}

func TestEnhancedAgent_ContextCompaction(t *testing.T) {
	registry := tool.NewRegistry()
	client := &mockLLMClient{}

	cfg := EnhancedConfig{
		Config: Config{
			MaxLoops: 5,
			Timeout:  time.Minute,
		},
		CompactionThreshold:  5,
		CompactionKeepRecent: 3,
	}

	agent, _ := NewEnhancedAgent(client, registry, nil, cfg)

	// 创建超过阈值的历史消息
	history := make([]llm.Message, 10)
	for i := 0; i < 10; i++ {
		history[i] = llm.NewUserMessage("Message")
	}

	req := ChatRequest{
		SessionID: "test-session",
		Message:   "Hello",
		History:   history,
	}

	messages := agent.buildMessagesWithCompaction(req)

	// 应该有: system + summary + 3 recent + user = 6
	if len(messages) > 7 {
		t.Errorf("Messages should be compacted, got %d", len(messages))
	}
}

func TestEnhancedAgent_DoomLoopReset(t *testing.T) {
	registry := tool.NewRegistry()
	client := &mockLLMClient{}

	cfg := EnhancedConfig{}
	agent, _ := NewEnhancedAgent(client, registry, nil, cfg)

	// 记录一些调用
	agent.doomDetector.Record("test_tool", nil)
	agent.doomDetector.Record("test_tool", nil)

	// 重置
	agent.ResetDoomDetector()

	// 检查是否重置
	isLoop, _ := agent.doomDetector.Check()
	if isLoop {
		t.Error("Doom detector should be reset")
	}
}

func TestEnhancedAgent_Snapshots(t *testing.T) {
	registry := tool.NewRegistry()
	client := &mockLLMClient{}

	cfg := EnhancedConfig{
		MaxSnapshots: 10,
	}
	agent, _ := NewEnhancedAgent(client, registry, nil, cfg)

	// 创建快照
	_, err := agent.snapshots.Create("session-1", "exec_command", nil, nil)
	if err != nil {
		t.Fatalf("Create snapshot error = %v", err)
	}

	// 获取快照列表
	snapshots := agent.GetSnapshots("session-1")
	if len(snapshots) != 1 {
		t.Errorf("Expected 1 snapshot, got %d", len(snapshots))
	}
}

func TestIsDangerousTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		want     bool
	}{
		{"exec_command is dangerous", "exec_command", true},
		{"run_script is dangerous", "run_script", true},
		{"check_cpu is safe", "check_cpu", false},
		{"list_hosts is safe", "list_hosts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDangerousTool(tt.toolName); got != tt.want {
				t.Errorf("isDangerousTool(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}
