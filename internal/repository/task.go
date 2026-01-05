package repository

import (
	"ai-ops/internal/model"
	"fmt"

	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *model.Task) error
	GetByID(id string) (*model.Task, error)
	List(params ListTaskParams) ([]model.Task, int64, error)
	Update(task *model.Task) error
	Delete(id string) error
	BatchDelete(ids []string) error
	GetStats() (*TaskStats, error)
}

type ListTaskParams struct {
	Page      int
	PageSize  int
	Status    string
	Type      string
	StartDate string
	EndDate   string
	Keyword   string
}

type TaskStats struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Success   int64 `json:"success"`
	Failed    int64 `json:"failed"`
	Cancelled int64 `json:"cancelled"`
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) GetByID(id string) (*model.Task, error) {
	var task model.Task
	err := r.db.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) List(params ListTaskParams) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.Model(&model.Task{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.StartDate != "" {
		query = query.Where("start_time >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("start_time <= ?", params.EndDate)
	}
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		query = query.Where("input LIKE ? OR output LIKE ? OR host_name LIKE ?", keyword, keyword, keyword)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.Order("start_time DESC").Offset(offset).Limit(params.PageSize).Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *taskRepository) Update(task *model.Task) error {
	return r.db.Model(&model.Task{}).Where("id = ?", task.ID).Updates(task).Error
}

func (r *taskRepository) Delete(id string) error {
	return r.db.Delete(&model.Task{}, "id = ?", id).Error
}

func (r *taskRepository) BatchDelete(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.Task{}).Error
}

func (r *taskRepository) GetStats() (*TaskStats, error) {
	var stats TaskStats

	if err := r.db.Model(&model.Task{}).Count(&stats.Total).Error; err != nil {
		return nil, fmt.Errorf("统计总数失败: %w", err)
	}

	statusCounts := []struct {
		Status string
		Count  int64
	}{}

	if err := r.db.Model(&model.Task{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("统计状态失败: %w", err)
	}

	for _, sc := range statusCounts {
		switch sc.Status {
		case "pending":
			stats.Pending = sc.Count
		case "running":
			stats.Running = sc.Count
		case "success":
			stats.Success = sc.Count
		case "failed":
			stats.Failed = sc.Count
		case "cancelled":
			stats.Cancelled = sc.Count
		}
	}

	return &stats, nil
}
