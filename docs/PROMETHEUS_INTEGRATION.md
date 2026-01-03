# AI-Ops Prometheus 集成方案

## 概述

本文档详细说明如何在 AI-Ops 项目中集成 Prometheus，实现故障复盘时光机和容量规划 What-if 分析功能。

## 1. 故障复盘时光机 (Incident Replay)

### 1.1 核心功能

通过 Prometheus API 查询指定时间范围的历史指标数据，实现故障分析和根因诊断。

### 1.2 关键指标采集

```yaml
# 必需的 Node Exporter 指标
- node_cpu_seconds_total          # CPU 使用率
- node_memory_MemAvailable_bytes  # 可用内存
- node_memory_MemTotal_bytes      # 总内存
- node_disk_io_time_seconds_total # 磁盘 I/O 时间
- node_network_receive_bytes_total # 网络接收字节
- node_network_transmit_bytes_total # 网络发送字节
- node_load1                       # 1分钟负载
- node_context_switches_total      # 上下文切换
- node_procs_running              # 运行进程数
- node_filefd_allocated           # 文件描述符
- node_sockstat_TCP_inuse         # TCP 连接数
```

### 1.3 时间窗口查询设计

```
故障前15分钟 ← [故障发生时间] → 恢复后15分钟
|-------------|--------|-------------|
PreWindow    Incident  PostWindow
```

### 1.4 PromQL 查询示例

#### CPU 使用率
```promql
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle",instance="<host>"}[5m])) * 100)
```

#### 内存使用率
```promql
(1 - (node_memory_MemAvailable_bytes{instance="<host>"} / node_memory_MemTotal_bytes{instance="<host>"})) * 100
```

#### 磁盘 I/O
```promql
rate(node_disk_io_time_seconds_total{instance="<host>"}[5m])
```

#### 网络流量
```promql
rate(node_network_receive_bytes_total{instance="<host>"}[5m]) + rate(node_network_transmit_bytes_total{instance="<host>"}[5m])
```

#### 负载平均值
```promql
node_load1{instance="<host>"}
```

### 1.5 API 接口

#### 查询故障时间段指标
```
POST /api/prometheus/incident-replay
Content-Type: application/json

{
  "incident_time": "2026-01-03T10:30:00Z",
  "host": "web-01:9100",
  "lookback_minutes": 15
}
```

**响应示例：**
```json
{
  "incident_time": "2026-01-03T10:30:00Z",
  "host": "web-01:9100",
  "affected_metrics": [
    "node_cpu_seconds_total",
    "node_memory_MemAvailable_bytes"
  ],
  "root_causes": [
    "CPU 使用率异常升高",
    "内存使用率异常升高"
  ],
  "recommendations": [
    "检查高 CPU 进程，考虑优化代码或扩容",
    "检查内存泄漏，考虑增加内存或优化应用"
  ],
  "metrics_timeline": {
    "node_cpu_seconds_total": {
      "baseline": 45.2,
      "peak": 92.5,
      "anomaly": true,
      "points": 30
    },
    "node_memory_MemAvailable_bytes": {
      "baseline": 8589934592,
      "peak": 2147483648,
      "anomaly": true,
      "points": 30
    }
  }
}
```

## 2. 容量规划 What-if 分析

### 2.1 核心功能

基于历史负载数据进行趋势分析，预测容量耗尽时间，支持多种场景分析。

### 2.2 趋势分析算法

#### 线性回归
```
y = slope * x + intercept
slope = (n*Σxy - Σx*Σy) / (n*Σx² - (Σx)²)
intercept = (Σy - slope*Σx) / n
```

#### 指数平滑
```
S_t = α * Y_t + (1 - α) * S_{t-1}
其中 α 为平滑系数（通常 0.1-0.3）
```

### 2.3 容量预测

```
DaysToCapacity = (CapacityThreshold - CurrentValue) / GrowthRate
ProjectedPeak = slope * futureIndex + intercept
```

### 2.4 API 接口

#### 容量规划分析
```
POST /api/prometheus/capacity-planning
Content-Type: application/json

{
  "metric_query": "100 - (avg by (instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
  "days": 30,
  "capacity_threshold": 90.0,
  "current_capacity": 100.0,
  "cost_per_unit": 100.0
}
```

**响应示例：**
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
      "peak_multiplier": 1.0,
      "projected_metrics": {
        "30day_peak": 85.3,
        "days_to_capacity": 10
      },
      "recommended_action": "立即规划扩容",
      "time_to_action_days": 7
    },
    {
      "name": "加速增长(1.5x)",
      "growth_multiplier": 1.5,
      "peak_multiplier": 1.2,
      "projected_metrics": {
        "30day_peak": 102.4,
        "days_to_capacity": 6
      },
      "recommended_action": "紧急扩容计划",
      "time_to_action_days": 3
    },
    {
      "name": "业务高峰(2x)",
      "growth_multiplier": 2.0,
      "peak_multiplier": 2.0,
      "projected_metrics": {
        "30day_peak": 170.6,
        "days_to_capacity": 3
      },
      "recommended_action": "立即启动应急扩容",
      "time_to_action_days": 1
    }
  ],
  "recommendation": {
    "current_usage": 72.5,
    "projected_usage": 85.3,
    "recommended_capacity": 102.4,
    "utilization_rate": 0.725,
    "safety_margin": 0.275,
    "priority": "medium",
    "action_items": [
      "计划扩容，安全余量不足20%",
      "增长率为 0.50/天，建议加快扩容进度"
    ]
  },
  "resource_allocation": {
    "current_allocation": 100.0,
    "recommended_allocation": 102.4,
    "allocation_ratio": 1.024,
    "timeline": "2周内",
    "cost_estimate": 240.0,
    "roi": 2400.0
  },
  "data_points": 100
}
```

## 3. 与现有工具系统的集成

### 3.1 内置工具注册

在 `internal/tool/builtin/prometheus_tool.go` 中定义了两个内置工具：

#### 故障复盘工具
```go
tool.Name: "prometheus_incident_replay"
tool.Parameters:
  - incident_time (string, required): 故障发生时间
  - host (string, required): 目标主机
  - lookback_minutes (integer, optional): 回溯时间
