# AI-Ops 智能运维功能设计文档

## 概述

本文档详细描述了 AI-Ops 平台四个核心智能运维功能的技术设计方案，由多个专业 Agent 协作完成分析和设计。

---

## 功能一：故障复盘"时光机" (Incident Post-mortem)

### 1.1 功能概述

用户在对话框输入"复盘昨晚 2 点的 CPU 飙升"，AI 自动复原当时的现场，生成包含"异常起始时间、受影响服务、关键指标拐点、推测根因"的复盘报告。

### 1.2 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户输入                                  │
│              "复盘昨晚2点的CPU飙升"                              │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Agent ReAct 循环                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ 意图识别    │→ │ 工具选择    │→ │ incident_postmortem     │  │
│  └─────────────┘  └─────────────┘  └───────────┬─────────────┘  │
└───────────────────────────────────────────────┼─────────────────┘
                                                │
                      ┌─────────────────────────┼─────────────────┐
                      │                         ▼                 │
                      │  ┌─────────────────────────────────────┐  │
                      │  │      时间解析器 (TimeParser)        │  │
                      │  │   "昨晚2点" → 2026-01-03 02:00      │  │
                      │  └─────────────────┬───────────────────┘  │
                      │                    │                      │
                      │                    ▼                      │
                      │  ┌─────────────────────────────────────┐  │
                      │  │      数据采集器 (DataCollector)     │  │
                      │  │  ┌───────────┬───────────┬───────┐  │  │
                      │  │  │Prometheus │  Logs     │Config │  │  │
                      │  │  │  Metrics  │  Events   │Changes│  │  │
                      │  │  └───────────┴───────────┴───────┘  │  │
                      │  └─────────────────┬───────────────────┘  │
                      │                    │                      │
                      │                    ▼                      │
                      │  ┌─────────────────────────────────────┐  │
                      │  │      分析引擎 (AnalysisEngine)      │  │
                      │  │  • 拐点检测                         │  │
                      │  │  • 异常关联                         │  │
                      │  │  • 根因推断                         │  │
                      │  └─────────────────┬───────────────────┘  │
                      │                    │                      │
                      │                    ▼                      │
                      │  ┌─────────────────────────────────────┐  │
                      │  │      报告生成器 (ReportGenerator)   │  │
                      │  └─────────────────────────────────────┘  │
                      │           incident_postmortem Tool        │
                      └───────────────────────────────────────────┘
```

### 1.3 核心代码结构

```
internal/
├── postmortem/
│   ├── timeparser.go      # 自然语言时间解析
│   ├── collector.go       # 数据采集器
│   ├── analyzer.go        # 分析引擎
│   ├── report.go          # 报告生成
│   └── types.go           # 类型定义
├── prometheus/
│   ├── client.go          # Prometheus 客户端
│   ├── incident_replay.go # 故障复盘查询
│   └── capacity_planning.go # 容量规划
└── tool/builtin/
    └── prometheus_tool.go # Prometheus 工具集成
```

### 1.4 自然语言时间解析

支持的时间表达式：
- 相对日期：今天、昨天、前天、大前天
- 时间段：今晚、昨晚、今早、昨早
- 星期：周一到周日、上周X
- 具体时间：2点、14:30、2点半

```go
// 示例：解析 "昨晚2点"
parser := NewTimeParser()
t, _ := parser.Parse("昨晚2点")
// 结果: 2026-01-02 02:00:00
```

### 1.5 关键 PromQL 查询

```promql
# CPU 使用率
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# 内存使用率
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# 磁盘 I/O
rate(node_disk_io_time_seconds_total[5m])

# 网络流量
rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])

# 负载平均值
node_load1
```

### 1.6 拐点检测算法

使用滑动窗口检测指标突变：

```go
func detectInflectionPoints(metrics []MetricSeries) []InflectionPoint {
    windowSize := 5
    for i := windowSize; i < len(dataPoints)-windowSize; i++ {
        beforeAvg := average(dataPoints[i-windowSize : i])
        afterAvg := average(dataPoints[i : i+windowSize])
        changeRate := (afterAvg - beforeAvg) / beforeAvg * 100

        // 变化超过 50% 视为拐点
        if math.Abs(changeRate) > 50 {
            // 记录拐点
        }
    }
}
```

### 1.7 API 接口

```
POST /api/prometheus/incident-replay
{
    "incident_time": "2026-01-03T02:00:00Z",
    "host": "192.168.1.100",
    "lookback_minutes": 15
}

