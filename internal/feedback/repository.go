package feedback

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"gorm.io/gorm"
)

// SQLiteRepository SQLite 实现的反馈存储
type SQLiteRepository struct {
	db *gorm.DB
}

// FeedbackRecord 数据库记录
type FeedbackRecord struct {
	ID        string    `gorm:"primaryKey"`
	SessionID string    `gorm:"index"`
	MessageID string    `gorm:"uniqueIndex"`
	Question  string    `gorm:"type:text"`
	Answer    string    `gorm:"type:text"`
	Rating    int       `gorm:"index"`
	Comment   string    `gorm:"type:text"`
	Embedding string    `gorm:"type:text"` // JSON 存储向量
	CreatedAt time.Time `gorm:"index"`
}

func (FeedbackRecord) TableName() string {
	return "feedbacks"
}

// NewSQLiteRepository 创建 SQLite 存储
func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	db.AutoMigrate(&FeedbackRecord{})
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Save(ctx context.Context, fb *Feedback) error {
	embeddingJSON := ""
	if len(fb.Embedding) > 0 {
		data, _ := json.Marshal(fb.Embedding)
		embeddingJSON = string(data)
	}

	record := &FeedbackRecord{
		ID:        fb.ID,
		SessionID: fb.SessionID,
		MessageID: fb.MessageID,
		Question:  fb.Question,
		Answer:    fb.Answer,
		Rating:    fb.Rating,
		Comment:   fb.Comment,
		Embedding: embeddingJSON,
		CreatedAt: fb.CreatedAt,
	}

	return r.db.WithContext(ctx).Save(record).Error
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*Feedback, error) {
	var record FeedbackRecord
	if err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return r.recordToFeedback(&record), nil
}

func (r *SQLiteRepository) GetByMessageID(ctx context.Context, messageID string) (*Feedback, error) {
	var record FeedbackRecord
	if err := r.db.WithContext(ctx).First(&record, "message_id = ?", messageID).Error; err != nil {
		return nil, err
	}
	return r.recordToFeedback(&record), nil
}

func (r *SQLiteRepository) ListGoodFeedbacks(ctx context.Context, limit int) ([]*Feedback, error) {
	var records []FeedbackRecord
	if err := r.db.WithContext(ctx).
		Where("rating = 1").
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}

	feedbacks := make([]*Feedback, len(records))
	for i, record := range records {
		feedbacks[i] = r.recordToFeedback(&record)
	}
	return feedbacks, nil
}

// SearchSimilar 搜索相似问题（使用余弦相似度）
func (r *SQLiteRepository) SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]*Feedback, error) {
	// 获取所有好评反馈
	var records []FeedbackRecord
	if err := r.db.WithContext(ctx).
		Where("rating = 1 AND embedding != ''").
		Find(&records).Error; err != nil {
		return nil, err
	}

	// 计算相似度并排序
	type scored struct {
		record *FeedbackRecord
		score  float64
	}
	var scored_list []scored

	for i := range records {
		var storedEmb []float32
		if err := json.Unmarshal([]byte(records[i].Embedding), &storedEmb); err != nil {
			continue
		}
		score := cosineSimilarity(embedding, storedEmb)
		if score > 0.7 { // 相似度阈值
			scored_list = append(scored_list, scored{&records[i], score})
		}
	}

	// 按相似度排序
	for i := 0; i < len(scored_list)-1; i++ {
		for j := i + 1; j < len(scored_list); j++ {
			if scored_list[j].score > scored_list[i].score {
				scored_list[i], scored_list[j] = scored_list[j], scored_list[i]
			}
		}
	}

	// 取前 N 个
	if len(scored_list) > limit {
		scored_list = scored_list[:limit]
	}

	feedbacks := make([]*Feedback, len(scored_list))
	for i, s := range scored_list {
		feedbacks[i] = r.recordToFeedback(s.record)
	}
	return feedbacks, nil
}

func (r *SQLiteRepository) recordToFeedback(record *FeedbackRecord) *Feedback {
	fb := &Feedback{
		ID:        record.ID,
		SessionID: record.SessionID,
		MessageID: record.MessageID,
		Question:  record.Question,
		Answer:    record.Answer,
		Rating:    record.Rating,
		Comment:   record.Comment,
		CreatedAt: record.CreatedAt,
	}
	if record.Embedding != "" {
		json.Unmarshal([]byte(record.Embedding), &fb.Embedding)
	}
	return fb
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
