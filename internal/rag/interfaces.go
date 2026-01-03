package rag

import "context"

// VectorStore 向量存储接口
type VectorStore interface {
	// Upsert 插入或更新向量
	Upsert(ctx context.Context, chunks []*Chunk) error

	// Search 向量检索
	Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)

	// Delete 删除向量
	Delete(ctx context.Context, chunkIDs []string) error

	// HybridSearch 混合检索（向量 + 关键词）
	HybridSearch(ctx context.Context, query SearchQuery) ([]SearchResult, error)
}

// Embedder 向量化接口
type Embedder interface {
	// Embed 生成文本向量
	Embed(ctx context.Context, text string) ([]float32, error)

	// BatchEmbed 批量生成向量
	BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension 返回向量维度
	Dimension() int
}

// DocumentRepository 文档存储接口
type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id string) (*Document, error)
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*Document, error)
}

// ChunkRepository Chunk 存储接口
type ChunkRepository interface {
	BatchCreate(ctx context.Context, chunks []*Chunk) error
	GetByID(ctx context.Context, id string) (*Chunk, error)
	GetByDocumentID(ctx context.Context, docID string) ([]*Chunk, error)
	DeleteByDocumentID(ctx context.Context, docID string) error
}

// Reranker 重排序接口
type Reranker interface {
	Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error)
}

// Retriever 检索器接口
type Retriever interface {
	HybridSearch(ctx context.Context, query SearchQuery) ([]SearchResult, error)
	Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error)
}

// QueryRewriter 查询改写器接口
type QueryRewriter interface {
	Rewrite(ctx context.Context, query string, context map[string]interface{}) ([]string, error)
}

// ContextBuilder 上下文构建器接口
type ContextBuilder interface {
	BuildContext(results []SearchResult) string
}