Response:
{
    "incident_time": "2026-01-03T02:00:00Z",
    "affected_metrics": ["cpu_usage", "memory_usage"],
    "root_causes": ["CPU 使用率异常升高"],
    "recommendations": ["检查高 CPU 进程，考虑优化代码或扩容"],
    "timeline": {...}
}
```

---

## 功能二：交互式容量规划 (What-if Analysis)

### 2.1 功能概述

用户问："如果下周用户量翻倍，这台机器扛得住吗？"，AI 调取历史数据，给出模拟推演结果和扩容建议。

### 2.2 核心算法

#### 线性回归趋势分析

```go
func linearRegression(points []TrendPoint) (slope, intercept float64) {
    n := float64(len(points))
    var sumX, sumY, sumXY, sumX2 float64
    for i, point := range points {
        x := float64(i)
        y := point.Value
        sumX += x
        sumY += y
        sumXY += x * y
        sumX2 += x * x
    }
    slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
    intercept = (sumY - slope*sumX) / n
    return
}
```

#### 指数平滑预测

```go
func ExponentialSmoothing(points []TrendPoint, alpha float64, periods int) []TrendPoint {
    s := points[0].Value
    for i := 0; i < len(points); i++ {
        s = alpha*points[i].Value + (1-alpha)*s
        points[i].Predicted = s
    }
    // 预测未来周期...
}
```

### 2.3 What-if 场景分析

| 场景 | 增长倍数 | 峰值倍数 | 建议行动 |
|------|----------|----------|----------|
| 正常增长 | 1.0x | 1.0x | 监控趋势 |
| 加速增长 | 1.5x | 1.2x | 紧急扩容计划 |
| 业务高峰 | 2.0x | 2.0x | 立即启动应急扩容 |

### 2.4 API 接口

```
POST /api/prometheus/capacity-planning
{
    "metric_query": "node_cpu_seconds_total",
    "days": 30,
    "capacity_threshold": 90,
    "current_capacity": 100
}

Response:
{
    "trend": {
        "type": "increasing",
        "growth_rate": 0.5,
        "projected_peak": 85.2,
        "days_to_capacity": 15
    },
    "scenarios": [...],
    "recommendation": {
        "priority": "medium",
        "action_items": ["计划扩容，安全余量不足20%"]
    }
}
```

---

## 功能三：运维知识库闭环 (RAG + Ops)

### 3.1 功能概述

将公司 SOP 手册、历史故障处理记录喂给 AI，当告警发生时，AI 能够基于历史经验给出处理建议。

### 3.2 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     RAG 知识库系统                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  文档导入   │    │  Embedding  │    │    向量数据库       │  │
│  │  Pipeline   │ →  │   生成器    │ →  │  (Chromem/Milvus)   │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│        ↑                                        ↓               │
│  ┌─────────────┐                       ┌─────────────────────┐  │
│  │ SOP 文档    │                       │    相似度检索       │  │
│  │ Wiki 页面   │                       │    Top-K 召回       │  │
│  │ 故障记录    │                       └─────────────────────┘  │
│  └─────────────┘                                ↓               │
│                                        ┌─────────────────────┐  │
│                                        │    重排序 Rerank    │  │
│                                        └─────────────────────┘  │
│                                                 ↓               │
│                                        ┌─────────────────────┐  │
│                                        │   LLM 生成回答      │  │
│                                        └─────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 代码结构

```
internal/
├── rag/
│   ├── knowledge_base.go   # 知识库核心
│   ├── embedder.go         # Embedding 生成
│   ├── retriever.go        # 检索器
│   ├── reranker.go         # 重排序
│   ├── chunker.go          # 文档分块
│   └── types.go            # 类型定义
├── vectordb/
│   ├── chromem.go          # Chromem 适配器
│   ├── milvus.go           # Milvus 适配器
│   └── interface.go        # 向量库接口
```

### 3.4 向量数据库选型

| 特性 | Chromem-go | Milvus | Qdrant |
|------|------------|--------|--------|
| 部署复杂度 | 低（嵌入式） | 高 | 中 |
| 性能 | 中 | 高 | 高 |
| 适用场景 | 小规模 | 大规模 | 中大规模 |
| 推荐 | ✅ 初期 | 生产环境 | 备选 |

### 3.5 文档分块策略

```go
type ChunkConfig struct {
    ChunkSize    int  // 默认 512 tokens
    ChunkOverlap int  // 默认 50 tokens
    Separator    string
}

