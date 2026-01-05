package repository

import (
	"ai-ops/internal/model"

	"gorm.io/gorm"
)

// ScriptRepository 脚本仓库接口
type ScriptRepository interface {
	Create(script *model.Script) error
	Update(script *model.Script) error
	Delete(id string) error
	GetByID(id string) (*model.Script, error)
	GetByName(name string) (*model.Script, error)
	List(filter ScriptFilter) ([]*model.Script, error)
	UpdateEnabled(id string, enabled bool) error
}

// ScriptFilter 脚本过滤条件
type ScriptFilter struct {
	Keyword  string
	Language string
	Enabled  *bool
}

// scriptRepository 脚本仓库实现
type scriptRepository struct {
	db *gorm.DB
}

// NewScriptRepository 创建脚本仓库
func NewScriptRepository(db *gorm.DB) ScriptRepository {
	return &scriptRepository{db: db}
}

// Create 创建脚本
func (r *scriptRepository) Create(script *model.Script) error {
	return r.db.Create(script).Error
}

// Update 更新脚本
func (r *scriptRepository) Update(script *model.Script) error {
	return r.db.Model(&model.Script{}).Where("id = ?", script.ID).Updates(script).Error
}

// Delete 删除脚本
func (r *scriptRepository) Delete(id string) error {
	return r.db.Delete(&model.Script{}, "id = ?", id).Error
}

// GetByID 根据ID获取脚本
func (r *scriptRepository) GetByID(id string) (*model.Script, error) {
	var script model.Script
	err := r.db.Where("id = ?", id).First(&script).Error
	if err != nil {
		return nil, err
	}
	return &script, nil
}

// GetByName 根据名称获取脚本
func (r *scriptRepository) GetByName(name string) (*model.Script, error) {
	var script model.Script
	err := r.db.Where("name = ?", name).First(&script).Error
	if err != nil {
		return nil, err
	}
	return &script, nil
}

// List 列表查询
func (r *scriptRepository) List(filter ScriptFilter) ([]*model.Script, error) {
	var scripts []*model.Script
	query := r.db.Model(&model.Script{})

	if filter.Language != "" {
		query = query.Where("language = ?", filter.Language)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
	}

	err := query.Order("created_at DESC").Find(&scripts).Error
	if err != nil {
		return nil, err
	}

	return scripts, nil
}

// UpdateEnabled 更新启用状态
func (r *scriptRepository) UpdateEnabled(id string, enabled bool) error {
	return r.db.Model(&model.Script{}).Where("id = ?", id).Update("enabled", enabled).Error
}
