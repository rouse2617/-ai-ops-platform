package repository

import (
	"ai-ops/internal/model"

	"gorm.io/gorm"
)

// OperationRepository 操作仓库
type OperationRepository struct {
	db *gorm.DB
}

// NewOperationRepository 创建操作仓库
func NewOperationRepository(db *gorm.DB) *OperationRepository {
	return &OperationRepository{db: db}
}

// Create 创建确认记录
func (r *OperationRepository) Create(confirmation *model.OperationConfirmation) error {
	return r.db.Create(confirmation).Error
}

// Update 更新确认记录
func (r *OperationRepository) Update(confirmation *model.OperationConfirmation) error {
	return r.db.Save(confirmation).Error
}

// FindByID 根据 ID 查找
func (r *OperationRepository) FindByID(id string) (*model.OperationConfirmation, error) {
	var confirmation model.OperationConfirmation
	err := r.db.Where("id = ?", id).First(&confirmation).Error
	if err != nil {
		return nil, err
	}
	return &confirmation, nil
}

// FindBySession 根据会话查找
func (r *OperationRepository) FindBySession(sessionID string, limit int) ([]*model.OperationConfirmation, error) {
	var confirmations []*model.OperationConfirmation
	query := r.db.Where("session_id = ?", sessionID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&confirmations).Error
	return confirmations, err
}

// FindBySessionAndStatus 根据会话和状态查找
func (r *OperationRepository) FindBySessionAndStatus(sessionID string, status string) ([]*model.OperationConfirmation, error) {
	var confirmations []*model.OperationConfirmation
	err := r.db.Where("session_id = ? AND status = ?", sessionID, status).
		Order("created_at DESC").
		Find(&confirmations).Error
	return confirmations, err
}
