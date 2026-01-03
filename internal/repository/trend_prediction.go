package repository

import (
	"ai-ops/internal/model"
	"gorm.io/gorm"
)

type TrendPredictionRepository struct {
	db *gorm.DB
}

func NewTrendPredictionRepository(db *gorm.DB) *TrendPredictionRepository {
	return &TrendPredictionRepository{db: db}
}

func (r *TrendPredictionRepository) Create(prediction *model.TrendPrediction) error {
	return r.db.Create(prediction).Error
}

func (r *TrendPredictionRepository) FindByHost(hostID string, limit int) ([]*model.TrendPrediction, error) {
	var predictions []*model.TrendPrediction
	err := r.db.Where("host_id = ?", hostID).
		Order("created_at DESC").
		Limit(limit).
		Find(&predictions).Error
	return predictions, err
}

func (r *TrendPredictionRepository) FindAlerts(hostID string) ([]*model.TrendPrediction, error) {
	var predictions []*model.TrendPrediction
	err := r.db.Where("host_id = ? AND alert_level != ''", hostID).
		Order("created_at DESC").
		Find(&predictions).Error
	return predictions, err
}
