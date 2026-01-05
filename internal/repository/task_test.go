package repository

import (
	"ai-ops/internal/model"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTaskTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpFile := "test_task_" + uuid.New().String() + ".db"
	db, err := gorm.Open(sqlite.Open(tmpFile), &gorm.Config{})
	require.NoError(t, err, "创建测试数据库失败")

	err = db.AutoMigrate(&model.Task{})
	require.NoError(t, err, "数据库迁移失败")

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		os.Remove(tmpFile)
	}

	return db, cleanup
}

func TestTaskRepository_Create(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	task := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "command",
		Status:    "pending",
		Input:     "ls -la",
		HostID:    "host-1",
		HostName:  "test-host",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(task)
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.SessionID, retrieved.SessionID)
	assert.Equal(t, task.Type, retrieved.Type)
	assert.Equal(t, task.Status, retrieved.Status)
	assert.Equal(t, task.Input, retrieved.Input)
}

func TestTaskRepository_GetByID(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	task := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "command",
		Status:    "success",
		Input:     "echo test",
		Output:    "test",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(task)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(task.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Output, retrieved.Output)
}

func TestTaskRepository_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	_, err := repo.GetByID("non-existent-id")
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestTaskRepository_List_Pagination(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	for i := 0; i < 25; i++ {
		task := &model.Task{
			ID:        uuid.New().String(),
			SessionID: "session-1",
			Type:      "command",
			Status:    "success",
			Input:     "test command",
			StartTime: time.Now().Add(time.Duration(i) * time.Second),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(task)
		require.NoError(t, err)
	}

	params := ListTaskParams{
		Page:     1,
		PageSize: 10,
	}
	tasks, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, tasks, 10)

	params.Page = 2
	tasks, total, err = repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, tasks, 10)

	params.Page = 3
	tasks, total, err = repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, tasks, 5)
}

func TestTaskRepository_List_FilterByStatus(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	statuses := []string{"pending", "running", "success", "failed"}
	for i, status := range statuses {
		for j := 0; j < 3; j++ {
			task := &model.Task{
				ID:        uuid.New().String(),
				SessionID: "session-1",
				Type:      "command",
				Status:    status,
				Input:     "test",
				StartTime: time.Now().Add(time.Duration(i*3+j) * time.Second),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := repo.Create(task)
			require.NoError(t, err)
		}
	}

	params := ListTaskParams{
		Page:     1,
		PageSize: 20,
		Status:   "success",
	}
	tasks, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, tasks, 3)
	for _, task := range tasks {
		assert.Equal(t, "success", task.Status)
	}
}

func TestTaskRepository_List_FilterByType(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	types := []string{"command", "script", "analysis"}
	for i, taskType := range types {
		for j := 0; j < 2; j++ {
			task := &model.Task{
				ID:        uuid.New().String(),
				SessionID: "session-1",
				Type:      taskType,
				Status:    "success",
				Input:     "test",
				StartTime: time.Now().Add(time.Duration(i*2+j) * time.Second),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := repo.Create(task)
			require.NoError(t, err)
		}
	}

	params := ListTaskParams{
		Page:     1,
		PageSize: 20,
		Type:     "script",
	}
	tasks, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	for _, task := range tasks {
		assert.Equal(t, "script", task.Type)
	}
}

func TestTaskRepository_List_FilterByDateRange(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		task := &model.Task{
			ID:        uuid.New().String(),
			SessionID: "session-1",
			Type:      "command",
			Status:    "success",
			Input:     "test",
			StartTime: baseTime.Add(time.Duration(i*24) * time.Hour),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(task)
		require.NoError(t, err)
	}

	params := ListTaskParams{
		Page:      1,
		PageSize:  20,
		StartDate: baseTime.Add(2 * 24 * time.Hour).Format(time.RFC3339),
		EndDate:   baseTime.Add(5 * 24 * time.Hour).Format(time.RFC3339),
	}
	tasks, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, tasks, 3)
}

