package feedback

import (
	"context"
	"time"
)

// Feedback 用户反馈记录
type Feedback struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	MessageID  string    `json:"message_id"`
	Question   string    `json:"question"`
	Answer     string    `json:"answer"`
	Rating     int       `json:"rating"` // 1=好, -1=差, 0=未评价
	Comment    string    `json:"comment,omitempty"`
	Embedding  []float32 `json:"-"` // 问题的向量表示
	CreatedAt  time.Time `json:"created_at"`
}

// Repository 反馈存储接口
type Repository interface {
	Save(ctx context.Context, fb *Feedback) error
	GetByID(ctx context.Context, id string) (*Feedback, error)
	GetByMessageID(ctx context.Context, messageID string) (*Feedback, error)
	ListGoodFeedbacks(ctx context.Context, limit int) ([]*Feedback, error)
	SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]*Feedback, error)
}

// Embedder 向量化接口
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Service 反馈学习服务
type Service struct {
	repo     Repository
	embedder Embedder
}

// NewService 创建反馈服务
func NewService(repo Repository, embedder Embedder) *Service {
	return &Service{
		repo:     repo,
		embedder: embedder,
	}
}

// RecordFeedback 记录用户反馈
func (s *Service) RecordFeedback(ctx context.Context, fb *Feedback) error {
	// 生成问题的向量表示
	if s.embedder != nil && fb.Question != "" {
		embedding, err := s.embedder.Embed(ctx, fb.Question)
		if err == nil {
			fb.Embedding = embedding
		}
	}
	fb.CreatedAt = time.Now()
	return s.repo.Save(ctx, fb)
}

// GetSimilarGoodAnswers 获取相似问题的好答案
func (s *Service) GetSimilarGoodAnswers(ctx context.Context, question string, limit int) ([]*Feedback, error) {
	if s.embedder == nil {
		return nil, nil
	}

	embedding, err := s.embedder.Embed(ctx, question)
	if err != nil {
		return nil, err
	}

	return s.repo.SearchSimilar(ctx, embedding, limit)
}

// GetByMessageID 根据消息 ID 获取反馈
func (s *Service) GetByMessageID(ctx context.Context, messageID string) (*Feedback, error) {
	return s.repo.GetByMessageID(ctx, messageID)
}

// BuildPromptEnhancement 构建提示词增强
func (s *Service) BuildPromptEnhancement(ctx context.Context, question string) string {
	similar, err := s.GetSimilarGoodAnswers(ctx, question, 3)
	if err != nil || len(similar) == 0 {
		return ""
	}

	enhancement := "\n\n## 参考历史优质回答\n"
	for i, fb := range similar {
		enhancement += "\n### 示例 " + string(rune('1'+i)) + "\n"
		enhancement += "问题: " + fb.Question + "\n"
		enhancement += "回答: " + fb.Answer + "\n"
	}
	return enhancement
}
