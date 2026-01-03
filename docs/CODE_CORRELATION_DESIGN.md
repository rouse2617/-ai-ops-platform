# 代码级关联诊断功能技术设计文档

## 1. 功能概述

当 AI 监控系统检测到 Prometheus 指标异常时，自动分析最近的 Git 提交记录，识别可能导致异常的代码变更，并通过 WebSocket 主动推送智能诊断结果到前端。

### 核心价值
- **快速定位根因**：将 MTTR（平均修复时间）从小时级降低到分钟级
- **智能关联分析**：基于时间窗口、文件路径、关键词的多维度评分
- **主动通知机制**：无需人工查询，AI 主动推送可疑变更

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         Alert Pipeline                          │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│  AlertManager (internal/monitor/alert_manager.go)               │
│  - HandleAlert()                                                │
│  - analyzeCodeCorrelation()                                     │
└─────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    ▼                           ▼
┌──────────────────────────────┐  ┌──────────────────────────────┐
│  CodeChangeCorrelator        │  │  AlertNotifier               │
│  (internal/monitor/          │  │  (internal/monitor/          │
│   correlator.go)             │  │   notifier.go)               │
│  - AnalyzeAlert()            │  │  - NotifyCodeCorrelation()   │
│  - rankSuspiciousFiles()     │  │  - WebSocket broadcast       │
│  - calculateScore()          │  └──────────────────────────────┘
└──────────────────────────────┘                │
                    │                            │
                    ▼                            ▼
┌──────────────────────────────┐  ┌──────────────────────────────┐
│  GitChangeTracker            │  │  Frontend Components         │
│  (internal/tool/builtin/     │  │  - CodeCorrelationCard.vue   │
│   git_change_tracker.go)     │  │  - DiffViewer.vue            │
│  - GetCommitsInTimeRange()   │  │  - WebSocket listener        │
│  - GetDiff()                 │  └──────────────────────────────┘
│  - getChangedFiles()         │
└──────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────┐
│  Git Repositories            │
│  (go-git library)            │
└──────────────────────────────┘
```

## 3. 数据流图

```
Alert Triggered
      │
      ▼
[1] AlertManager.HandleAlert()
      │
      ├─► [2] Notifier.Notify() ──────────► Frontend (Basic Alert)
      │
      └─► [3] Go Routine: analyzeCodeCorrelation()
                │
                ▼
          [4] Correlator.AnalyzeAlert()
                │
                ├─► [5] GitTracker.GetCommitsInTimeRange()
                │         │
                │         └─► Git Repository (go-git)
                │
                ├─► [6] rankSuspiciousFiles()
                │         - Time proximity scoring
                │         - Path matching
                │         - Keyword analysis
                │
                └─► [7] calculateCorrelationScore()
                          │
                          ▼
                    [8] Score >= 0.3?
                          │
                          ├─ Yes ─► [9] Notifier.NotifyCodeCorrelation()
                          │                │
                          │                └─► WebSocket ─► Frontend Card
                          │
                          └─ No ──► Log & Skip
```

## 4. 核心组件设计

### 4.1 GitChangeTracker (Git 变更追踪器)

**文件**: `internal/tool/builtin/git_change_tracker.go`

**职责**:
- 管理多个 Git 仓库连接
- 查询指定时间范围内的提交记录
- 提取 commit 的变更文件列表
- 生成 diff 内容

**关键方法**:
```go
// 添加仓库到追踪列表
AddRepository(name, path string) error

// 获取时间范围内的提交
GetCommitsInTimeRange(ctx, repoName, start, end) ([]GitCommit, error)

// 获取提交的 diff
GetDiff(repoName, commitHash string) (string, error)

// 提取变更文件列表
getChangedFiles(commit) ([]string, error)
```

**依赖**: `github.com/go-git/go-git/v5`

### 4.2 CodeChangeCorrelator (关联分析器)

**文件**: `internal/monitor/correlator.go`

**职责**:
- 维护服务与代码路径的映射关系
- 分析告警与代码变更的关联性
- 对可疑文件进行评分排序
- 计算整体关联度分数

**关联度评分算法**:
```
文件得分 = 时间得分 + 路径匹配得分 + 关键词得分 + 文件类型得分

时间得分 = 1.0 - (告警时间 - 提交时间) / 30分钟
  - 越接近告警时间，得分越高

路径匹配得分 = 0.3 (如果文件路径匹配服务配置的 code_paths)
  - 例如: user-service 配置了 "src/services/user/**"

关键词得分 = 0.2 (如果 commit message 包含服务关键词)
  - 例如: commit message 包含 "database" 或 "query"

告警上下文得分 = 0.3 (如果文件名包含告警相关关键词)
  - 例如: 告警名为 "high_database_latency"，文件名包含 "db_query.py"

文件类型得分 = 0.1 (如果是代码文件 .py/.go/.java)

总关联度 = max(所有文件得分) + 提交数量加成 + 严重性加成
```

**关键方法**:
```go
// 注册服务映射
RegisterServiceMapping(mapping *ServiceCodeMapping)

