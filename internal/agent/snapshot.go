package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Snapshot 表示一个操作快照
type Snapshot struct {
	ID        string
	SessionID string
	ToolName  string
	Args      map[string]any
	State     map[string]any
	CreatedAt time.Time
}

// SnapshotManager 管理快照的创建、查询和删除
type SnapshotManager struct {
	mu        sync.RWMutex
	snapshots map[string]*Snapshot
	bySession map[string][]string
	maxCount  int
}

// NewSnapshotManager 创建快照管理器
func NewSnapshotManager(maxCount int) *SnapshotManager {
	if maxCount <= 0 {
		maxCount = 50
	}
	return &SnapshotManager{
		snapshots: make(map[string]*Snapshot),
		bySession: make(map[string][]string),
		maxCount:  maxCount,
	}
}

// Create 创建新快照
func (m *SnapshotManager) Create(sessionID, toolName string, args, state map[string]any) (*Snapshot, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID cannot be empty")
	}
	if toolName == "" {
		return nil, fmt.Errorf("toolName cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查快照总数限制
	if len(m.snapshots) >= m.maxCount {
		return nil, fmt.Errorf("snapshot limit reached: %d", m.maxCount)
	}

	snapshot := &Snapshot{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		ToolName:  toolName,
		Args:      args,
		State:     state,
		CreatedAt: time.Now(),
	}

	m.snapshots[snapshot.ID] = snapshot
	m.bySession[sessionID] = append(m.bySession[sessionID], snapshot.ID)

	return snapshot, nil
}

// Get 获取指定快照
func (m *SnapshotManager) Get(snapshotID string) (*Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot, ok := m.snapshots[snapshotID]
	if !ok {
		return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	return snapshot, nil
}

// List 列出会话的所有快照
func (m *SnapshotManager) List(sessionID string) []*Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := m.bySession[sessionID]
	result := make([]*Snapshot, 0, len(ids))
	for _, id := range ids {
		if snapshot, ok := m.snapshots[id]; ok {
			result = append(result, snapshot)
		}
	}

	return result
}

// Delete 删除指定快照
func (m *SnapshotManager) Delete(snapshotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot, ok := m.snapshots[snapshotID]
	if !ok {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	// 从 snapshots 中删除
	delete(m.snapshots, snapshotID)

	// 从 bySession 中删除
	sessionIDs := m.bySession[snapshot.SessionID]
	for i, id := range sessionIDs {
		if id == snapshotID {
			m.bySession[snapshot.SessionID] = append(sessionIDs[:i], sessionIDs[i+1:]...)
			break
		}
	}

	// 如果会话没有快照了，删除会话记录
	if len(m.bySession[snapshot.SessionID]) == 0 {
		delete(m.bySession, snapshot.SessionID)
	}

	return nil
}
