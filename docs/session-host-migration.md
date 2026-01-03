# Session-Host 关联数据库迁移方案

## 概述

本文档描述了 AI-Ops 项目中 Session 与 Host 关联关系的数据库迁移方案，实现了会话与主机的关联追踪，支持切换主机时的对话隔离。

## 数据模型变更

### 1. Session 模型（已有）

```go
type Session struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title"`
    Hosts     []string  `json:"hosts" gorm:"serializer:json"`  // 关联的主机列表
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**说明**：
- `Hosts` 字段存储会话关联的所有主机 ID
- 支持一个会话关联多个主机
- 使用 JSON 序列化存储

### 2. Message 模型（新增字段）

```go
type Message struct {
    ID        string     `json:"id" gorm:"primaryKey"`
    SessionID string     `json:"session_id" gorm:"index;not null"`
    HostID    string     `json:"host_id" gorm:"index"`           // 新增：执行消息的目标主机
    Role      string     `json:"role"`
    Content   string     `json:"content" gorm:"type:text"`
    ToolCalls []ToolCall `json:"tool_calls" gorm:"serializer:json"`
    CreatedAt time.Time  `json:"created_at"`
}
```

**变更**：
- 新增 `HostID` 字段，记录消息关联的主机
- 添加索引以优化查询性能

## 数据库索引优化

### 新增索引

```sql
-- 复合索引：session_id + host_id
CREATE INDEX IF NOT EXISTS idx_messages_session_host
ON messages(session_id, host_id);
```

**用途**：
- 快速查询特定会话中特定主机的消息
- 支持主机隔离查询
- 优化按主机过滤的性能

## Repository 层变更

### SessionRepository 接口扩展

```go
type SessionRepository interface {
    // 原有方法
    Create(session *model.Session) error
    GetByID(id string) (*model.Session, error)
    List(limit int) ([]*model.Session, error)
    Delete(id string) error
    AddMessage(sessionID string, msg *model.Message) error
    GetMessages(sessionID string) ([]*model.Message, error)
    UpdateTitle(sessionID string, title string) error

    // 新增方法
    GetMessagesByHost(sessionID string, hostID string) ([]*model.Message, error)
    UpdateHosts(sessionID string, hosts []string) error
}
```

**新增功能**：
1. `GetMessagesByHost`：获取会话中指定主机的消息（支持主机隔离）
2. `UpdateHosts`：更新会话关联的主机列表

## Service 层变更

### ChatService 修改

```go
// 保存消息时自动记录 host_id
func (s *ChatService) saveUserMessage(req ChatRequest) error {
    hostID := s.getCurrentHostID(req.Hosts)
    userMsg := &model.Message{
        ID:        uuid.New().String(),
        SessionID: req.SessionID,
        HostID:    hostID,  // 记录当前主机
        Role:      model.RoleUser,
        Content:   req.Message,
        CreatedAt: time.Now(),
    }
    return s.sessionRepo.AddMessage(req.SessionID, userMsg)
}

// 获取当前活动主机
func (s *ChatService) getCurrentHostID(hosts []string) string {
    if len(hosts) == 0 {
        return ""
    }
    return hosts[0]  // 使用第一个主机作为当前主机
}
```

**实现逻辑**：
- 每次保存消息时，自动记录当前活动的主机 ID
- 用户消息和助手消息都会记录 host_id
- 支持空主机（全局会话）

## 迁移工具

### 执行迁移

```bash
# 查看迁移统计（模拟运行）
go run cmd/migrate-session-host/main.go --dry-run

# 执行实际迁移
go run cmd/migrate-session-host/main.go

# 指定数据库路径
go run cmd/migrate-session-host/main.go --db ./data/aiops.db
```

### 迁移工具功能

1. **自动迁移表结构**：添加 `host_id` 字段
2. **创建索引**：创建 `idx_messages_session_host` 索引
3. **统计信息**：显示迁移前后的数据统计

## 使用场景

### 1. 单主机会话