// 分析告警
AnalyzeAlert(ctx, alert) (*CodeChangeCorrelation, error)

// 排序可疑文件
rankSuspiciousFiles(commits, mapping, alert) []SuspiciousFile

// 计算关联分数
calculateCorrelationScore(files, commits, alert) float64
```

### 4.3 AlertManager (告警管理器)

**文件**: `internal/monitor/alert_manager.go`

**职责**:
- 接收告警事件
- 触发代码关联分析（异步）
- 生成用户友好的通知消息
- 推送到前端

**关键方法**:
```go
// 处理告警
HandleAlert(ctx, alert) error

// 后台分析（goroutine）
analyzeCodeCorrelation(ctx, alert)

// 构建通知
buildCorrelationNotification(correlation) map[string]interface{}

// 生成消息
generateCorrelationMessage(correlation) string
```

**消息生成逻辑**:
```
IF 时间差 < 5分钟: "刚刚"
ELSE IF 时间差 < 15分钟: "10分钟前"
ELSE: "30分钟内"

IF 提交数 == 1:
  "指标异常！我发现{时间}有一条新的代码合并，改动了 {文件名}。
   这次指标波动极大概率是由代码变更引起的，是否查看改动对比？"
ELSE:
  "指标异常！我发现{时间}有 {N} 条代码合并，其中 {文件名} 的改动最可疑。
   这次指标波动极大概率是由代码变更引起的，是否查看改动对比？"
