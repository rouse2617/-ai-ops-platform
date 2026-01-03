package repository

import (
	"context"
	"time"

	"ai-ops/internal/rag"

	"gorm.io/gorm"
)

// DocumentRepository 文档存储实现
type DocumentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建文档存储
func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create 创建文档
func (r *DocumentRepository) Create(ctx context.Context, doc *rag.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// GetByID 根据 ID 获取文档
func (r *DocumentRepository) GetByID(ctx context.Context, id string) (*rag.Document, error) {
	var doc rag.Document
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error
	return &doc, err
}

// Update 更新文档
func (r *DocumentRepository) Update(ctx context.Context, doc *rag.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

// Delete 删除文档
func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&rag.Document{}).Error
}

// List 列出文档
func (r *DocumentRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*rag.Document, error) {
	var docs []*rag.Document
	query := r.db.WithContext(ctx)

	// 应用过滤条件
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	err := query.Limit(limit).Offset(offset).Find(&docs).Error
	return docs, err
}

// ChunkRepository Chunk 存储实现
type ChunkRepository struct {
	db *gorm.DB
}

// NewChunkRepository 创建 Chunk 存储
func NewChunkRepository(db *gorm.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

// BatchCreate 批量创建 Chunks
func (r *ChunkRepository) BatchCreate(ctx context.Context, chunks []*rag.Chunk) error {
	return r.db.WithContext(ctx).CreateInBatches(chunks, 100).Error
}

// GetByID 根据 ID 获取 Chunk
func (r *ChunkRepository) GetByID(ctx context.Context, id string) (*rag.Chunk, error) {
	var chunk rag.Chunk
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&chunk).Error
	return &chunk, err
}

// GetByDocumentID 根据文档 ID 获取所有 Chunks
func (r *ChunkRepository) GetByDocumentID(ctx context.Context, docID string) ([]*rag.Chunk, error) {
	var chunks []*rag.Chunk
	err := r.db.WithContext(ctx).Where("document_id = ?", docID).Find(&chunks).Error
	return chunks, err
}

// DeleteByDocumentID 根据文档 ID 删除所有 Chunks
func (r *ChunkRepository) DeleteByDocumentID(ctx context.Context, docID string) error {
	return r.db.WithContext(ctx).Where("document_id = ?", docID).Delete(&rag.Chunk{}).Error
}

// IncidentRecordRepository 故障记录存储
type IncidentRecordRepository struct {
	db *gorm.DB
}

// NewIncidentRecordRepository 创建故障记录存储
func NewIncidentRecordRepository(db *gorm.DB) *IncidentRecordRepository {
	return &IncidentRecordRepository{db: db}
}

// Create 创建故障记录
func (r *IncidentRecordRepository) Create(ctx context.Context, record *rag.IncidentRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetByService 根据服务获取故障记录
func (r *IncidentRecordRepository) GetByService(ctx context.Context, service string, limit int) ([]*rag.IncidentRecord, error) {
	var records []*rag.IncidentRecord
	err := r.db.WithContext(ctx).
		Where("service = ?", service).
		Order("occurred_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// GetRecent 获取最近的故障记录
func (r *IncidentRecordRepository) GetRecent(ctx context.Context, days int, limit int) ([]*rag.IncidentRecord, error) {
	var records []*rag.IncidentRecord
	since := time.Now().AddDate(0, 0, -days)
	err := r.db.WithContext(ctx).
		Where("occurred_at >= ?", since).
		Order("occurred_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}
