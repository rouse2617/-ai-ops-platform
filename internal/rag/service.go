package rag

import (
	"context"
	"fmt"
	"time"

	"ai-ops/internal/llm"
	"ai-ops/internal/rag/chunker"
	"ai-ops/pkg/logger"

	"go.uber.org/zap"
)

// Service RAG 服务
type Service struct {
	vectorStore    VectorStore
	embedder       Embedder
	retriever      Retriever
	docRepo        DocumentRepository
	chunkRepo      ChunkRepository
	queryRewriter  QueryRewriter
	contextBuilder ContextBuilder
}

// Config RAG 服务配置
type Config struct {
	VectorStoreURL string
	EmbedderType   string        // openai / local
	EmbedderModel  string
	MaxTokens      int
	TopK           int
	EnableRerank   bool
}

// NewService 创建 RAG 服务
func NewService(
	vectorStore VectorStore,
	embedder Embedder,
	retriever Retriever,
	queryRewriter QueryRewriter,
	contextBuilder ContextBuilder,
	docRepo DocumentRepository,
	chunkRepo ChunkRepository,
) *Service {
	return &Service{
		vectorStore:    vectorStore,
		embedder:       embedder,
		retriever:      retriever,
		docRepo:        docRepo,
		chunkRepo:      chunkRepo,
		queryRewriter:  queryRewriter,
		contextBuilder: contextBuilder,
	}
}

// NewServiceWithConfig 使用配置创建 RAG 服务（工厂方法）
func NewServiceWithConfig(cfg Config, llmClient llm.Client, docRepo DocumentRepository, chunkRepo ChunkRepository) (*Service, error) {
	// 注意：所有组件需要在外部创建并传入，以避免循环依赖
	// 调用者应该：
	// 1. 创建 vectorstore.NewQdrantStore(...)
	// 2. 创建 embedder.NewOpenAIEmbedder(...)
	// 3. 创建 retriever.NewRetriever(...)
	// 4. ��建 retriever.NewQueryRewriter(...)
	// 5. 创建 retriever.NewContextBuilder(...)
	// 6. 调用 NewService(...) 组装服务
	return nil, fmt.Errorf("use NewService with pre-created components to avoid circular dependency")
}

// IndexDocument 索引文档
func (s *Service) IndexDocument(ctx context.Context, doc *Document) error {
	startTime := time.Now()

	// 1. 保存文档元数据
	if err := s.docRepo.Create(ctx, doc); err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	// 2. 文档分块
	chunks := s.chunkDocument(doc)
	logger.Info("Document chunked", zap.String("doc_id", doc.ID), zap.Int("chunks", len(chunks)))

	// 3. 生成 Embeddings
	for i := range chunks {
		embedding, err := s.embedder.Embed(ctx, chunks[i].Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
		}
		chunks[i].Embedding = embedding
	}

	// 4. 存储到向量数据库
	if err := s.vectorStore.Upsert(ctx, chunks); err != nil {
		return fmt.Errorf("failed to upsert chunks: %w", err)
	}

	// 5. 保存 Chunk 元数据
	if err := s.chunkRepo.BatchCreate(ctx, chunks); err != nil {
		return fmt.Errorf("failed to save chunks: %w", err)
	}

	logger.Info("Document indexed successfully",
		zap.String("doc_id", doc.ID),
		zap.Int("chunks", len(chunks)),
		zap.Duration("duration", time.Since(startTime)),
	)

	return nil
}

// Search 检索知识
func (s *Service) Search(ctx context.Context, query SearchQuery) (*KnowledgeContext, error) {
	startTime := time.Now()

	// 1. 查询改写（可选）
	queries := []string{query.Query}
	if query.Rerank {
		rewrittenQueries, err := s.queryRewriter.Rewrite(ctx, query.Query, nil)
		if err != nil {
			logger.Warn("Query rewriting failed, using original query", zap.Error(err))
		} else {
			queries = append(queries, rewrittenQueries...)
		}
	}

	// 2. 混合检索
	var allResults []SearchResult
	for _, q := range queries {
		// 创建新的查询对象
		searchQuery := SearchQuery{
			Query:          q,
			TopK:           query.TopK,
			Filters:        query.Filters,
			MinScore:       query.MinScore,
			IncludeContent: query.IncludeContent,
			Rerank:         query.Rerank,
		}

		results, err := s.retriever.HybridSearch(ctx, searchQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to search: %w", err)
		}

		allResults = append(allResults, results...)
	}

	// 3. 去重和重排序
	allResults = s.deduplicateResults(allResults)
	if query.Rerank && len(allResults) > 0 {
		allResults, _ = s.retriever.Rerank(ctx, query.Query, allResults)
	}

	// 4. 限制结果数量
	if len(allResults) > query.TopK {
		allResults = allResults[:query.TopK]
	}

	// 5. 构建上下文
	contextWindow := s.contextBuilder.BuildContext(allResults)

	// 6. 计算置信度
	confidence := s.calculateConfidence(allResults)

	retrievalTime := time.Since(startTime).Milliseconds()

	return &KnowledgeContext{
		Query:         query.Query,
		Results:       allResults,
		TotalResults:  len(allResults),
		RetrievalTime: retrievalTime,
		ContextWindow: contextWindow,
		Confidence:    confidence,
	}, nil
}

