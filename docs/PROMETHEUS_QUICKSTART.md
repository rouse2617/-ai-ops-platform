# AI-Ops Prometheus 集成 - 快速开始指南

## 项目概述

本指南说明如何在 AI-Ops 项目中集成 Prometheus，实现两个核心功能：
1. **故障复盘时光机** - 查询历史指标数据进行故障分析
2. **容量规划 What-if 分析** - 基于趋势预测进行容量规划

## 已创建的文件清单

### 核心实现文件

#### 1. Prometheus 客户端
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\client.go`
- Prometheus API 客户端封装
- 范围查询 (QueryRange)
- 即时查询 (QueryInstant)
- 元数据和标签值获取

#### 2. 故障复盘模块
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\incident_replay.go`
- 故障时间线查询
- 关键指标采集（CPU、内存、磁盘、网络等）
- 异常检测和基线计算
- 故障报告生成
- 常用 PromQL 查询语句集合

**关键类型**:
```go
type IncidentTimeline struct {
    IncidentTime time.Time
    PreWindow    time.Time
    PostWindow   time.Time
    Metrics      map[string]*MetricTimeSeries
}

type MetricTimeSeries struct {
    MetricName string
    Labels     map[string]string
    Values     []MetricPoint
    Baseline   float64
    Peak       float64
    Anomaly    bool
}
```

#### 3. 容量规划模块
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\capacity_planning.go`
- 历史数据获取和趋势分析
- 线性回归算法
- 指数平滑预测
- What-if 场景分析（正常增长、加速增长、业务高峰）
- 容量耗尽时间预测
- 资源分配策略计算

**关键类型**:
```go
type LoadTrend struct {
    MetricName    string
    DataPoints    []TrendPoint
    Trend         TrendType
    GrowthRate    float64
    ProjectedPeak float64
    DaysToCapacity int
}

type WhatIfScenario struct {
    Name              string
    GrowthMultiplier  float64
    PeakMultiplier    float64
    ProjectedMetrics  map[string]float64
    RecommendedAction string
    TimeToAction      int
}
```

#### 4. 初始化模块
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\init.go`
- Prometheus 集成初始化
- 客户端创建和验证
- 工具注册

### 工具系统集成

#### 5. 内置工具定义
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\tool\builtin\prometheus_tool.go`

**工具1: 故障复盘工具**
```go
Name: "prometheus_incident_replay"
Parameters:
  - incident_time (string, required): RFC3339 格式时间
  - host (string, required): 主机名或 IP
  - lookback_minutes (integer, optional): 回溯时间
```

**工具2: 容量规划工具**
```go
Name: "prometheus_capacity_planning"
Parameters:
  - metric_query (string, required): PromQL 查询语句
  - days (integer, optional): 历史数据天数
  - capacity_threshold (number, optional): 容量阈值
```

### API 接口层

#### 6. HTTP 处理器
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\api\handler\prometheus.go`

**API 端点**:
- `POST /api/prometheus/incident-replay` - 故障复盘查询
- `POST /api/prometheus/capacity-planning` - 容量规划分析
- `POST /api/prometheus/query` - 自定义 PromQL 查询
- `GET /api/prometheus/metrics` - 获取指标元数据
- `GET /api/prometheus/labels/:label` - 获取标签值

#### 7. 路由配置
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\api\router.go` (已更新)
- 添加 `PrometheusAddr` 配置字段
- 注册 Prometheus API 路由组

### 配置文件

#### 8. 配置结构
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\config\prometheus.go`
```go
type PrometheusConfig struct {
    Enabled bool   // 是否启用
    Address string // Prometheus 服务地址
    Timeout int    // 查询超时时间（秒）
}
```

#### 9. 配置示例
**文件**: `C:\Users\hrp\Downloads\ai-pro\docs\prometheus-config-example.yaml`
- AI-Ops 配置示例
- Prometheus 服务器配置
- Recording Rules 配置
- Alert Rules 配置

### 文档

#### 10. 完整集成文档
**文件**: `C:\Users\hrp\Downloads\ai-pro\docs\PROMETHEUS_INTEGRATION.md`
- 详细的功能说明
- PromQL 查询示例
- API 接口文档
- 使用示例
- 性能优化建议
- 故障排查指南

## 快���集成步骤

### 步骤1: 添加依赖

```bash
cd C:\Users\hrp\Downloads\ai-pro
go get github.com/prometheus/client_golang@latest
```

### 步骤2: 更新配置文件

在 `config.yaml` 中添加：

```yaml
prometheus:
  enabled: true
  address: "http://localhost:9090"
  timeout: 30
```

### 步骤3: 初始化 Prometheus 集成

在 `cmd/server/main.go` 中添加初始化代码：

```go
import (
    "ai-ops/internal/prometheus"
    "ai-ops/internal/tool/builtin"
)

// 在主函数中添加
if cfg.Prometheus.Enabled {
    err := builtin.RegisterPrometheusTools(toolRegistry, cfg.Prometheus.Address)
    if err != nil {
        logger.Warn("注册 Prometheus 工具失败", zap.Error(err))
    }
}

// 传递 Prometheus 地址给路由
router := api.NewRouter(api.RouterConfig{
    // ... 其他配置 ...
    PrometheusAddr: cfg.Prometheus.Address,
})
```

### 步骤4: 启动 Prometheus 服务

