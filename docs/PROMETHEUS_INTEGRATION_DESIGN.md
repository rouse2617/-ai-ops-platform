# AI-Ops Prometheus 集成架构设计文档

## 1. 架构概览

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                     AI-Ops 运维平台                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              Prometheus 集成层 (prometheus)              │   │
│  ├──────────────────────────────────────────────────────────┤   │
│  │                                                            │   │
│  │  ┌─────────────────┐  ┌──────────────────┐              │   │
│  │  │ PrometheusClient│  │ MetricCollector  │              │   │
│  │  │ (API 交互)      │  │ (指标采集)       │              │   │
│  │  └─────────────────┘  └──────────────────┘              │   │
│  │                                                            │   │
│  │  ┌─────────────────┐  ┌──────────────────┐              │   │
│  │  │ PromQLGenerator │  │ SemanticEngine   │              │   │
│  │  │ (PromQL生成)    │  │ (语义翻译)       │              │   │
│  │  └─────────────────┘  └──────────────────┘              │   │
│  │                                                            │   │
│  │  ┌─────────────────┐  ┌──────────────────┐              │   │
│  │  │ ThresholdDetector│ │ PlaybookEngine   │              │   │
│  │  │ (动态阈值)      │  │ (自愈剧本)       │              │   │
│  │  └─────────────────┘  └──────────────────┘              │   │
│  │                                                            │   │
│  │  ┌─────────────────────────────────────────┐            │   │
│  │  │    DynamicDashboard (千人千面看板)      │            │   │
│  │  └─────────────────────────────────────────┘            │   │
│  │                                                            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           ▲                                      │
│                           │                                      │
│  ┌────────────────────────┴──────────────────────────────────┐  │
│  │              API Handler 层 (handler)                      │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  PrometheusHandler  │  MetricsHandler  │  DashboardHandler  │
│  └──────────────────────────────────────────────────────────┘  │
│                           ▲                                      │
│                           │                                      │
│  ┌────────────────────────┴──────────────────────────────────┐  │
│  │              LLM 分析层 (agent)                            │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  SemanticAnalyzer  │  PlaybookGenerator  │  Correlator     │
│  └──────────────────────────────────────────────────────────┘  │
│                           ▲                                      │
│                           │                                      │
└───────────────────────────┼──────────────────────────────────────┘
                            │
                    ┌───────┴────────┐
                    │                │
            ┌───────▼────────┐  ┌────▼──────────┐
            │  Prometheus    │  │  Grafana      │
            │  Server        │  │  Dashboard    │
            └────────────────┘  └───────────────┘
```

### 1.2 核心模块

| 模块 | 功能 | 文件 |
|------|------|------|
| Prometheus-1 | 故障语义化翻译 (Metric → NL) | `prometheus/semantic.go` |
| Prometheus-2 | 对话式 PromQL (NL → PromQL) | `prometheus/promql_generator.go` |
| Prometheus-3 | 动态阈值检测 | `prometheus/threshold_detector.go` |
| Prometheus-4 | 自愈排障剧本 | `prometheus/playbook_engine.go` |
| Prometheus-5 | 动态看板 | `prometheus/dashboard_generator.go` |

---

## 2. 详细架构设计

### 2.1 Prometheus-1: 故障语义化翻译

**目标**: 将 Prometheus 指标转换为自然语言描述

**核心功能**:
- 指标采集与解析
- 异常检测与分类
- 自然语言生成
- 上下文感知的描述

**数据流**:
```
Prometheus Metrics → MetricParser → AnomalyDetector → SemanticEngine → Natural Language
```

**关键接口**:
```go
type MetricEvent struct {
    MetricName string
    Labels     map[string]string
    Value      float64
    Timestamp  time.Time
    Severity   AlertLevel
}

type SemanticTranslation struct {
    Title       string
    Description string
    RootCause   string
    Impact      string
    Suggestions []string
}
```

### 2.2 Prometheus-2: 对话式 PromQL

**目标**: 将自然语言查询转换为 PromQL

**核心功能**:
- 自然语言理解
- PromQL 语法生成
- 查询优化
- 结果解释

**示例对话**:
```
用户: "最近一小时内存使用率超过80%的主机有哪些?"
系统:
  PromQL: node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes < 0.2
  时间范围: 1h
  结果: [host1, host2, host3]
