package handler

import (
	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskRepo repository.TaskRepository
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskRepo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{
		taskRepo: taskRepo,
	}
}

// ListTasks 获取任务列表
// GET /api/tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := repository.ListTaskParams{
		Page:      page,
		PageSize:  pageSize,
		Status:    c.Query("status"),
		Type:      c.Query("type"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Keyword:   c.Query("keyword"),
	}

	tasks, total, err := h.taskRepo.List(params)
	if err != nil {
		InternalError(c, "获取任务列表失败: "+err.Error())
		return
	}

	Success(c, gin.H{
		"list":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTask 获取单个任务
// GET /api/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	task, err := h.taskRepo.GetByID(id)
	if err != nil {
		NotFound(c, "任务不存在")
		return
	}

	Success(c, task)
}

// CancelTask 取消任务
// POST /api/tasks/:id/cancel
func (h *TaskHandler) CancelTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	task, err := h.taskRepo.GetByID(id)
	if err != nil {
		NotFound(c, "任务不存在")
		return
	}

	if task.Status != "pending" && task.Status != "running" {
		ParamError(c, "只能取消待执行或执行中的任务")
		return
	}

	task.Status = "cancelled"
	now := time.Now()
	task.EndTime = &now
	if task.StartTime.IsZero() {
		task.StartTime = now
	}
	task.Duration = now.Sub(task.StartTime).Milliseconds()

	if err := h.taskRepo.Update(task); err != nil {
		InternalError(c, "取消任务失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "任务已取消", task)
}

// RetryTask 重试任务
// POST /api/tasks/:id/retry
func (h *TaskHandler) RetryTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	oldTask, err := h.taskRepo.GetByID(id)
	if err != nil {
		NotFound(c, "任务不存在")
		return
	}

	newTask := &model.Task{
		ID:        uuid.New().String(),
		SessionID: oldTask.SessionID,
		Type:      oldTask.Type,
		Status:    "pending",
		Input:     oldTask.Input,
		HostID:    oldTask.HostID,
		HostName:  oldTask.HostName,
		ToolName:  oldTask.ToolName,
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.taskRepo.Create(newTask); err != nil {
		InternalError(c, "创建重试任务失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "任务已重新创建", newTask)
}

// DeleteTask 删除任务
// DELETE /api/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ParamError(c, "id 不能为空")
		return
	}

	if err := h.taskRepo.Delete(id); err != nil {
		InternalError(c, "删除任务失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "删除成功", nil)
}

// BatchDeleteTasks 批量删除任务
// POST /api/tasks/batch-delete
func (h *TaskHandler) BatchDeleteTasks(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, "参数错误: "+err.Error())
		return
	}

	if err := h.taskRepo.BatchDelete(req.IDs); err != nil {
		InternalError(c, "批量删除失败: "+err.Error())
		return
	}

	SuccessWithMessage(c, "删除成功", gin.H{
		"deleted": len(req.IDs),
	})
}

// GetTaskStats 获取任务统计
// GET /api/tasks/stats
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	stats, err := h.taskRepo.GetStats()
	if err != nil {
		InternalError(c, "获取统计信息失败: "+err.Error())
		return
	}

	Success(c, stats)
}
