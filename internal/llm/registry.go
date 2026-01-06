package llm

import (
	"fmt"
	"sync"
)

// ProviderFactory Provider 工厂函数
type ProviderFactory func(cfg ProviderConfig) (Provider, error)

// ProviderInfo Provider 信息
type ProviderInfo struct {
	ID           string
	Name         string
	SupportsTools bool
	DefaultModel string
	Factory      ProviderFactory
}

// Registry Provider 注册表
type Registry struct {
	mu        sync.RWMutex
	providers map[string]*ProviderInfo
}

var globalRegistry = NewRegistry()

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	r := &Registry{
		providers: make(map[string]*ProviderInfo),
	}
	r.registerBuiltinProviders()
	return r
}

// registerBuiltinProviders 注册内置 Provider
func (r *Registry) registerBuiltinProviders() {
	r.Register(&ProviderInfo{
		ID:            "openai",
		Name:          "OpenAI",
		SupportsTools: true,
		DefaultModel:  "gpt-4",
		Factory: func(cfg ProviderConfig) (Provider, error) {
			return newOpenAIProvider(cfg), nil
		},
	})

	r.Register(&ProviderInfo{
		ID:            "anthropic",
		Name:          "Anthropic Claude",
		SupportsTools: true,
		DefaultModel:  "claude-sonnet-4-5-20250929",
		Factory: func(cfg ProviderConfig) (Provider, error) {
			return newAnthropicProvider(cfg), nil
		},
	})

	r.Register(&ProviderInfo{
		ID:            "ollama",
		Name:          "Ollama",
		SupportsTools: true,
		DefaultModel:  "qwen2.5:14b",
		Factory: func(cfg ProviderConfig) (Provider, error) {
			return newOllamaProvider(cfg), nil
		},
	})
}

// Register 注册 Provider
func (r *Registry) Register(info *ProviderInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[info.ID] = info
}

// Get 获取 Provider 信息
func (r *Registry) Get(id string) (*ProviderInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", id)
	}
	return info, nil
}

// List 列出所有 Provider
func (r *Registry) List() []*ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*ProviderInfo, 0, len(r.providers))
	for _, info := range r.providers {
		list = append(list, info)
	}
	return list
}

// Create 创建 Provider 实例
func (r *Registry) Create(id string, cfg ProviderConfig) (Provider, error) {
	info, err := r.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider info: %w", err)
	}

	provider, err := info.Factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// GlobalRegistry 返回全局注册表
func GlobalRegistry() *Registry {
	return globalRegistry
}

// RegisterProvider 注册 Provider 到全局注册表
func RegisterProvider(info *ProviderInfo) {
	globalRegistry.Register(info)
}

// GetProvider 从全局注册表获取 Provider 信息
func GetProvider(id string) (*ProviderInfo, error) {
	return globalRegistry.Get(id)
}

// ListProviders 列出全局注册表中的所有 Provider
func ListProviders() []*ProviderInfo {
	return globalRegistry.List()
}

// CreateProvider 从全局注册表创建 Provider 实例
func CreateProvider(id string, cfg ProviderConfig) (Provider, error) {
	return globalRegistry.Create(id, cfg)
}