```

**关键接口**:
```go
type NLQuery struct {
    Question string
    Context  map[string]interface{}
}

type PromQLQuery struct {
    Query      string
    Range      string
    Step       string
    Explanation string
}
```

### 2.3 Prometheus-3: 动态阈值检测

**目标**: 基于历史基线的异常检测

**核心功能**:
- 历史基线学习
- 动态阈值计算
- 异常检测算法
- 趋势分析

**算法**:
```
动态阈值 = 历史平均值 + (标准差 × 敏感度系数)
异常判定 = |当前值 - 历史平均值| > 动态阈值
```

**关键接口**:
```go
type BaselineMetrics struct {
    MetricName string
    Mean       float64
    StdDev     float64
    P95        float64
    P99        float64
}

type AnomalyResult struct {
    IsAnomaly      bool
    Severity       float64
    Deviation      float64
    Recommendation string
}
```

### 2.4 Prometheus-4: 自愈排障剧本

**目标**: AI 驱动的自动化排障和修复

**核心功能**:
- 故障诊断
- 剧本生成
- 自动执行
- 结果验证

**剧本结构**:
```yaml
playbook:
  name: "高CPU使用率处理"
  trigger: "cpu_usage > 80%"
  steps:
    - name: "诊断"
      actions:
        - "ps aux --sort=-%cpu | head -10"
        - "top -bn1 | head -20"
    - name: "分析"
      ai_analysis: true
    - name: "修复"
      actions:
        - "kill -9 <pid>"
      requires_approval: true
    - name: "验证"
      check: "cpu_usage < 50%"
```

**关键接口**:
```go
type Playbook struct {
    ID          string
    Name        string
    Trigger     string
    Steps       []PlaybookStep
    AutoExecute bool
}

type PlaybookStep struct {
    Name            string
    Actions         []string
    AIAnalysis      bool
    RequiresApproval bool
}
```

### 2.5 Prometheus-5: 动态看板

**目标**: 千人千面的个性化 Dashboard

**核心功能**:
- 用户偏好学习
- 动态面板生成
- 实时数据更新
- 交���式探索

**看板类型**:
- 系统概览看板
- 应用性能看板
- 业务指标看板
- 告警看板
- 自定义看板

**关键接口**:
```go
type DashboardConfig struct {
    UserID    string
    Name      string
    Panels    []PanelConfig
    Refresh   string
    TimeRange string
}

type PanelConfig struct {
    ID       string
    Title    string
    Type     string // graph, stat, table, heatmap
    Query    string
    Options  map[string]interface{}
}
```

---

## 3. Prometheus API 集成方式

### 3.1 HTTP API 端点

```
GET  /api/v1/query
GET  /api/v1/query_range
GET  /api/v1/series
GET  /api/v1/label/{label_name}/values
GET  /api/v1/targets
GET  /api/v1/rules
POST /api/v1/query
```

### 3.2 客户端配置

```go
type PrometheusConfig struct {
    URL            string
    Timeout        time.Duration
    BasicAuth      *BasicAuth
    TLSConfig      *tls.Config
    RetryPolicy    *RetryPolicy
}

type BasicAuth struct {
    Username string
    Password string
}
```

### 3.3 连接管理

- 连接池管理
- 自动重试机制
- 超时控制
- 错误恢复

---

## 4. 核心数据结构

### 4.1 指标数据结构

```go
// 指标样本
type MetricSample struct {
    Metric    map[string]string
    Value     [2]interface{} // [timestamp, value]
    Timestamp int64
}

// 指标范围查询结果
type RangeQueryResult struct {
    ResultType string
    Result     []RangeSeries
}

type RangeSeries struct {
    Metric map[string]string
    Values [][2]interface{}
}

// 指标即时查询结果
type InstantQueryResult struct {
    ResultType string
    Result     []MetricSample
}
```

### 4.2 告警数据结构

```go
type AlertMetadata struct {
    AlertName  string
    Severity   string
    Description string
    Runbook    string
    Dashboard  string
}