```

### 4.4 AlertNotifier (通知器扩展)

**文件**: `internal/monitor/notifier.go`

**新增方法**:
```go
// 推送代码关联通知
NotifyCodeCorrelation(correlation interface{}) error
```

**WebSocket 消息格式**:
```json
{
  "type": "code_correlation",
  "data": {
    "alert_id": "alert-20260103120000",
    "alert_name": "high_database_latency",
    "service": "database",
    "correlation_score": 0.85,
    "message": "指标异常！我发现10分钟前有一条新的代码合并...",
    "commits": [...],
    "suspicious_files": [
      {
        "file_path": "src/db/db_query.py",
        "commit_hash": "abc123def456",
        "score": 0.85,
        "reason": "Recent change, matches service path, relevant commit message"
      }
    ],
    "time_window": {
      "alert_time": "2026-01-03T12:00:00Z",
      "search_start": "2026-01-03T11:30:00Z",
      "search_end": "2026-01-03T12:00:00Z"
    }
  }
}
```

## 5. 前端组件设计

### 5.1 CodeCorrelationCard.vue

**文件**: `web/src/components/monitor/CodeCorrelationCard.vue`

**功能**:
- 显示代码关联分析结果
- 展示可疑文件列表（按得分排序）
- 显示相关提交信息
- 提供查看 diff 和忽略操作

**UI 结构**:
```
┌─────────────────────────────────────────┐
│ 🔍 代码关联分析    关联度: 85%  [×]     │
├─────────────────────────────────────────┤
│ high_database_latency  [database]       │
│                                         │
│ ⚠️ 指标异常！我发现10分钟前有一条新的    │
│    代码合并，改动了 db_query.py...      │
│                                         │
│ 可疑文件变更:                           │
│ ┌─────────────────────────────────┐    │
│ │ src/db/db_query.py         85%  │    │
│ │ abc123d • Recent change...      │    │
│ └─────────────────────────────────┘    │
│                                         │
│ 相关提交 (1):                           │
│ ┌─────────────────────────────────┐    │
│ │ John Doe        10分钟前         │    │
│ │ Optimize database query          │    │
│ │ 3 个文件变更                     │    │
│ └─────────────────────────────────┘    │
│                                         │
│ [查看所有改动]  [忽略]                  │
└─────────────────────────────────────────┘
```

**关键特性**:
- 动画滑入效果
- 关联度颜色编码（高/中/低）
- 点击文件打开 DiffViewer
- 自动格式化时间显示

### 5.2 DiffViewer.vue

**文件**: `web/src/components/monitor/DiffViewer.vue`

**功能**:
- 全屏模态框显示 diff
- 语法高亮（可选）
- 复制 diff 内容
- 跳转到 Git 托管平台

**UI 结构**:
```
┌──���──────────────────────────────────────────────┐
│ src/db/db_query.py                         [×]  │
│ abc123d  [backend-api]                          │
├─────────────────────────────────────────────────┤
│                                                 │
│ - def query_users(limit=100):                   │
│ + def query_users(limit=100, use_index=True):   │
│ -     return db.execute("SELECT * FROM users")  │
│ + if use_index:                                 │
│ +     return db.execute("SELECT * FROM users    │
│ +                        USE INDEX (idx_user)") │
│                                                 │
├─────────────────────────────────────────────────┤
│                    [复制 Diff]  [在 Git 中查看] │
└─────────────────────────────────────────────────┘
```

## 6. 配置文件设计

**文件**: `config/code_correlation.yaml`

```yaml
repositories:
  - name: "backend-api"
    path: "/path/to/backend-api"
    enabled: true

service_mappings:
  - service_name: "database"
    repositories: ["backend-api"]
    code_paths:
      - "**/*db*.py"
      - "**/*query*.go"
    keywords:
      - "database"
      - "query"
      - "sql"

correlation:
  time_window_minutes: 30
  min_score_threshold: 0.3
  max_suspicious_files: 5
```

## 7. API 接口设计

### 7.1 WebSocket 接口

**Endpoint**: `GET /ws/monitor`

**消息类型**:
- `alert`: 基础告警通知
- `metrics`: 指标数据
- `code_correlation`: 代码关联分析结果

### 7.2 HTTP 接口

**获取 Diff**:
```
GET /api/git/:repository/diff/:commit?file=path/to/file
Response: { "diff": "...", "repository": "...", "commit": "..." }
```

**查询提交**:
```
POST /api/git/commits
Body: { "repository": "...", "start_time": "...", "end_time": "..." }
Response: { "commits": [...] }
```

## 8. 部署集成

### 8.1 依赖安装

```bash
go get github.com/go-git/go-git/v5
```

### 8.2 初始化流程

```go
// 1. 创建 Git Tracker
gitTracker := builtin.NewGitChangeTracker("")
gitTracker.AddRepository("backend-api", "/path/to/repo")

// 2. 创建 Notifier
notifier := monitor.NewAlertNotifier(logger)

// 3. 创建 AlertManager
alertManager := monitor.NewAlertManager(notifier, gitTracker, logger)

// 4. 注册服务映射
alertManager.RegisterServiceMapping(&model.ServiceCodeMapping{
    ServiceName: "database",
    Repositories: []string{"backend-api"},
    CodePaths: []string{"**/*db*.py"},
    Keywords: []string{"database", "query"},
})

// 5. 处理告警
alertManager.HandleAlert(ctx, alert)
```

### 8.3 与 Agent 系统集成

```go
// 注册为 Agent Tool
gitTool := builtin.NewGitChangeTrackerTool("/repos")
agent.RegisterTool(gitTool)

// Agent 可以调用
result := agent.Execute("git_change_tracker", map[string]interface{}{
    "repository": "backend-api",
    "alert_time": "2026-01-03T12:00:00Z",
    "time_range_minutes": 30,
})
```

## 9. 性能优化

### 9.1 异步处理
- 代码关联分析在 goroutine 中执行，不阻塞告警通知
- 超时控制：30秒内必须完成分析

### 9.2 缓存策略
- Git 仓库对象缓存（避免重复打开）
- Commit 对象缓存（LRU，最多1000个）

### 9.3 并发控制
- 使用 worker pool 限制并发分析数量
- 每个仓库独立的 mutex 锁

## 10. 监控指标

```
# 关联分析性能
code_correlation_analysis_duration_seconds
code_correlation_analysis_total
code_correlation_analysis_errors_total

# 关联结果
code_correlation_score_distribution
code_correlation_notifications_sent_total
code_correlation_commits_analyzed_total
```

## 11. 测试策略

### 11.1 单元测试
- GitChangeTracker: 模拟 Git 仓库
- Correlator: 测试评分算法
- AlertManager: 测试消息生成

### 11.2 集成测试
- 端到端流程：Alert → Analysis → Notification
- WebSocket 消息验证
- 多仓库场景测试

### 11.3 性能测试
- 1000+ commits 的查询性能
- 并发告警处理能力
- 内存占用监控

## 12. 文件清单

### 后端文件
- `internal/model/git_change.go` - 数据模型
- `internal/tool/builtin/git_change_tracker.go` - Git 追踪器
- `internal/monitor/correlator.go` - 关联分析器
- `internal/monitor/alert_manager.go` - 告警管理器
- `internal/monitor/notifier.go` - 通知器（已扩展）
- `internal/api/handler/git_change.go` - HTTP 处理器
- `config/code_correlation.yaml` - 配置文件
- `cmd/server/main_with_code_correlation.go.example` - 集成示例

### 前端文件
- `web/src/components/monitor/CodeCorrelationCard.vue` - 关联卡片
- `web/src/components/monitor/DiffViewer.vue` - Diff 查看器
- `web/src/api/monitor.ts` - API 接口

## 13. 后续优化方向

1. **AI 增强**:
   - 使用 LLM 分析 commit message 和 diff 内容
   - 自动生成根因分析报告

2. **历史学习**:
   - 记录历史关联结果的准确性
   - 基于反馈优化评分算法

3. **多维度关联**:
   - 关联 CI/CD 部署记录
   - 关联配置变更（ConfigMap、环境变量）
   - 关联依赖库更新

4. **可视化增强**:
   - 时间线视图：告警 + 提交 + 部署
   - 影响范围分析图
   - 代码热力图

---

**文档版本**: 1.0
**创建日期**: 2026-01-03
**作者**: AI-Ops Team
