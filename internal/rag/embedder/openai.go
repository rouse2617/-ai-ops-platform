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

// OpenAIEmbedder OpenAI Embedding 实现
type OpenAIEmbedder struct {
	apiKey     string
	endpoint   string
	model      string
	dimension  int
	httpClient *http.Client
}

// NewOpenAIEmbedder 创建 OpenAI Embedder
func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	dimension := 1536
	if model == "text-embedding-3-large" {
		dimension = 3072
	}

	return &OpenAIEmbedder{
		apiKey:     apiKey,
		endpoint:   "https://api.openai.com/v1/embeddings",
		model:      model,
		dimension:  dimension,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// NewOpenAIEmbedderWithEndpoint 创建自定义端点的 Embedder
func NewOpenAIEmbedderWithEndpoint(apiKey, endpoint, model string, dimension int) *OpenAIEmbedder {
	return &OpenAIEmbedder{
		apiKey:     apiKey,
		endpoint:   endpoint,
		model:      model,
		dimension:  dimension,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Embed 生成文本向量
func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := e.BatchEmbed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}
	return results[0], nil
}

// BatchEmbed 批量生成向量
func (e *OpenAIEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	const batchSize = 2048
	var allResults [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		results, err := e.doBatchEmbed(ctx, texts[i:end])
		if err != nil {
			return nil, fmt.Errorf("batch %d failed: %w", i/batchSize, err)
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func (e *OpenAIEmbedder) doBatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := map[string]interface{}{
		"input": texts,
		"model": e.model,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(texts))
	for _, item := range result.Data {
		if item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}

	return embeddings, nil
}

// Dimension 返回向量维度
func (e *OpenAIEmbedder) Dimension() int {
	return e.dimension
}
