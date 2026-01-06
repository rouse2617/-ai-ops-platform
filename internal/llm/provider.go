package llm

import (
	"context"
	"fmt"
	"time"
)

// Provider LLM 提供商接口（扩展 Client 接口）
type Provider interface {
	Client

	// ID 返回提供商唯一标识
	ID() string

	// SupportsTools 是否支持工具调用
	SupportsTools() bool

	// MaxTokens 返回最大 token 数
	MaxTokens() int
}

// ProviderType Provider 类型
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderOllama    ProviderType = "ollama"
)

// ProviderConfig Provider 配置
type ProviderConfig struct {
	Type      ProviderType
	Endpoint  string
	Model     string
	APIKey    string
	Timeout   time.Duration
	MaxTokens int
}

// NewProvider 创建 Provider 实例
func NewProvider(cfg ProviderConfig) (Provider, error) {
	switch cfg.Type {
	case ProviderOpenAI:
		return newOpenAIProvider(cfg), nil
	case ProviderAnthropic:
		return newAnthropicProvider(cfg), nil
	case ProviderOllama:
		return newOllamaProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
}

// NewClient 创建 LLM 客户端（向后兼容）
func NewClient(cfg ProviderConfig) (Client, error) {
	return NewProvider(cfg)
}

// openAIProvider OpenAI Provider 实现
type openAIProvider struct {
	client *OpenAIClient
}

func newOpenAIProvider(cfg ProviderConfig) Provider {
	return &openAIProvider{
		client: NewOpenAIClient(OpenAIConfig{
			Endpoint:  cfg.Endpoint,
			Model:     cfg.Model,
			APIKey:    cfg.APIKey,
			Timeout:   cfg.Timeout,
			MaxTokens: cfg.MaxTokens,
		}),
	}
}

func (p *openAIProvider) ID() string {
	return "openai"
}

func (p *openAIProvider) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	return p.client.Chat(ctx, messages)
}

func (p *openAIProvider) ChatWithTools(ctx context.Context, messages []Message, tools []ToolDef) (*ChatResponse, error) {
	return p.client.ChatWithTools(ctx, messages, tools)
}

func (p *openAIProvider) ChatStream(ctx context.Context, messages []Message, callback StreamCallback) error {
	return p.client.ChatStream(ctx, messages, callback)
}

func (p *openAIProvider) ChatStreamWithTools(ctx context.Context, messages []Message, tools []ToolDef, callback StreamCallback) error {
	return p.client.ChatStreamWithTools(ctx, messages, tools, callback)
}

func (p *openAIProvider) SupportsTools() bool {
	return true
}

func (p *openAIProvider) MaxTokens() int {
	return p.client.maxTokens
}

// anthropicProvider Anthropic Provider 实现
type anthropicProvider struct {
	client *AnthropicClient
}

func newAnthropicProvider(cfg ProviderConfig) Provider {
	return &anthropicProvider{
		client: NewAnthropicClient(AnthropicConfig{
			APIKey:    cfg.APIKey,
			BaseURL:   cfg.Endpoint,
			Model:     cfg.Model,
			MaxTokens: cfg.MaxTokens,
		}),
	}
}

func (p *anthropicProvider) ID() string {
	return "anthropic"
}

func (p *anthropicProvider) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	return p.client.Chat(ctx, messages)
}

func (p *anthropicProvider) ChatWithTools(ctx context.Context, messages []Message, tools []ToolDef) (*ChatResponse, error) {
	return p.client.ChatWithTools(ctx, messages, tools)
}

func (p *anthropicProvider) ChatStream(ctx context.Context, messages []Message, callback StreamCallback) error {
	return p.client.ChatStream(ctx, messages, callback)
}

func (p *anthropicProvider) ChatStreamWithTools(ctx context.Context, messages []Message, tools []ToolDef, callback StreamCallback) error {
	return p.client.ChatStreamWithTools(ctx, messages, tools, callback)
}

func (p *anthropicProvider) SupportsTools() bool {
	return true
}

func (p *anthropicProvider) MaxTokens() int {
	return int(p.client.maxTokens)
}

// ollamaProvider Ollama Provider 实现（复用 OpenAI 客户端）
type ollamaProvider struct {
	client *OpenAIClient
}

func newOllamaProvider(cfg ProviderConfig) Provider {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:11434/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen2.5:14b"
	}

	return &ollamaProvider{
		client: NewOpenAIClient(OpenAIConfig{
			Endpoint:  cfg.Endpoint,
			Model:     cfg.Model,
			APIKey:    cfg.APIKey,
			Timeout:   cfg.Timeout,
			MaxTokens: cfg.MaxTokens,
		}),
	}
}

func (p *ollamaProvider) ID() string {
	return "ollama"
}

func (p *ollamaProvider) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	return p.client.Chat(ctx, messages)
}

func (p *ollamaProvider) ChatWithTools(ctx context.Context, messages []Message, tools []ToolDef) (*ChatResponse, error) {
	return p.client.ChatWithTools(ctx, messages, tools)
}

func (p *ollamaProvider) ChatStream(ctx context.Context, messages []Message, callback StreamCallback) error {
	return p.client.ChatStream(ctx, messages, callback)
}

func (p *ollamaProvider) ChatStreamWithTools(ctx context.Context, messages []Message, tools []ToolDef, callback StreamCallback) error {
	return p.client.ChatStreamWithTools(ctx, messages, tools, callback)
}

func (p *ollamaProvider) SupportsTools() bool {
	return true
}

func (p *ollamaProvider) MaxTokens() int {
	return p.client.maxTokens
}
