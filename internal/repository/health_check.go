package repository

import (
	"ai-ops/internal/model"
	"gorm.io/gorm"
)

type HealthCheckRepository struct {
	db *gorm.DB
}

func NewHealthCheckRepository(db *gorm.DB) *HealthCheckRepository {
	return &HealthCheckRepository{db: db}
}

func (r *HealthCheckRepository) Create(check *model.HealthCheck) error {
	return r.db.Create(check).Error
}

func (r *HealthCheckRepository) FindByHost(hostID string, limit int) ([]*model.HealthCheck, error) {
	var checks []*model.HealthCheck
	err := r.db.Where("host_id = ?", hostID).
		Order("checked_at DESC").
		Limit(limit).
		Find(&checks).Error
	return checks, err
}

func (r *HealthCheckRepository) FindByID(id string) (*model.HealthCheck, error) {
	var check model.HealthCheck
	err := r.db.Where("id = ?", id).First(&check).Error
	return &check, err
}

// FindSimilarIssues 查找历史上类似的问题
func (r *HealthCheckRepository) FindSimilarIssues(hostID string, issueKeywords []string, limit int) ([]*model.HealthCheck, error) {
	var checks []*model.HealthCheck

	query := r.db.Where("host_id = ?", hostID).
		Where("status IN ?", []string{"warning", "critical"})

	// 如果有关键词，添加模糊匹配
	if len(issueKeywords) > 0 {
		for _, keyword := range issueKeywords {
			query = query.Where("issues LIKE ?", "%"+keyword+"%")
		}
	}

	err := query.Order("checked_at DESC").
		Limit(limit).
		Find(&checks).Error

	return checks, err
}

// FindRecentIssuesByType 按问题类型查找最近的问题
func (r *HealthCheckRepository) FindRecentIssuesByType(hostID string, checkType string, days int) ([]*model.HealthCheck, error) {
	var checks []*model.HealthCheck
	err := r.db.Where("host_id = ? AND check_type = ?", hostID, checkType).
		Where("checked_at > datetime('now', ?)", "-"+string(rune(days))+" days").
		Where("status IN ?", []string{"warning", "critical"}).
		Order("checked_at DESC").
		Limit(10).
		Find(&checks).Error
	return checks, err
}
