package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

// PromptLoader 加载和渲染 prompt 模板
type PromptLoader struct {
	promptDir string
	cache     map[string]*template.Template
	mu        sync.RWMutex
}

// NewPromptLoader 创建 Prompt 加载器
func NewPromptLoader(promptDir string) *PromptLoader {
	return &PromptLoader{
		promptDir: promptDir,
		cache:     make(map[string]*template.Template),
	}
}

// Load 加载指定 provider 的 prompt 模板
func (l *PromptLoader) Load(provider string) (string, error) {
	tmpl, err := l.getTemplate(provider)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Render 渲染 prompt 模板
func (l *PromptLoader) Render(provider string, data map[string]any) (string, error) {
	tmpl, err := l.getTemplate(provider)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// getTemplate 获取模板（带缓存）
func (l *PromptLoader) getTemplate(provider string) (*template.Template, error) {
	l.mu.RLock()
	if tmpl, ok := l.cache[provider]; ok {
		l.mu.RUnlock()
		return tmpl, nil
	}
	l.mu.RUnlock()

	// 尝试加载 provider 特定模板
	providerPath := filepath.Join(l.promptDir, "providers", provider+".txt")
	content, err := os.ReadFile(providerPath)
	if err != nil {
		// 回退到 base.txt
		basePath := filepath.Join(l.promptDir, "base.txt")
		content, err = os.ReadFile(basePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read base template: %w", err)
		}
	}

	tmpl, err := template.New(provider).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	l.mu.Lock()
	l.cache[provider] = tmpl
	l.mu.Unlock()

	return tmpl, nil
}
