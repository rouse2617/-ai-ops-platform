package retriever

import (
	"context"
	"sort"

	"ai-ops/internal/rag"
)

// Retriever 检索器
type Retriever struct {
	vectorStore rag.VectorStore
	embedder    rag.Embedder
	reranker    rag.Reranker
}

// NewRetriever 创建检索器
func NewRetriever(vectorStore rag.VectorStore, embedder rag.Embedder, enableRerank bool) *Retriever {
	var reranker rag.Reranker
	// TODO: 实现 CrossEncoderReranker
	// if enableRerank {
	// 	reranker = NewCrossEncoderReranker()
	// }

	return &Retriever{
		vectorStore: vectorStore,
		embedder:    embedder,
		reranker:    reranker,
	}
}

// HybridSearch 混合检索
func (r *Retriever) HybridSearch(ctx context.Context, query rag.SearchQuery) ([]rag.SearchResult, error) {
	// 1. 向量检索
	vectorResults, err := r.vectorStore.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	// 2. 关键词检索（如果支持）
	// keywordResults, _ := r.vectorStore.KeywordSearch(ctx, query)

	// 3. RRF 融合
	results := r.fuseResults(vectorResults, nil, query.TopK)

	return results, nil
}

// Rerank 重排序
func (r *Retriever) Rerank(ctx context.Context, query string, results []rag.SearchResult) ([]rag.SearchResult, error) {
	if r.reranker == nil {
		return results, nil
	}

	return r.reranker.Rerank(ctx, query, results)
}

// fuseResults RRF 融合多个检索结果
func (r *Retriever) fuseResults(vectorResults, keywordResults []rag.SearchResult, topK int) []rag.SearchResult {
	const k = 60 // RRF 常数

	scores := make(map[string]float32)
	resultMap := make(map[string]rag.SearchResult)

	// 向量检索结果
	for rank, result := range vectorResults {
		scores[result.ChunkID] += 1.0 / float32(k+rank+1)
		resultMap[result.ChunkID] = result
	}

	// 关键词检索结果
	for rank, result := range keywordResults {
		scores[result.ChunkID] += 1.0 / float32(k+rank+1)
		if _, exists := resultMap[result.ChunkID]; !exists {
			resultMap[result.ChunkID] = result
		}
	}

	// 按融合分数排序
	type scoredResult struct {
		result rag.SearchResult
		score  float32
	}

	var sorted []scoredResult
	for chunkID, score := range scores {
		result := resultMap[chunkID]
		result.Score = score
		sorted = append(sorted, scoredResult{result, score})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	// 返回 Top-K
	results := make([]rag.SearchResult, 0, topK)
	for i := 0; i < len(sorted) && i < topK; i++ {
		results = append(results, sorted[i].result)
	}

	return results
}