// SearchByAlert 根据告警检索知识
// TODO: 需要实现 alert 和 retriever 包
func (s *Service) SearchByAlert(ctx context.Context, alertData interface{}) (*KnowledgeContext, error) {
	// 暂时返回空，等待 alert 和 retriever 包实现
	return nil, fmt.Errorf("SearchByAlert not implemented yet")
}

// UpdateDocument 更新文档
func (s *Service) UpdateDocument(ctx context.Context, docID string, newContent string) error {
	// 1. 获取旧文档
	oldDoc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// 2. 检测变化类型
	changeType := s.detectChanges(oldDoc.Content, newContent)

	switch changeType {
	case "no_change":
		return nil

	case "metadata_only":
		// 仅更新元数据
		oldDoc.UpdatedAt = time.Now()
		return s.docRepo.Update(ctx, oldDoc)

	case "partial":
		// 部分更新：删除旧 chunks，重新索引
		if err := s.deleteDocumentChunks(ctx, docID); err != nil {
			return err
		}
		oldDoc.Content = newContent
		oldDoc.UpdatedAt = time.Now()
		oldDoc.Version++
		return s.IndexDocument(ctx, oldDoc)

	case "full":
		// 完全重新索引
		if err := s.DeleteDocument(ctx, docID); err != nil {
			return err
		}
		oldDoc.Content = newContent
		oldDoc.UpdatedAt = time.Now()
		oldDoc.Version++
		return s.IndexDocument(ctx, oldDoc)

	default:
		return fmt.Errorf("unknown change type: %s", changeType)
	}
}

// DeleteDocument 删除文档
func (s *Service) DeleteDocument(ctx context.Context, docID string) error {
	// 1. 删除 chunks
	if err := s.deleteDocumentChunks(ctx, docID); err != nil {
		return err
	}

	// 2. 删除文档元数据
	if err := s.docRepo.Delete(ctx, docID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// chunkDocument 文档分块
func (s *Service) chunkDocument(doc *Document) []*Chunk {
	splitter := chunker.NewRecursiveCharacterSplitter(chunker.ChunkConfig{
		ChunkSize:    800,
		ChunkOverlap: 200,
		Separators:   []string{"\n## ", "\n### ", "\n\n", "\n", ". ", " "},
	})

	textChunks := splitter.Split(doc.Content)
	chunks := make([]*Chunk, len(textChunks))

	for i, text := range textChunks {
		chunks[i] = &Chunk{
			ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, i),
			DocumentID: doc.ID,
			Content:    text,
			Metadata: map[string]interface{}{
				"source":      doc.Source,
				"category":    doc.Category,
				"tags":        doc.Tags,
				"created_at":  doc.CreatedAt,
			},
			Position:   i,
			TokenCount: len(text) / 4, // 粗略估算
		}
	}

	return chunks
}

// deleteDocumentChunks 删除文档的所有 chunks
func (s *Service) deleteDocumentChunks(ctx context.Context, docID string) error {
	// 1. 获取所有 chunk IDs
	chunks, err := s.chunkRepo.GetByDocumentID(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to get chunks: %w", err)
	}

	chunkIDs := make([]string, len(chunks))
	for i, chunk := range chunks {
		chunkIDs[i] = chunk.ID
	}

	// 2. 从向量数据库删除
	if err := s.vectorStore.Delete(ctx, chunkIDs); err != nil {
		return fmt.Errorf("failed to delete from vector store: %w", err)
	}

	// 3. 从数据库删除
	if err := s.chunkRepo.DeleteByDocumentID(ctx, docID); err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}

	return nil
}

// detectChanges 检测文档变化
func (s *Service) detectChanges(oldContent, newContent string) string {
	if oldContent == newContent {
		return "no_change"
	}

	// 简单的相似度计算（实际应使用更复杂的算法）
	similarity := s.calculateSimilarity(oldContent, newContent)

	if similarity > 0.9 {
		return "metadata_only"
	} else if similarity > 0.7 {
		return "partial"
	}

	return "full"
}

// calculateSimilarity 计算文本相似度
func (s *Service) calculateSimilarity(text1, text2 string) float32 {
	// 简单实现：基于长度差异
	len1, len2 := float32(len(text1)), float32(len(text2))
	if len1 == 0 || len2 == 0 {
		return 0
	}

	diff := len1 - len2
	if diff < 0 {
		diff = -diff
	}

	return 1.0 - (diff / max(len1, len2))
}

// deduplicateResults 去重检索结果
func (s *Service) deduplicateResults(results []SearchResult) []SearchResult {
	seen := make(map[string]bool)
	unique := make([]SearchResult, 0, len(results))

	for _, result := range results {
		if !seen[result.ChunkID] {
			seen[result.ChunkID] = true
			unique = append(unique, result)
		}
	}

	return unique
}

// calculateConfidence 计算置信度
func (s *Service) calculateConfidence(results []SearchResult) float32 {
	if len(results) == 0 {
		return 0
	}

	var totalScore float32
	for _, result := range results {
		totalScore += result.Score
	}

	return totalScore / float32(len(results))
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
