package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaEmbedder Ollama 本地 Embedding 实现
type OllamaEmbedder struct {
	endpoint   string
	model      string
	dimension  int
	httpClient *http.Client
}

// NewOllamaEmbedder 创建 Ollama Embedder
func NewOllamaEmbedder(model string) *OllamaEmbedder {
	return NewOllamaEmbedderWithEndpoint("http://localhost:11434", model)
}

// NewOllamaEmbedderWithEndpoint 创建自定义端点的 Ollama Embedder
func NewOllamaEmbedderWithEndpoint(endpoint, model string) *OllamaEmbedder {
	// 常见模型的维度
	dimension := 768 // 默认维度
	switch model {
	case "nomic-embed-text":
		dimension = 768
	case "mxbai-embed-large":
		dimension = 1024
	case "all-minilm":
		dimension = 384
	}

	return &OllamaEmbedder{
		endpoint:   endpoint,
		model:      model,
		dimension:  dimension,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Embed 生成文本向量
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  e.model,
		"prompt": text,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.endpoint+"/api/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Ollama 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API 错误 %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 更新实际维度
	if len(result.Embedding) > 0 && e.dimension != len(result.Embedding) {
		e.dimension = len(result.Embedding)
	}

	return result.Embedding, nil
}

// BatchEmbed 批量生成向量
func (e *OllamaEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Ollama 不支持批量，逐个处理
	results := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := e.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("embedding 第 %d 个文本失败: %w", i, err)
		}
		results[i] = embedding
	}

	return results, nil
}

// Dimension 返回向量维度
func (e *OllamaEmbedder) Dimension() int {
	return e.dimension
}
