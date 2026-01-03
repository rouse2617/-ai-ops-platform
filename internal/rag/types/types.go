package types

import "time"

// SearchQuery 检索查询
type SearchQuery struct {
	Query          string            `json:"query"`
	TopK           int               `json:"top_k"`            // 返回结果数量
	Filters        map[string]string `json:"filters"`          // 过滤条件
	MinScore       float32           `json:"min_score"`        // 最小相似度分数
	IncludeContent bool              `json:"include_content"`  // 是否包含完整内容
	Rerank         bool              `json:"rerank"`           // 是否重排序
}

// SearchResult 检索结果
type SearchResult struct {
	ChunkID    string                 `json:"chunk_id"`
	DocumentID string                 `json:"document_id"`
	Content    string                 `json:"content"`
	Score      float32                `json:"score"`          // 相似度分数
	Metadata   map[string]interface{} `json:"metadata"`
	Highlights []string               `json:"highlights"`     // 高亮片段
}

// Chunk 文档分块
type Chunk struct {
	ID         string                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	Content    string                 `json:"content"`
	Embedding  []float32              `json:"embedding"`     // 向量表示
	Metadata   map[string]interface{} `json:"metadata"`      // 继承文档元数据
	Position   int                    `json:"position"`      // 在文档中的位置
	TokenCount int                    `json:"token_count"`   // Token 数量
}

// Document 知识库文档
type Document struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Source      string                 `json:"source"`       // 文档来源: sop/incident/manual/alert
	SourcePath  string                 `json:"source_path"`  // 原始文件路径
	Category    string                 `json:"category"`     // 分类: database/network/application
	Tags        []string               `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"`     // 扩展元数据
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     int                    `json:"version"`      // 文档版本
	Status      string                 `json:"status"`       // active/archived/draft
}

// KnowledgeContext RAG 上下文
type KnowledgeContext struct {
	Query          string         `json:"query"`
	Results        []SearchResult `json:"results"`
	TotalResults   int            `json:"total_results"`
	RetrievalTime  int64          `json:"retrieval_time_ms"`
	ContextWindow  string         `json:"context_window"`   // 组装后的上下文
	Confidence     float32        `json:"confidence"`       // 整体置信度
}

// IncidentRecord 故障处理记录
type IncidentRecord struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Service       string            `json:"service"`        // 受影响服务
	Severity      string            `json:"severity"`       // critical/high/medium/low
	RootCause     string            `json:"root_cause"`     // 根因分析
	Solution      string            `json:"solution"`       // 解决方案
	Commands      []string          `json:"commands"`       // 执行的命令
	Scripts       []string          `json:"scripts"`        // 使用的脚本
	Duration      int               `json:"duration"`       // 持续时间(分钟)
	OccurredAt    time.Time         `json:"occurred_at"`
	ResolvedAt    time.Time         `json:"resolved_at"`
	Tags          []string          `json:"tags"`
	RelatedAlerts []string          `json:"related_alerts"` // 关联告警
}