func ChunkDocument(doc Document, config ChunkConfig) []Chunk {
    // 1. 按段落分割
    // 2. 合并小段落
    // 3. 拆分大段落
    // 4. 添加重叠
}
```

### 3.6 检索流程

```go
func (kb *KnowledgeBase) Search(query string, topK int) ([]Document, error) {
    // 1. Query 改写（可选）
    rewrittenQuery := kb.rewriter.Rewrite(query)

    // 2. 生成 Embedding
    embedding := kb.embedder.Embed(rewrittenQuery)

    // 3. 向量检索
    candidates := kb.vectorDB.Search(embedding, topK*3)

    // 4. 关键词检索（混合检索）
    keywordResults := kb.keywordSearch(query, topK*3)

    // 5. 合并去重
    merged := merge(candidates, keywordResults)

    // 6. 重排序
    reranked := kb.reranker.Rerank(query, merged)

    return reranked[:topK], nil
}
```

### 3.7 与告警系统集成

```go
func (am *AlertManager) HandleAlert(alert *Alert) {
    // 1. 发送基础告警
    am.notifier.Notify(alert)

    // 2. 检索相关知识
    query := fmt.Sprintf("%s %s 处理方法", alert.Service, alert.Name)
    docs, _ := am.knowledgeBase.Search(query, 3)

    // 3. 生成处理建议
    if len(docs) > 0 {
        suggestion := am.generateSuggestion(alert, docs)
        am.notifier.NotifyWithSuggestion(alert, suggestion)
    }
}
```

---

## 功能四：代码级关联诊断 (Linkage to Code)

### 4.1 功能概述

AI 监控到异常后，自动查询 Git 提交记录，主动弹出："指标异常！我发现 10 分钟前有一条新的代码合并，改动了 db_query.py。这次指标波动极大概率是由代码变更引起的，是否查看改动对比？"

### 4.2 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    代码关联诊���系统                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  告警触发   │ →  │ Git Tracker │ →  │   关联分析引擎      │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│                                                 ↓               │
│                                        ┌─────────────────────┐  │
│                                        │   可疑文件排序      │  │
│                                        └─────────────────────┘  │
│                                                 ↓               │
│                                        ┌─────────────────────┐  │
│                                        │  WebSocket 推送     │  │
│                                        └─────────────────────┘  │
│                                                 ↓               │
│                                        ┌─────────────────────┐  │
│                                        │  前端弹窗展示       │  │
│                                        └─────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 4.3 代码结构

```
internal/
├── tool/builtin/
│   └── git_change_tracker.go  # Git 变更追踪工具
├── monitor/
│   ├── correlator.go          # 代码关联分析器
│   └── alert_manager.go       # 告警管理器
├── model/
│   └── code_correlation.go    # 关联模型定义
├── api/handler/
│   └── git_change.go          # Git API 处理器

web/src/
├── components/monitor/
│   ├── CodeCorrelationCard.vue  # 关联卡片组件
│   └── DiffViewer.vue           # Diff 查看器
├── api/
│   └── monitor.ts               # 监控 API
```

### 4.4 服务-代码映射配置

```yaml
# config/code_correlation.yaml
service_mappings:
  - service_name: "api-gateway"
    repositories:
      - "backend-api"
    code_paths:
      - "src/gateway/**"
      - "src/middleware/**"
    keywords:
      - "gateway"
      - "routing"

  - service_name: "database"
    repositories:
      - "backend-api"
      - "data-pipeline"
    code_paths:
      - "**/*db*.py"
      - "**/*query*.go"
    keywords:
      - "database"
      - "query"
      - "sql"