```json
{
  "session_id": "session-123",
  "message": "查看磁盘使用情况",
  "hosts": ["host-001"]
}
```

所有消息的 `host_id` 都是 `host-001`。

### 2. 多主机会话

```json
{
  "session_id": "session-123",
  "message": "查看磁盘使用情况",
  "hosts": ["host-001", "host-002"]
}
```

当前消息的 `host_id` 是 `host-001`（第一个主机）。

### 3. 切换主机

```json
// 第一条消息
{
  "session_id": "session-123",
  "message": "查看 host-001 的状态",
  "hosts": ["host-001"]
}

// 切换主机后的消息
{
  "session_id": "session-123",
  "message": "查看 host-002 的状态",
  "hosts": ["host-002"]
}
```

每条消息都记录了对应的主机 ID，实现对话隔离。

### 4. 按主机过滤消息

```go
// 获取会话中特定主机的消息
messages, err := sessionRepo.GetMessagesByHost("session-123", "host-001")
```

## 兼容性说明

### 向后兼容

- 旧数据的 `host_id` 字段为空字符串
- 查询时会包含所有消息（包括旧消息）
- 不影响现有功能

### 数据迁移

- 使用 GORM AutoMigrate 自动添加字段
- 不需要手动修改数据库
- 支持 SQLite 的 ALTER TABLE 操作

## 性能优化

### 索引策略

1. **单列索引**：
   - `session_id`（已有）
   - `host_id`（新增）

2. **复合索引**：
   - `(session_id, host_id)`：优化按会话和主机过滤
   - `(session_id, created_at)`（已有）：优化按时间排序

### 查询优化

```go
// 优化前：全表扫描
SELECT * FROM messages WHERE session_id = ? AND host_id = ?

// 优化后：使用复合索引
SELECT * FROM messages WHERE session_id = ? AND host_id = ?
-- 使用 idx_messages_session_host 索引
```

## 测试建议

### 单元测试

```go
func TestGetMessagesByHost(t *testing.T) {
    // 创建测试会话和消息
    session := &model.Session{ID: "test-session", Hosts: []string{"host-1", "host-2"}}
    repo.Create(session)

    // 添加不同主机的消息
    repo.AddMessage("test-session", &model.Message{HostID: "host-1", Content: "msg1"})
    repo.AddMessage("test-session", &model.Message{HostID: "host-2", Content: "msg2"})

    // 测试按主机过滤
    messages, err := repo.GetMessagesByHost("test-session", "host-1")
    assert.NoError(t, err)
    assert.Len(t, messages, 1)
    assert.Equal(t, "msg1", messages[0].Content)
}
```

### 集成测试

1. 创建会话并关联多个主机
2. 发送消息到不同主机
3. 验证消息的 host_id 正确记录
4. 测试按主机过滤功能
5. 测试切换主机场景

## 回滚方案

如果需要回滚：

```sql
-- 删除索引
DROP INDEX IF EXISTS idx_messages_session_host;

-- 删除字段（SQLite 不支持 DROP COLUMN，需要重建表）
-- 建议保留字段，设置为空即可
UPDATE messages SET host_id = '';
```

## 总结

本次迁移实现了以下功能：

1. ✅ Session 记录关联的 hosts
2. ✅ Message 记录执行的目标主机
3. ✅ 支持切换主机时的对话隔离
4. ✅ 优化查询性能（复合索引）
5. ✅ 向后兼容旧数据
6. ✅ 提供迁移工具

## 相关文件

- 模型定义：`C:\Users\hrp\Downloads\ai-pro\internal\model\message.go`
- 数据库初始化：`C:\Users\hrp\Downloads\ai-pro\internal\repository\db.go`
- Repository 实现：`C:\Users\hrp\Downloads\ai-pro\internal\repository\session.go`
- Service 实现：`C:\Users\hrp\Downloads\ai-pro\internal\service\chat_service.go`
- 迁移工具：`C:\Users\hrp\Downloads\ai-pro\cmd\migrate-session-host\main.go`