type AlertContext struct {
    Metric     MetricSample
    Baseline   BaselineMetrics
    Trend      TrendAnalysis
    Related    []RelatedMetric
}
```

### 4.3 看板数据结构

```go
type DashboardMetadata struct {
    ID          string
    Title       string
    Description string
    Tags        []string
    Panels      []PanelMetadata
}

type PanelMetadata struct {
    ID          string
    Title       string
    Type        string
    DataSource  string
    Targets     []Target
    Options     PanelOptions
}
```

---

## 5. 接口定义

### 5.1 Prometheus 客户端接口

```go
type PrometheusClient interface {
    // 即时查询
    Query(ctx context.Context, query string) (interface{}, error)

    // 范围查询
    QueryRange(ctx context.Context, query string, r Range) (interface{}, error)

    // 获取标签值
    LabelValues(ctx context.Context, label string) ([]string, error)

    // 获取系列
    Series(ctx context.Context, matches []string) ([]map[string]string, error)

    // 获取告警规则
    Rules(ctx context.Context) (interface{}, error)
}
```

### 5.2 语义引擎接口

```go
type SemanticEngine interface {
    // 翻译指标为自然语言
    TranslateMetric(metric MetricEvent) (SemanticTranslation, error)

    // 生成告警描述
    GenerateAlertDescription(alert AlertMetadata, context AlertContext) (string, error)

    // 生成建议
    GenerateRecommendations(anomaly AnomalyResult) ([]string, error)
}
```

### 5.3 PromQL 生成器接口

```go
type PromQLGenerator interface {
    // 自然语言转 PromQL
    GenerateQuery(nl NLQuery) (PromQLQuery, error)

    // 优化 PromQL
    OptimizeQuery(query string) (string, error)

    // 解释 PromQL
    ExplainQuery(query string) (string, error)
}
```

### 5.4 阈值检测器接口

```go
type ThresholdDetector interface {
    // 学习基线
    LearnBaseline(metrics []MetricSample) (BaselineMetrics, error)

    // 检测异常
    DetectAnomaly(current MetricSample, baseline BaselineMetrics) (AnomalyResult, error)

    // 更新动态阈值
    UpdateThreshold(metric string, baseline BaselineMetrics) error
}
```

### 5.5 剧本引擎接口

```go
type PlaybookEngine interface {
    // 创建剧本
    CreatePlaybook(playbook Playbook) error

    // 执行剧本
    ExecutePlaybook(playbookID string, context map[string]interface{}) (ExecutionResult, error)

    // 获取剧本
    GetPlaybook(playbookID string) (Playbook, error)

    // 列表剧本
    ListPlaybooks(filter PlaybookFilter) ([]Playbook, error)
}
```

### 5.6 看板生成器接口

```go
type DashboardGenerator interface {
    // 生成看板
    GenerateDashboard(userID string, config DashboardConfig) (Dashboard, error)

    // 更新看板
    UpdateDashboard(dashboardID string, config DashboardConfig) error

    // 获取看板
    GetDashboard(dashboardID string) (Dashboard, error)

    // 推荐面板
    RecommendPanels(userID string) ([]PanelConfig, error)
}
```

---

## 6. 实现方案

### 6.1 文件结构

```
internal/prometheus/
├── client.go              # Prometheus 客户端
├── semantic.go            # 语义翻译引擎
├── promql_generator.go    # PromQL 生成器
├── threshold_detector.go  # 动态阈值检测
├── playbook_engine.go     # 自愈剧本引擎
├── dashboard_generator.go # 动态看板生成
├── models.go              # 数据模型
└── types.go               # 类型定义

internal/api/handler/
├── prometheus.go          # Prometheus API 处理器
├── metrics.go             # 指标处理器
├── dashboard.go           # 看板处理器
└── playbook.go            # 剧本处理器

web/src/
├── api/prometheus.ts      # 前端 API 调用
├── components/
│   ├── PrometheusQuery.vue    # PromQL 查询组件
│   ├── MetricViewer.vue       # 指标查看器
│   ├── DynamicDashboard.vue   # 动态看板
│   └── PlaybookExecutor.vue   # 剧本执行器
└── stores/
    └── prometheus.ts      # 状态管理