```bash
# 使用 Docker
docker run -d \
  -p 9090:9090 \
  -v /path/to/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# 或本地安装
prometheus --config.file=prometheus.yml
```

### 步骤5: 配置 Node Exporter

在监控的主机上安装 Node Exporter：

```bash
# Linux
wget https://github.com/prometheus/node_exporter/releases/download/v1.6.1/node_exporter-1.6.1.linux-amd64.tar.gz
tar xvfz node_exporter-1.6.1.linux-amd64.tar.gz
./node_exporter-1.6.1.linux-amd64/node_exporter
```

## 核心 PromQL 查询语句

### CPU 使用率
```promql
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle",instance="<host>"}[5m])) * 100)
```

### 内存使用率
```promql
(1 - (node_memory_MemAvailable_bytes{instance="<host>"} / node_memory_MemTotal_bytes{instance="<host>"})) * 100
```

### 磁盘 I/O
```promql
rate(node_disk_io_time_seconds_total{instance="<host>"}[5m])
```

### 网络流量
```promql
rate(node_network_receive_bytes_total{instance="<host>"}[5m]) + rate(node_network_transmit_bytes_total{instance="<host>"}[5m])
```

### 负载平均值
```promql
node_load1{instance="<host>"}
```

## API 使用示例

### 故障复盘查询

```bash
curl -X POST http://localhost:8080/api/prometheus/incident-replay \
  -H "Content-Type: application/json" \
  -d '{
    "incident_time": "2026-01-03T10:30:00Z",
    "host": "web-01:9100",
    "lookback_minutes": 15
  }'
```

**响应**:
```json
{
  "incident_time": "2026-01-03T10:30:00Z",
  "host": "web-01:9100",
  "affected_metrics": ["node_cpu_seconds_total", "node_memory_MemAvailable_bytes"],
  "root_causes": ["CPU 使用率异常升高", "内存使用率异常升高"],
  "recommendations": ["检查高 CPU 进程", "检查内存泄漏"],
  "metrics_timeline": {
    "node_cpu_seconds_total": {
      "baseline": 45.2,
      "peak": 92.5,
      "anomaly": true,
      "points": 30
    }
  }
}
```

### 容量规划分析

```bash
curl -X POST http://localhost:8080/api/prometheus/capacity-planning \
  -H "Content-Type: application/json" \
  -d '{
    "metric_query": "100 - (avg by (instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
    "days": 30,
    "capacity_threshold": 90.0,
    "current_capacity": 100.0,
    "cost_per_unit": 100.0
  }'
```

**响应**:
```json
{
  "trend": {
    "type": "increasing",
    "growth_rate": 0.5,
    "projected_peak": 85.3,
    "days_to_capacity": 10
  },
  "scenarios": [
    {
      "name": "正常增长",
      "growth_multiplier": 1.0,
      "recommended_action": "立即规划扩容",
      "time_to_action_days": 7
    }
  ],
  "recommendation": {
    "current_usage": 72.5,
    "projected_usage": 85.3,
    "recommended_capacity": 102.4,
    "priority": "medium",
    "action_items": ["计划扩容，安全余量不足20%"]
  }
}
```

## 通过 AI 对话使用

用户可以通过自然语言��� AI 交互：

```
用户: 分析一下 2026-01-03 10:30 时 web-01 主机的故障原因
AI: 我来帮你查询故障时间段的指标数据...
[调用 prometheus_incident_replay 工具]
结果: 发现 CPU 和内存在故障时间段异常升高...

用户: 预测一下 CPU 使用率什么时候会达到容量上限
AI: 我来分析 CPU 的历史趋势...
[调用 prometheus_capacity_planning 工具]
结果: 基于过去 30 天的数据，预计 10 天内达到容量上限...
```

## 关键特性

### 故障复盘时光机
- 查询指定时间范围的历史指标
- 自动检测异常指标
- 计算基线值和峰值
- 生成根因分析报告
- 提供处理建议

### 容量规划 What-if 分析
- 线性回归趋势分析
- 指数平滑预测
- 多场景分析（正常、加速、高峰）
- 容量耗尽时间预测
- 资源分配成本估算
- ROI 计算

## 性能指标

- 查询响应时间: < 5 秒
- 支持最多 30 天历史数据
- 并发查询限制: 10 个
- 缓存 TTL: 5 分钟
- 元数据缓存 TTL: 1 小时

## 下一步

1. 在 `cmd/server/main.go` 中集成初始化代码
2. 更新 `config.yaml` 配置文件
3. 启动 Prometheus 和 Node Exporter
4. 测试 API 端点
5. 在前端集成 Prometheus 分析面板
6. 配置告警规则和通知

## 文件总结

| 文件 | 功能 |
|------|------|
| `internal/prometheus/client.go` | Prometheus 客户端 |
| `internal/prometheus/incident_replay.go` | 故障复盘实现 |
| `internal/prometheus/capacity_planning.go` | 容量规划实现 |
| `internal/prometheus/init.go` | 初始化代码 |
| `internal/tool/builtin/prometheus_tool.go` | 内置工具 |
| `internal/api/handler/prometheus.go` | HTTP 处理器 |
| `internal/api/router.go` | 路由配置 (已更新) |
| `internal/config/prometheus.go` | 配置结构 |
| `docs/PROMETHEUS_INTEGRATION.md` | 完整文档 |
| `docs/prometheus-config-example.yaml` | 配置示例 |

所有文件都已创建完成，可以直接集成到项目中。