```

#### 容量规划工具
```go
tool.Name: "prometheus_capacity_planning"
tool.Parameters:
  - metric_query (string, required): PromQL 查询
  - days (integer, optional): 历史数据天数
  - capacity_threshold (number, optional): 容量阈值
```

### 3.2 工具注册方式

在 `cmd/server/main.go` 中添加：

```go
import "ai-ops/internal/tool/builtin"

// 注册 Prometheus 工具
if cfg.Prometheus.Enabled {
    err := builtin.RegisterPrometheusTools(toolRegistry, cfg.Prometheus.Address)
    if err != nil {
        logger.Warn("注册 Prometheus 工具失败", zap.Error(err))
    }
}
```

### 3.3 配置文件

在 `config.yaml` 中添加：

```yaml
prometheus:
  enabled: true
  address: "http://localhost:9090"
```

## 4. 完整的代码文件清单

### 核心实现文件

1. **C:\Users\hrp\Downloads\ai-pro\internal\prometheus\client.go**
   - Prometheus 客户端封装
   - 范围查询和即时查询
   - 元数据和标签值获取

2. **C:\Users\hrp\Downloads\ai-pro\internal\prometheus\incident_replay.go**
   - 故障复盘查询实现
   - PromQL 查询语句集合
   - 故障报告生成

3. **C:\Users\hrp\Downloads\ai-pro\internal\prometheus\capacity_planning.go**
   - 容量规划器实现
   - 趋势分析算法
   - What-if 场景分析
   - 资源分配策略

4. **C:\Users\hrp\Downloads\ai-pro\internal\tool\builtin\prometheus_tool.go**
   - 内置工具定义
   - 工具注册函数
   - 参数处理和结果格式化

5. **C:\Users\hrp\Downloads\ai-pro\internal\api\handler\prometheus.go**
   - HTTP 处理器实现
   - API 端点定义
   - 请求/响应处理

6. **C:\Users\hrp\Downloads\ai-pro\internal\api\router.go** (已更新)
   - 添加 PrometheusAddr 配置字段
   - 注册 Prometheus API 路由

## 5. 依赖管理

需要在 `go.mod` 中添加 Prometheus 客户端依赖：

```bash
go get github.com/prometheus/client_golang@latest
```

## 6. 使用示例

### 6.1 通过 AI 对话使用

```
用户: 分析一下 2026-01-03 10:30 时 web-01 主机的故障
AI: 我来帮你查询故障时间段的指标数据...
[调用 prometheus_incident_replay 工具]
结果: 发现 CPU 和内存异常升高，建议检查进程...
```

### 6.2 通过 API 直接调用

```bash
# 故障复盘
curl -X POST http://localhost:8080/api/prometheus/incident-replay \
  -H "Content-Type: application/json" \
  -d '{
    "incident_time": "2026-01-03T10:30:00Z",
    "host": "web-01:9100"
  }'

# 容量规划
curl -X POST http://localhost:8080/api/prometheus/capacity-planning \
  -H "Content-Type: application/json" \
  -d '{
    "metric_query": "100 - (avg by (instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
    "days": 30,
    "capacity_threshold": 90.0
  }'
```

## 7. 性能优化建议

### 7.1 查询优化
- 使用合适的 step 参数（通常 30s-5m）
- 限制查询时间范围（最多 30 天）
- 使用 recording rules 预计算常用指标

### 7.2 缓存策略
- 缓存元数据查询结果（TTL: 1 小时）
- 缓存历史数据查询结果（TTL: 5 分钟）
- 实现增量查询避免重复数据

### 7.3 并发控制
- 限制并发查询数（最多 10 个）
- 实现查询队列和超时机制
- 监控 Prometheus 服务器负载

## 8. 故障排查

### 问题：查询超时
**解决方案：**
- 减少查询时间范围
- 增加 step 参数
- 检查 Prometheus 服务器性能

### 问题：数据不完整
**解决方案：**
- 确保 Node Exporter 正常运行
- 检查 Prometheus 数据保留期设置
- 验证指标标签配置

### 问题：预测不准确
**解决方案：**
- 增加历史数据天数（至少 30 天）
- 检查数据中是否存在异常值
- 调整平滑系数或回归模型

## 9. 安全考虑

- 限制 Prometheus API 访问权限
- 使用 HTTPS 加密传输
- 实现请求认证和授权
- 审计所有查询操作

## 10. 扩展方向

- 支持多个 Prometheus 实例
- 集成告警规则自动生成
- 支持自定义预测模型
- 集成成本分析和优化建议