```

### 6.2 集成步骤

1. **第一步**: 实现 Prometheus 客户端
2. **第二步**: 实现语义翻译引擎
3. **第三步**: 实现 PromQL 生成器
4. **第四步**: 实现动态阈值检测
5. **第五步**: 实现自愈剧本引擎
6. **第六步**: 实现动态看板生成
7. **第七步**: 实现 API 处理器
8. **第八步**: 实现前端组件

---

## 7. 配置示例

### 7.1 Prometheus 配置

```yaml
prometheus:
  url: "http://localhost:9090"
  timeout: 30s
  basic_auth:
    username: "admin"
    password: "password"
  retry_policy:
    max_retries: 3
    backoff: 1s
```

### 7.2 阈值配置

```yaml
thresholds:
  cpu:
    baseline_window: 7d
    sensitivity: 2.0
    min_threshold: 50
    max_threshold: 95
  memory:
    baseline_window: 7d
    sensitivity: 2.0
    min_threshold: 60
    max_threshold: 90
  disk:
    baseline_window: 30d
    sensitivity: 1.5
    min_threshold: 70
    max_threshold: 95
```

### 7.3 剧本配置

```yaml
playbooks:
  - name: "高CPU处理"
    trigger: "cpu_usage > 80%"
    auto_execute: false
    steps:
      - name: "诊断"
        actions:
          - "ps aux --sort=-%cpu | head -10"
  - name: "高内存处理"
    trigger: "memory_usage > 85%"
    auto_execute: false
    steps:
      - name: "诊断"
        actions:
          - "free -h"
```

---

## 8. 测试用例

### 8.1 单元测试

- PrometheusClient 连接测试
- 语义翻译准确性测试
- PromQL 生成正确性测试
- 阈值检测准确性测试
- 剧本执行流程测试

### 8.2 集成测试

- 端到端查询流程
- 告警触发与处理流程
- 剧本自动执行流程
- 看板动态生成流程

### 8.3 浏览器交互测试

- PromQL 查询界面
- 指标可视化
- 动态看板交互
- 剧本执行界面

---

## 9. 性能指标

| 指标 | 目标 | 说明 |
|------|------|------|
| 查询响应时间 | < 500ms | 单个 PromQL 查询 |
| 语义翻译延迟 | < 1s | 指标转自然语言 |
| 阈值检测延迟 | < 100ms | 异常检测 |
| 剧本执行时间 | < 5s | 自动化修复 |
| 看板加载时间 | < 2s | 动态看板生成 |

---

## 10. 安全考虑

- Prometheus 认证与授权
- 敏感数据脱敏
- 审计日志记录
- 剧本执行权限控制
- API 速率限制

---

## 11. 扩展性设计

- 支持多个 Prometheus 实例
- 支持自定义指标解析器
- 支持自定义剧本模板
- 支持第三方告警源集成
- 支持自定义看板模板

---

## 12. 部署架构

```
┌─────────────────────────────────────────┐
│         AI-Ops 应用服务                  │
│  (Prometheus 集成模块)                   │
└──────────────────┬──────────────────────┘
                   │
        ┌──────────┼──────────┐
        │          │          │
   ┌────▼──┐  ┌───▼───┐  ┌──▼────┐
   │Prom-1 │  │Prom-2 │  │Prom-3 │
   │语义   │  │PromQL │  │阈值   │
   └────┬──┘  └───┬───┘  └──┬────┘
        │         │         │
   ┌────▼─────────▼─────────▼────┐
   │   Prometheus Server          │
   │   (指标存储与查询)            │
   └──────────────────────────────┘
```

---

## 13. 后续优化方向

1. **机器学习集成**: 使用 ML 模型进行更精准的异常检测
2. **多源数据融合**: 集成其他监控系统数据
3. **智能告警聚合**: 基于 AI 的告警关联与聚合
4. **自适应阈值**: 根据业务周期动态调整阈值
5. **预测性告警**: 提前预测潜在故障
6. **自动化修复**: 更多场景的自动化修复支持
7. **知识库集成**: 集成运维知识库与最佳实践
8. **成本优化**: 基于指标的资源成本优化建议

