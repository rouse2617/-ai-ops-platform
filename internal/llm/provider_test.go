package llm_test

import (
	"context"
	"testing"
	"time"

	"ai-ops/internal/llm"
)

// TestProviderRegistry 测试 Provider 注册表
func TestProviderRegistry(t *testing.T) {
	registry := llm.NewRegistry()

	// 测试列出所有 Provider
	providers := registry.List()
	if len(providers) != 3 {
		t.Errorf("expected 3 providers, got %d", len(providers))
	}

	// 测试获取 Provider 信息
	info, err := registry.Get("openai")
	if err != nil {
		t.Fatalf("failed to get openai provider: %v", err)
	}
	if info.ID != "openai" {
		t.Errorf("expected provider ID 'openai', got '%s'", info.ID)
	}
	if !info.SupportsTools {
		t.Error("openai should support tools")
	}

	// 测试获取不存在的 Provider
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

// TestProviderCreation 测试 Provider 创建
func TestProviderCreation(t *testing.T) {
	tests := []struct {
		name         string
		providerType llm.ProviderType
		expectedID   string
	}{
		{"OpenAI", llm.ProviderOpenAI, "openai"},
		{"Anthropic", llm.ProviderAnthropic, "anthropic"},
		{"Ollama", llm.ProviderOllama, "ollama"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := llm.NewProvider(llm.ProviderConfig{
				Type:      tt.providerType,
				Endpoint:  "http://localhost:8080",
				Model:     "test-model",
				APIKey:    "test-key",
				Timeout:   30 * time.Second,
				MaxTokens: 4096,
			})

			if err != nil {
				t.Fatalf("failed to create provider: %v", err)
			}

			if provider.ID() != tt.expectedID {
				t.Errorf("expected ID '%s', got '%s'", tt.expectedID, provider.ID())
			}

			if !provider.SupportsTools() {
				t.Errorf("provider %s should support tools", tt.expectedID)
			}

			if provider.MaxTokens() != 4096 {
				t.Errorf("expected MaxTokens 4096, got %d", provider.MaxTokens())
			}
		})
	}
}

// TestProviderInterface 测试 Provider 实现了 Client 接口
func TestProviderInterface(t *testing.T) {
	provider, err := llm.NewProvider(llm.ProviderConfig{
		Type:     llm.ProviderOpenAI,
		Endpoint: "http://localhost:11434/v1",
		Model:    "test-model",
	})

	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// 验证 Provider 可以作为 Client 使用
	var _ llm.Client = provider

	// 验证基本方法存在
	ctx := context.Background()
	messages := []llm.Message{
		llm.NewUserMessage("test"),
	}

	// 这些调用会失败（因为没有真实的服务器），但验证了接口实现
	_, err = provider.Chat(ctx, messages)
	if err == nil {
		t.Log("Chat method exists and can be called")
	}

	_, err = provider.ChatWithTools(ctx, messages, nil)
	if err == nil {
		t.Log("ChatWithTools method exists and can be called")
	}
}

// TestGlobalRegistry 测试全局注册表
func TestGlobalRegistry(t *testing.T) {
	// 测试全局注册表函数
	providers := llm.ListProviders()
	if len(providers) < 3 {
		t.Errorf("expected at least 3 providers in global registry, got %d", len(providers))
	}

	// 测试获取 Provider
	info, err := llm.GetProvider("anthropic")
	if err != nil {
		t.Fatalf("failed to get anthropic provider: %v", err)
	}
	if info.Name != "Anthropic Claude" {
		t.Errorf("expected name 'Anthropic Claude', got '%s'", info.Name)
	}

	// 测试创建 Provider
	provider, err := llm.CreateProvider("ollama", llm.ProviderConfig{
		Endpoint:  "http://localhost:11434/v1",
		Model:     "qwen2.5:14b",
		MaxTokens: 8192,
	})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	if provider.ID() != "ollama" {
		t.Errorf("expected ID 'ollama', got '%s'", provider.ID())
	}
}

// TestCustomProviderRegistration 测试自定义 Provider 注册
func TestCustomProviderRegistration(t *testing.T) {
	registry := llm.NewRegistry()

	// 注册自定义 Provider
	customInfo := &llm.ProviderInfo{
		ID:            "custom",
		Name:          "Custom Provider",
		SupportsTools: false,
		DefaultModel:  "custom-model",
		Factory: func(cfg llm.ProviderConfig) (llm.Provider, error) {
			// 返回一个基于 OpenAI 的自定义实现
			return llm.NewProvider(llm.ProviderConfig{
				Type:      llm.ProviderOpenAI,
				Endpoint:  cfg.Endpoint,
				Model:     cfg.Model,
				APIKey:    cfg.APIKey,
				Timeout:   cfg.Timeout,
				MaxTokens: cfg.MaxTokens,
			})
		},
	}

	registry.Register(customInfo)

	// 验证注册成功
	info, err := registry.Get("custom")
	if err != nil {
		t.Fatalf("failed to get custom provider: %v", err)
	}
	if info.Name != "Custom Provider" {
		t.Errorf("expected name 'Custom Provider', got '%s'", info.Name)
	}

	// 验证可以创建实例
	provider, err := registry.Create("custom", llm.ProviderConfig{
		Endpoint: "http://localhost:8080",
		Model:    "test",
	})
	if err != nil {
		t.Fatalf("failed to create custom provider: %v", err)
	}
	if provider == nil {
		t.Error("provider should not be nil")
	}
}
