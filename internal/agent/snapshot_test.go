package agent

import (
	"sync"
	"testing"
)

func TestSnapshotManager_Create(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		toolName  string
		args      map[string]any
		state     map[string]any
		wantErr   bool
	}{
		{
			name:      "valid snapshot",
			sessionID: "session-1",
			toolName:  "execute_command",
			args:      map[string]any{"cmd": "ls"},
			state:     map[string]any{"cwd": "/tmp"},
			wantErr:   false,
		},
		{
			name:      "empty sessionID",
			sessionID: "",
			toolName:  "execute_command",
			args:      map[string]any{},
			state:     map[string]any{},
			wantErr:   true,
		},
		{
			name:      "empty toolName",
			sessionID: "session-1",
			toolName:  "",
			args:      map[string]any{},
			state:     map[string]any{},
			wantErr:   true,
		},
		{
			name:      "nil args and state",
			sessionID: "session-1",
			toolName:  "execute_command",
			args:      nil,
			state:     nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSnapshotManager(50)
			snapshot, err := m.Create(tt.sessionID, tt.toolName, tt.args, tt.state)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if snapshot == nil {
					t.Error("Create() returned nil snapshot")
					return
				}
				if snapshot.ID == "" {
					t.Error("snapshot ID is empty")
				}
				if snapshot.SessionID != tt.sessionID {
					t.Errorf("sessionID = %v, want %v", snapshot.SessionID, tt.sessionID)
				}
				if snapshot.ToolName != tt.toolName {
					t.Errorf("toolName = %v, want %v", snapshot.ToolName, tt.toolName)
				}
				if snapshot.CreatedAt.IsZero() {
					t.Error("CreatedAt is zero")
				}
			}
		})
	}
}

func TestSnapshotManager_MaxCount(t *testing.T) {
	m := NewSnapshotManager(3)

	// 创建 3 个快照应该成功
	for i := 0; i < 3; i++ {
		_, err := m.Create("session-1", "tool", nil, nil)
		if err != nil {
			t.Fatalf("Create() failed at %d: %v", i, err)
		}
	}

	// 第 4 个应该失败
	_, err := m.Create("session-1", "tool", nil, nil)
	if err == nil {
		t.Error("Create() should fail when limit reached")
	}
}

func TestSnapshotManager_Get(t *testing.T) {
	m := NewSnapshotManager(50)
	snapshot, _ := m.Create("session-1", "execute_command", map[string]any{"cmd": "ls"}, nil)

	tests := []struct {
		name       string
		snapshotID string
		wantErr    bool
	}{
		{
			name:       "existing snapshot",
			snapshotID: snapshot.ID,
			wantErr:    false,
		},
		{
			name:       "non-existing snapshot",
			snapshotID: "non-existing-id",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Get(tt.snapshotID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Get() returned nil snapshot")
			}
		})
	}
}

func TestSnapshotManager_List(t *testing.T) {
	m := NewSnapshotManager(50)

	// 创建多个会话的快照
	m.Create("session-1", "tool-1", nil, nil)
	m.Create("session-1", "tool-2", nil, nil)
	m.Create("session-2", "tool-3", nil, nil)

	tests := []struct {
		name      string
		sessionID string
		wantCount int
	}{
		{
			name:      "session with 2 snapshots",
			sessionID: "session-1",
			wantCount: 2,
		},
		{
			name:      "session with 1 snapshot",
			sessionID: "session-2",
			wantCount: 1,
		},
		{
			name:      "session with no snapshots",
			sessionID: "session-3",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshots := m.List(tt.sessionID)
			if len(snapshots) != tt.wantCount {
				t.Errorf("List() returned %d snapshots, want %d", len(snapshots), tt.wantCount)
			}
		})
	}
}

func TestSnapshotManager_Delete(t *testing.T) {
	m := NewSnapshotManager(50)
	snapshot, _ := m.Create("session-1", "tool", nil, nil)

	// 删除存在的快照
	err := m.Delete(snapshot.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// 验证快照已删除
	_, err = m.Get(snapshot.ID)
	if err == nil {
		t.Error("Get() should fail after Delete()")
	}

	// 验证会话列表已更新
	snapshots := m.List("session-1")
	if len(snapshots) != 0 {
		t.Errorf("List() returned %d snapshots after delete, want 0", len(snapshots))
	}

	// 删除不存在的快照
	err = m.Delete("non-existing-id")
	if err == nil {
		t.Error("Delete() should fail for non-existing snapshot")
	}
}

func TestSnapshotManager_Concurrent(t *testing.T) {
	m := NewSnapshotManager(100)
	var wg sync.WaitGroup

	// 并发创建快照
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_, err := m.Create("session-1", "tool", nil, nil)
				if err != nil {
					t.Errorf("concurrent Create() failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证创建了 50 个快照
	snapshots := m.List("session-1")
	if len(snapshots) != 50 {
		t.Errorf("concurrent operations created %d snapshots, want 50", len(snapshots))
	}
}

func TestNewSnapshotManager_DefaultMaxCount(t *testing.T) {
	tests := []struct {
		name     string
		maxCount int
		want     int
	}{
		{
			name:     "positive maxCount",
			maxCount: 10,
			want:     10,
		},
		{
			name:     "zero maxCount",
			maxCount: 0,
			want:     50,
		},
		{
			name:     "negative maxCount",
			maxCount: -1,
			want:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSnapshotManager(tt.maxCount)
			if m.maxCount != tt.want {
				t.Errorf("maxCount = %d, want %d", m.maxCount, tt.want)
			}
		})
	}
}
