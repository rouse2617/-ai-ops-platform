package llm

import (
	"fmt"
	"time"
)

// ProviderType Provider 类型
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
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

// NewClient 创建 LLM 客户端
func NewClient(cfg ProviderConfig) (Client, error) {
	switch cfg.Type {
	case ProviderOpenAI:
		return NewOpenAIClient(OpenAIConfig{
			Endpoint:  cfg.Endpoint,
			Model:     cfg.Model,
			APIKey:    cfg.APIKey,
			Timeout:   cfg.Timeout,
			MaxTokens: cfg.MaxTokens,
		}), nil
	case ProviderAnthropic:
		return nil, fmt.Errorf("anthropic provider not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
}