func TestTaskRepository_List_FilterByKeyword(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	testCases := []struct {
		input    string
		output   string
		hostName string
	}{
		{"ls -la /home", "file list", "web-server"},
		{"ps aux", "process list", "db-server"},
		{"docker ps", "container list", "app-server"},
	}

	for i, tc := range testCases {
		task := &model.Task{
			ID:        uuid.New().String(),
			SessionID: "session-1",
			Type:      "command",
			Status:    "success",
			Input:     tc.input,
			Output:    tc.output,
			HostName:  tc.hostName,
			StartTime: time.Now().Add(time.Duration(i) * time.Second),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(task)
		require.NoError(t, err)
	}

	params := ListTaskParams{
		Page:     1,
		PageSize: 20,
		Keyword:  "docker",
	}
	tasks, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Contains(t, tasks[0].Input, "docker")

	params.Keyword = "server"
	tasks, total, err = repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
}

func TestTaskRepository_List_EmptyResult(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	params := ListTaskParams{
		Page:     1,
		PageSize: 20,
	}
	tasks, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, tasks, 0)
}

func TestTaskRepository_Update(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	task := &model.Task{
		ID:        uuid.New().String(),
		SessionID: "session-1",
		Type:      "command",
		Status:    "pending",
		Input:     "test",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(task)
	require.NoError(t, err)

	task.Status = "success"
	task.Output = "test output"
	endTime := time.Now()
	task.EndTime = &endTime
	task.Duration = 1000

	err = repo.Update(task)
	assert.NoError(t, err)

	updated, err := repo.GetByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "success", updated.Status)
	assert.Equal(t, "test output", updated.Output)
	assert.NotNil(t, updated.EndTime)
	assert.Equal(t, int64(1000), updated.Duration)
}

func TestTaskRepository_Delete(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

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

	err := repo.Create(task)
	require.NoError(t, err)

	err = repo.Delete(task.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(task.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestTaskRepository_BatchDelete(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

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
		err := repo.Create(task)
		require.NoError(t, err)
	}

	err := repo.BatchDelete(ids[:3])
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err := repo.GetByID(ids[i])
		assert.Error(t, err)
	}

	for i := 3; i < 5; i++ {
		_, err := repo.GetByID(ids[i])
		assert.NoError(t, err)
	}
}

func TestTaskRepository_BatchDelete_EmptyList(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	err := repo.BatchDelete([]string{})
	assert.NoError(t, err)
}

func TestTaskRepository_GetStats(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

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
			err := repo.Create(task)
			require.NoError(t, err)
		}
	}

	stats, err := repo.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(20), stats.Total)
	assert.Equal(t, int64(2), stats.Pending)
	assert.Equal(t, int64(3), stats.Running)
	assert.Equal(t, int64(10), stats.Success)
	assert.Equal(t, int64(4), stats.Failed)
	assert.Equal(t, int64(1), stats.Cancelled)
}

func TestTaskRepository_GetStats_Empty(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	stats, err := repo.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.Total)
	assert.Equal(t, int64(0), stats.Pending)
	assert.Equal(t, int64(0), stats.Running)
	assert.Equal(t, int64(0), stats.Success)
	assert.Equal(t, int64(0), stats.Failed)
	assert.Equal(t, int64(0), stats.Cancelled)
}

func TestTaskRepository_List_OrderByStartTime(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		task := &model.Task{
			ID:        uuid.New().String(),
			SessionID: "session-1",
			Type:      "command",
			Status:    "success",
			Input:     "test",
			StartTime: baseTime.Add(time.Duration(i) * time.Hour),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(task)
		require.NoError(t, err)
	}

	params := ListTaskParams{
		Page:     1,
		PageSize: 20,
	}
	tasks, _, err := repo.List(params)
	assert.NoError(t, err)
	assert.Len(t, tasks, 5)

	for i := 0; i < len(tasks)-1; i++ {
		assert.True(t, tasks[i].StartTime.After(tasks[i+1].StartTime) || tasks[i].StartTime.Equal(tasks[i+1].StartTime))
	}
}
