package vectorstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai-ops/internal/rag/types"
)

// QdrantStore Qdrant 向量存储实现
type QdrantStore struct {
	endpoint       string
	collectionName string
	apiKey         string
	dimension      int
	httpClient     *http.Client
}

// NewQdrantStore 创建 Qdrant 存储
func NewQdrantStore(endpoint, collectionName string, dimension int) (*QdrantStore, error) {
	store := &QdrantStore{
		endpoint:       endpoint,
		collectionName: collectionName,
		dimension:      dimension,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}

	if err := store.ensureCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	return store, nil
}

// Upsert 插入或更新向量
func (q *QdrantStore) Upsert(ctx context.Context, chunks []*types.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	points := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		points[i] = map[string]interface{}{
			"id":     chunk.ID,
			"vector": chunk.Embedding,
			"payload": map[string]interface{}{
				"document_id": chunk.DocumentID,
				"content":     chunk.Content,
				"position":    chunk.Position,
				"metadata":    chunk.Metadata,
			},
		}
	}

	body := map[string]interface{}{"points": points}
	return q.doRequest(ctx, "PUT", fmt.Sprintf("/collections/%s/points", q.collectionName), body, nil)
}

// Search 向量检索
func (q *QdrantStore) Search(ctx context.Context, query types.SearchQuery) ([]types.SearchResult, error) {
	return q.vectorSearch(ctx, query)
}

// Delete 删除向量
func (q *QdrantStore) Delete(ctx context.Context, chunkIDs []string) error {
	if len(chunkIDs) == 0 {
		return nil
	}

	body := map[string]interface{}{
		"points": chunkIDs,
	}
	return q.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/delete", q.collectionName), body, nil)
}

// HybridSearch 混合检索（向量 + 关键词 RRF 融合）
func (q *QdrantStore) HybridSearch(ctx context.Context, query types.SearchQuery) ([]types.SearchResult, error) {
	vectorResults, err := q.vectorSearch(ctx, query)
	if err != nil {
		return nil, err
	}

	keywordResults, _ := q.keywordSearch(ctx, query)

	return q.rrfFusion(vectorResults, keywordResults, query.TopK), nil
}

// vectorSearch 向量检索
func (q *QdrantStore) vectorSearch(ctx context.Context, query types.SearchQuery) ([]types.SearchResult, error) {
	body := map[string]interface{}{
		"vector":       query.Query,
		"limit":        query.TopK,
		"with_payload": true,
		"score_threshold": query.MinScore,
	}

	if len(query.Filters) > 0 {
		body["filter"] = q.buildFilter(query.Filters)
	}

	var resp struct {
		Result []struct {
			ID      string                 `json:"id"`
			Score   float32                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	if err := q.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/search", q.collectionName), body, &resp); err != nil {
		return nil, err
	}

	results := make([]types.SearchResult, len(resp.Result))
	for i, r := range resp.Result {
		results[i] = types.SearchResult{
			ChunkID:    r.ID,
			DocumentID: getString(r.Payload, "document_id"),
			Content:    getString(r.Payload, "content"),
			Score:      r.Score,
			Metadata:   r.Payload,
		}
	}
	return results, nil
}

// keywordSearch 关键词检索
func (q *QdrantStore) keywordSearch(ctx context.Context, query types.SearchQuery) ([]types.SearchResult, error) {
	body := map[string]interface{}{
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{"key": "content", "match": map[string]interface{}{"text": query.Query}},
			},
		},
		"limit":        query.TopK,
		"with_payload": true,
	}

	var resp struct {
		Result struct {
			Points []struct {
				ID      string                 `json:"id"`
				Payload map[string]interface{} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	if err := q.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/scroll", q.collectionName), body, &resp); err != nil {
		return nil, err
	}

	results := make([]types.SearchResult, len(resp.Result.Points))
	for i, p := range resp.Result.Points {
		results[i] = types.SearchResult{
			ChunkID:    p.ID,
			DocumentID: getString(p.Payload, "document_id"),
			Content:    getString(p.Payload, "content"),
			Score:      0.5,
			Metadata:   p.Payload,
		}
	}
	return results, nil
}

// rrfFusion RRF 融合算法
func (q *QdrantStore) rrfFusion(vectorResults, keywordResults []types.SearchResult, topK int) []types.SearchResult {
	const k = 60.0
	scores := make(map[string]float32)
	resultMap := make(map[string]types.SearchResult)

	for rank, r := range vectorResults {
		scores[r.ChunkID] += float32(1.0 / (k + float64(rank+1)))
		resultMap[r.ChunkID] = r
	}
	for rank, r := range keywordResults {
		scores[r.ChunkID] += float32(1.0 / (k + float64(rank+1)))
		if _, exists := resultMap[r.ChunkID]; !exists {
			resultMap[r.ChunkID] = r
		}
	}

	results := make([]types.SearchResult, 0, len(resultMap))
	for id, result := range resultMap {
		result.Score = scores[id]
		results = append(results, result)
	}

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > topK {
		results = results[:topK]
	}
	return results
}

// ensureCollection 确保集合存在
func (q *QdrantStore) ensureCollection(ctx context.Context) error {
	var resp struct {
		Result struct {
			Exists bool `json:"exists"`
		} `json:"result"`
	}

	err := q.doRequest(ctx, "GET", fmt.Sprintf("/collections/%s", q.collectionName), nil, &resp)
	if err == nil {
		return nil
	}

	body := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     q.dimension,
			"distance": "Cosine",
		},
	}
	return q.doRequest(ctx, "PUT", fmt.Sprintf("/collections/%s", q.collectionName), body, nil)
}

// doRequest 执行 HTTP 请求
func (q *QdrantStore) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, q.endpoint+path, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if q.apiKey != "" {
		req.Header.Set("api-key", q.apiKey)
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant error %d: %s", resp.StatusCode, string(data))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (q *QdrantStore) buildFilter(filters map[string]string) map[string]interface{} {
	must := make([]map[string]interface{}, 0, len(filters))
	for key, value := range filters {
		must = append(must, map[string]interface{}{
			"key":   key,
			"match": map[string]interface{}{"value": value},
		})
	}
	return map[string]interface{}{"must": must}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