```

### 4.5 关联度评分算法

```go
func rankSuspiciousFiles(commits []GitCommit, alert *Alert) []SuspiciousFile {
    for _, commit := range commits {
        // 时间因子：越接近告警时间分数越高
        timeDelta := alert.Timestamp.Sub(commit.Timestamp).Minutes()
        timeScore := 1.0 - (timeDelta / 30.0)

        for _, file := range commit.ChangedFiles {
            score := timeScore

            // 路径匹配加分
            if matchesCodePaths(file, mapping.CodePaths) {
                score += 0.3
            }

            // 关键词匹配加分
            if containsKeywords(commit.Message, mapping.Keywords) {
                score += 0.2
            }

            // 告警上下文匹配加分
            if containsKeywords(file, extractKeywordsFromAlert(alert)) {
                score += 0.3
            }
        }
    }
    // 按分数排序返回 Top 5
}
```

### 4.6 WebSocket 消息格式

```json
{
    "type": "code_correlation",
    "data": {
        "alert_id": "alert-20260103020000",
        "alert_name": "high_cpu_usage",
        "service": "api-gateway",
        "correlation_score": 0.85,
        "message": "指标异常！我发现10分钟前有一条新的代码合并，改动了 db_query.py。这次指标波动极大概率是由代码变更引起的，是否查看改动对比？",
        "commits": [...],
        "suspicious_files": [
            {
                "file_path": "src/db/query.py",
                "commit_hash": "abc1234",
                "score": 0.85,
                "reason": "Recent change, matches service path"
            }
        ]
    }
}
```

### 4.7 前端组件

CodeCorrelationCard.vue 主要功能：
- 显示关联度评分（高/中/低）
- 展示可疑文件列表
- 显示相关提交信息
- 支持查看 Diff 对比
- 支持忽略/确认操作

---

## 实现优先级

| 优先级 | 功能 | 依赖 | 预估工作量 |
|--------|------|------|------------|
| P0 | Prometheus 集成 | 无 | 基础设施 |
| P1 | 故障复盘时光机 | Prometheus | 核心功能 |
| P1 | 代码关联诊断 | Git API | 高价值 |
| P2 | 容量规划 | Prometheus + 历史数据 | 需要数据积累 |
| P3 | RAG 知识库 | 向量数据库 + 文档整理 | 需要内容准备 |

---

## 已创建的文件清单

### 后端代码
- `internal/prometheus/client.go` - Prometheus 客户端
- `internal/prometheus/incident_replay.go` - 故障复盘查询
- `internal/prometheus/capacity_planning.go` - 容量规划
- `internal/tool/builtin/prometheus_tool.go` - Prometheus 工具
- `internal/tool/builtin/git_change_tracker.go` - Git 变更追踪
- `internal/monitor/correlator.go` - 代码关联分析器
- `internal/monitor/alert_manager.go` - 告警管理器
- `internal/api/handler/prometheus.go` - Prometheus API
- `internal/api/handler/git_change.go` - Git API
- `internal/rag/knowledge_base.go` - RAG 知识库
- `internal/rag/embedder.go` - Embedding 生成器
- `internal/rag/retriever.go` - 检索器
- `internal/rag/types.go` - RAG 类型定义
- `internal/model/code_correlation.go` - 代码关联模型

### 前端代码
- `web/src/components/monitor/CodeCorrelationCard.vue` - 关联卡片
- `web/src/components/monitor/DiffViewer.vue` - Diff 查看器
- `web/src/api/monitor.ts` - 监控 API

### 配置文件
- `config/code_correlation.yaml` - 代码关联配置

---

## 下一步行动

1. **集成 Prometheus** - 配置 Prometheus 地址，测试连接
2. **完善时间解析** - 支持更多自然语言时间表达式
3. **配置服务映射** - 根据实际项目配置服务与代码的映射关系
4. **准备知识库文档** - 整理 SOP、故障处理记录等文档
5. **前端集成** - 将组件集成到主界面

---

*文档生成时间: 2026-01-03*
*由 AI-Ops 专业 Agent 协作完成*
