package handler

import (
	"ai-ops/internal/model"
	"ai-ops/internal/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockTaskRepository struct {
	tasks map[string]*model.Task
}

func newMockTaskRepository() repository.TaskRepository {
	return &mockTaskRepository{
		tasks: make(map[string]*model.Task),
	}
}

func (m *mockTaskRepository) Create(task *model.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) GetByID(id string) (*model.Task, error) {
	task, ok := m.tasks[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return task, nil
}

func (m *mockTaskRepository) List(params repository.ListTaskParams) ([]model.Task, int64, error) {
	tasks := make([]model.Task, 0)
	for _, task := range m.tasks {
		if params.Status != "" && task.Status != params.Status {
			continue
		}
		if params.Type != "" && task.Type != params.Type {
			continue
		}
		if params.Keyword != "" {
			if task.Input != params.Keyword && task.Output != params.Keyword && task.HostName != params.Keyword {
				continue
			}
		}
		tasks = append(tasks, *task)
	}

	total := int64(len(tasks))
	start := params.Page*params.PageSize - params.PageSize
	end := start + params.PageSize
	if start > len(tasks) {
		return []model.Task{}, total, nil
	}
	if end > len(tasks) {
		end = len(tasks)
	}
	return tasks[start:end], total, nil
}

func (m *mockTaskRepository) Update(task *model.Task) error {
	if _, ok := m.tasks[task.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) Delete(id string) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepository) BatchDelete(ids []string) error {
	for _, id := range ids {
		delete(m.tasks, id)
	}
	return nil
}

func (m *mockTaskRepository) GetStats() (*repository.TaskStats, error) {
	stats := &repository.TaskStats{}
	for _, task := range m.tasks {
		stats.Total++
		switch task.Status {
		case "pending":
			stats.Pending++
		case "running":
			stats.Running++
		case "success":
			stats.Success++
		case "failed":
			stats.Failed++
		case "cancelled":
			stats.Cancelled++
		}
	}
	return stats, nil
}

func setupTaskHandler() (*TaskHandler, *gin.Engine) {
	taskRepo := newMockTaskRepository()
	handler := NewTaskHandler(taskRepo)

	r := gin.New()
	api := r.Group("/api")
	{
		tasks := api.Group("/tasks")
		{
			tasks.GET("", handler.ListTasks)
			tasks.GET("/:id", handler.GetTask)
			tasks.DELETE("/:id", handler.DeleteTask)
			tasks.POST("/batch-delete", handler.BatchDeleteTasks)
		}
		api.GET("/tasks/stats", handler.GetTaskStats)
	}

	return handler, r
}

func TestListTasks_Empty(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.NotNil(t, data["list"])
	assert.Equal(t, float64(0), data["total"])
}

func TestListTasks_WithPagination(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	for i := 0; i < 25; i++ {
		task := &model.Task{
			ID:        uuid.New().String(),
			SessionID: "session-1",
			Type:      "command",
			Status:    "success",
			Input:     "test command",
			StartTime: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(task)
	}

	req, _ := http.NewRequest("GET", "/api/tasks?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(25), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["page_size"])
	list := data["list"].([]interface{})
	assert.Len(t, list, 10)
}

func TestListTasks_FilterByStatus(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	statuses := []string{"pending", "running", "success", "failed"}
	for _, status := range statuses {
		for i := 0; i < 3; i++ {
			task := &model.Task{
				ID:        uuid.New().String(),
				SessionID: "session-1",
				Type:      "command",
				Status:    status,
				Input:     "test",
				StartTime: time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			repo.Create(task)
		}
	}

	req, _ := http.NewRequest("GET", "/api/tasks?status=success", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(3), data["total"])
}

func TestListTasks_FilterByType(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	types := []string{"command", "script", "analysis"}
	for _, taskType := range types {
		for i := 0; i < 2; i++ {
			task := &model.Task{
				ID:        uuid.New().String(),
				SessionID: "session-1",
				Type:      taskType,
				Status:    "success",
				Input:     "test",
				StartTime: time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			repo.Create(task)
		}
	}

	req, _ := http.NewRequest("GET", "/api/tasks?type=script", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(2), data["total"])
}

func TestListTasks_DefaultPagination(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	for i := 0; i < 5; i++ {
		task := &model.Task{
			ID:        uuid.New().String(),
			SessionID: "session-1",
			Type:      "command",
			Status:    "success",
			Input:     "test",
			StartTime: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Create(task)
	}

	req, _ := http.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(20), data["page_size"])
}

func TestGetTask_Success(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	task := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "command",
		Status:    "success",
		Input:     "ls -la",
		Output:    "file list",
		HostID:    "host-1",
		HostName:  "test-host",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(task)

	req, _ := http.NewRequest("GET", "/api/tasks/"+task.ID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, task.ID, data["id"])
	assert.Equal(t, task.Input, data["input"])
	assert.Equal(t, task.Output, data["output"])
}

func TestGetTask_NotFound(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("GET", "/api/tasks/non-existent-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1003, resp.Code)
}

func TestGetTask_EmptyID(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("GET", "/api/tasks/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEqual(t, 0, resp.Code)
}

func TestDeleteTask_Success(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	task := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "command",
		Status:    "success",
		Input:     "test",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(task)

	req, _ := http.NewRequest("DELETE", "/api/tasks/"+task.ID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	_, err = repo.GetByID(task.ID)
	assert.Error(t, err)
}

func TestDeleteTask_EmptyID(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("DELETE", "/api/tasks/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEqual(t, 0, resp.Code)
}

func TestBatchDeleteTasks_Success(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	ids := make([]string, 5)
	for i := 0; i < 5; i++ {
		task := &model.Task{
			ID:        uuid.New().String(),
			SessionID: "session-1",
			Type:      "command",
			Status:    "success",
			Input:     "test",
			StartTime: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		ids[i] = task.ID
		repo.Create(task)
	}

	body := map[string]interface{}{
		"ids": ids[:3],
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/tasks/batch-delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	for i := 0; i < 3; i++ {
		_, err := repo.GetByID(ids[i])
		assert.Error(t, err)
	}

	for i := 3; i < 5; i++ {
		_, err := repo.GetByID(ids[i])
		assert.NoError(t, err)
	}
}

func TestBatchDeleteTasks_EmptyList(t *testing.T) {
	_, r := setupTaskHandler()

	body := map[string]interface{}{
		"ids": []string{},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/tasks/batch-delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestBatchDeleteTasks_MissingIDs(t *testing.T) {
	_, r := setupTaskHandler()

	body := map[string]interface{}{}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/tasks/batch-delete", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1001, resp.Code)
}

func TestBatchDeleteTasks_InvalidJSON(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("POST", "/api/tasks/batch-delete", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1001, resp.Code)
}

func TestGetTaskStats_Success(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	testData := map[string]int{
		"pending":   2,
		"running":   3,
		"success":   10,
		"failed":    4,
		"cancelled": 1,
	}

	for status, count := range testData {
		for i := 0; i < count; i++ {
			task := &model.Task{
				ID:        uuid.New().String(),
				SessionID: "session-1",
				Type:      "command",
				Status:    status,
				Input:     "test",
				StartTime: time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			repo.Create(task)
		}
	}

	req, _ := http.NewRequest("GET", "/api/tasks/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(20), data["total"])
	assert.Equal(t, float64(2), data["pending"])
	assert.Equal(t, float64(3), data["running"])
	assert.Equal(t, float64(10), data["success"])
	assert.Equal(t, float64(4), data["failed"])
	assert.Equal(t, float64(1), data["cancelled"])
}

func TestGetTaskStats_Empty(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("GET", "/api/tasks/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(0), data["total"])
	assert.Equal(t, float64(0), data["pending"])
	assert.Equal(t, float64(0), data["running"])
	assert.Equal(t, float64(0), data["success"])
	assert.Equal(t, float64(0), data["failed"])
	assert.Equal(t, float64(0), data["cancelled"])
}

func TestListTasks_InvalidPageNumber(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("GET", "/api/tasks?page=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["page"])
}

func TestListTasks_InvalidPageSize(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("GET", "/api/tasks?page_size=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(20), data["page_size"])
}

func TestListTasks_PageSizeExceedsLimit(t *testing.T) {
	_, r := setupTaskHandler()

	req, _ := http.NewRequest("GET", "/api/tasks?page_size=200", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(20), data["page_size"])
}

func TestListTasks_MultipleFilters(t *testing.T) {
	handler, r := setupTaskHandler()
	repo := handler.taskRepo.(*mockTaskRepository)

	task1 := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "command",
		Status:    "success",
		Input:     "ls -la",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	task2 := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "script",
		Status:    "success",
		Input:     "backup.sh",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	task3 := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "command",
		Status:    "failed",
		Input:     "rm -rf",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.Create(task1)
	repo.Create(task2)
	repo.Create(task3)

	req, _ := http.NewRequest("GET", "/api/tasks?status=success&type=command", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
}
