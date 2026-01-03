# AI-Ops Prometheus 集成 - 实现检查清单

## 已完成的实现

### 核心模块 (4 个文件)

#### 1. Prometheus 客户端
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\client.go`
- [x] PrometheusClient 结构体定义
- [x] QueryRange 范围查询方法
- [x] QueryInstant 即时查询方法
- [x] GetMetricMetadata 元数据获取
- [x] GetLabelValues 标签值获取
- [x] 错误处理和超时管理

#### 2. 故障复盘模块
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\incident_replay.go`
- [x] IncidentReplayQuery 结构体
- [x] IncidentTimeline 时间线数据结构
- [x] MetricTimeSeries 指标时间序列
- [x] QueryIncidentMetrics 故障指标查询
- [x] calculateBaseline 基线计算
- [x] calculatePeak 峰值计算
- [x] detectAnomaly 异常检测
- [x] PromQLQueries 常用查询语句集合
  - CPUUsageQuery
  - MemoryUsageQuery
  - DiskIOQuery
  - NetworkBytesQuery
  - LoadAverageQuery
  - ContextSwitchQuery
  - ProcessCountQuery
  - FileDescriptorQuery
  - TCPConnectionQuery
- [x] GenerateIncidentReport 故障报告生成

#### 3. 容量规划模块
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\capacity_planning.go`
- [x] CapacityPlanner 结构体
- [x] LoadTrend 负载趋势数据结构
- [x] GetHistoricalData 历史数据获取
- [x] AnalyzeTrend 趋势分析
- [x] linearRegression 线性回归算法
- [x] ExponentialSmoothing 指数平滑预测
- [x] PredictCapacityExhaustion 容量耗尽预测
- [x] AnalyzeWhatIfScenarios What-if 场景分析
  - 正常增长场景
  - 加速增长场景 (1.5x)
  - 业务高峰场景 (2.0x)
- [x] GenerateCapacityRecommendation 容量建议生成
- [x] CalculateResourceAllocation 资源分配计算

#### 4. 初始化模块
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\prometheus\init.go`
- [x] InitPrometheusIntegration 初始化函数
- [x] 客户端创建和验证
- [x] 工具注册
- [x] 连接测试

### 工具系统集成 (1 个文件)

#### 5. 内置工具
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\tool\builtin\prometheus_tool.go`
- [x] PrometheusIncidentReplayTool 故障复盘工具
  - Name() 方法
  - Description() 方法
  - Parameters() 方法
  - Execute() 方法
- [x] PrometheusCapacityPlanningTool 容量规划工具
  - Name() 方法
  - Description() 方法
  - Parameters() 方法
  - Execute() 方法
- [x] RegisterPrometheusTools 工具注册函数
- [x] 结果格式化方法

### API 接口层 (1 个文件)

#### 6. HTTP 处理器
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\api\handler\prometheus.go`
- [x] PrometheusHandler 结构体
- [x] NewPrometheusHandler 构造函数
- [x] IncidentReplayRequest 请求结构
- [x] IncidentReplayResponse 响应结构
- [x] QueryIncidentMetrics 处理器方法
- [x] CapacityPlanningRequest 请求结构
- [x] CapacityPlanningResponse 响应结构
- [x] AnalyzeCapacityPlanning 处理器方法
- [x] PromQLQueryRequest 自定义查询请求
- [x] QueryPrometheus 自定义查询处理器
- [x] GetMetricsMetadata 元数据处理器
- [x] GetLabelValues 标签值处理器
- [x] formatScenarios 场景格式化方法

### 路由配置 (1 个文件)

#### 7. 路由集成
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\api\router.go`
- [x] RouterConfig 添加 PrometheusAddr 字段
- [x] Prometheus API 路由组注册
  - POST /api/prometheus/incident-replay
  - POST /api/prometheus/capacity-planning
  - POST /api/prometheus/query
  - GET /api/prometheus/metrics
  - GET /api/prometheus/labels/:label

### 配置结构 (1 个文件)

#### 8. 配置定义
**文件**: `C:\Users\hrp\Downloads\ai-pro\internal\config\prometheus.go`
- [x] PrometheusConfig 结构体
  - Enabled 字段
  - Address 字段
  - Timeout 字段

### 文档 (3 个文件)

#### 9. 完整集成文档
**文件**: `C:\Users\hrp\Downloads\ai-pro\docs\PROMETHEUS_INTEGRATION.md`
- [x] 功能概述
- [x] 关键指标说明
- [x] 时间窗口设计
- [x] PromQL 查询示例
- [x] API 接口文档
- [x] 工具系统集成说明
- [x] 配置文件示例
- [x] 使用示例
- [x] 性能优化建议
- [x] 故障排查指南

#### 10. 快速开始指南
**文件**: `C:\Users\hrp\Downloads\ai-pro\docs\PROMETHEUS_QUICKSTART.md`
- [x] 项目概述
- [x] 文件清单
- [x] 快速集成步骤
- [x] 核心 PromQL 查询
- [x] API 使用示例
- [x] AI 对话使用示例
- [x] 关键特性总结
- [x] 性能指标
- [x] 下一步工作

#### 11. 实现总结文档
**文件**: `C:\Users\hrp\Downloads\ai-pro\docs\PROMETHEUS_IMPLEMENTATION_SUMMARY.md`
- [x] 项目背景
- [x] 核心设计和架构图
- [x] 文件详细说明
- [x] 算法说明
- [x] 文件清单表格
- [x] 集成步骤
- [x] 使用示例
- [x] 关键特性总结
- [x] 性能指标
- [x] 下一步工作

#### 12. 配置示例
**文件**: `C:\Users\hrp\Downloads\ai-pro\docs\prometheus-config-example.yaml`
- [x] AI-Ops 配置示例
- [x] Prometheus 服务器配置
- [x] Recording Rules 配置
- [x] Alert Rules 配置

## 核心功能实现

### 故障复盘时光机

**功能**: 通过 Prometheus API 查询指定时间范围的历史指标数据

**实现方式**:
1. 接收故障时间和主机信息
2. 查询故障前 15 分钟到恢复后 15 分钟的指标数据
3. 计算基线值（故障前平均值）
4. 检测异常指标（偏差 > 50%）
5. 生成故障报告和根因分析

**��集的关键指标**:
- CPU 使用率
- 内存使用率
- 磁盘 I/O
- 网络流量
- 负载平均值
- 上下文切换
- 进程数
- 文件描述符
- TCP 连接数

**PromQL 查询示例**:
```promql
# CPU 使用率
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle",instance="<host>"}[5m])) * 100)

# 内存使用率
(1 - (node_memory_MemAvailable_bytes{instance="<host>"} / node_memory_MemTotal_bytes{instance="<host>"})) * 100

# 磁盘 I/O
rate(node_disk_io_time_seconds_total{instance="<host>"}[5m])

# 网络流量
rate(node_network_receive_bytes_total{instance="<host>"}[5m]) + rate(node_network_transmit_bytes_total{instance="<host>"}[5m])
```

### 容量规划 What-if 分析

**功能**: 基于历史负载数据进行趋势预测和容量规划

**实现方式**:
1. 获取指定天数的历史数据
2. 使用线性回归分析趋势
3. 计算增长率和预测峰值
4. 预测容量耗尽时间
5. 分析三种场景（正常、加速、高峰）
6. 生成容量建议和资源分配策略

**趋势分析算法**:
- 线性回归: 计算增长率和截距
- 指数平滑: 预测未来值
- 异常检测: 识别数据中的异常值

**What-if 场景**:
- 正常增长 (1.0x): 基于当前增长率
- 加速增长 (1.5x): 增长率提升 50%
- 业务高峰 (2.0x): 增长率翻倍

## API 接口总结

### 1. 故障复盘查询
```
POST /api/prometheus/incident-replay
```
**请求**: 故障时间、主机名
**响应**: 受影响指标、根因分析、建议

### 2. 容量规划分析
```
POST /api/prometheus/capacity-planning
```
**请求**: PromQL 查询、历史天数、容量阈值
**响应**: 趋势分析、What-if 场景、容量建议

### 3. 自定义 PromQL 查询
```
POST /api/prometheus/query
```
**请求**: PromQL 查询语句、时间范围、步长
**响应**: 查询结果

### 4. 获取指标元数据
```
GET /api/prometheus/metrics
```
**响应**: 所有可用指标列表

### 5. 获取标签值
```
GET /api/prometheus/labels/:label
```
**响应**: 指定标签的所有值

## 工具系统集成

### 工具1: 故障复盘工具
```
Name: prometheus_incident_replay
Parameters:
  - incident_time (string, required)
  - host (string, required)
  - lookback_minutes (integer, optional)
```

### 工具2: 容量规划工具
```
Name: prometheus_capacity_planning
Parameters:
  - metric_query (string, required)
  - days (integer, optional)
  - capacity_threshold (number, optional)
```

## 集成检查清单

### 前置条件
- [ ] Go 1.25.5 或更高版本
- [ ] Prometheus 服务器已部署
- [ ] Node Exporter 已在监控主机上运行

### 集成步骤
- [ ] 添加 Prometheus 客户端依赖: `go get github.com/prometheus/client_golang@latest`
- [ ] 复制所有 Prometheus 模块文件��� `internal/prometheus/`
- [ ] 复制工具文件到 `internal/tool/builtin/prometheus_tool.go`
- [ ] 复制处理器文件到 `internal/api/handler/prometheus.go`
- [ ] 更新 `internal/api/router.go` 添加 PrometheusAddr 字段和路由
- [ ] 复制配置文件到 `internal/config/prometheus.go`
- [ ] 在 `cmd/server/main.go` 中调用初始化函数
- [ ] 更新 `config.yaml` 添加 Prometheus 配置

### 验证步骤
- [ ] 编译项目: `go build ./cmd/server`
- [ ] 启动 Prometheus 服务
- [ ] 启动 Node Exporter
- [ ] 启动 AI-Ops 服务
- [ ] 测试故障复盘 API
- [ ] 测试容量规划 API
- [ ] 测试自定义查询 API
- [ ] 通过 AI 对话测试工具调用

## 性能指标

| 指标 | 目标值 |
|------|--------|
| 查询响应时间 | < 5 秒 |
| 支持历史数据 | 最多 30 天 |
| 并发查询限制 | 10 个 |
| 缓存 TTL | 5 分钟 |
| 元数据缓存 TTL | 1 小时 |

## 文件总结表

| 文件 | 行数 | 功能 |
|------|------|------|
| `internal/prometheus/client.go` | ~70 | Prometheus API 客户端 |
| `internal/prometheus/incident_replay.go` | ~250 | 故障复盘实现 |
| `internal/prometheus/capacity_planning.go` | ~350 | 容量规划实现 |
| `internal/prometheus/init.go` | ~50 | 初始化代码 |
| `internal/tool/builtin/prometheus_tool.go` | ~200 | 内置工具 |
| `internal/api/handler/prometheus.go` | ~300 | HTTP 处理器 |
| `internal/api/router.go` | +30 | 路由配置 (已更新) |
| `internal/config/prometheus.go` | ~10 | 配置结构 |
| **总计** | **~1,260** | **完整实现** |

## 下一步工作

### 立即可做
1. 在 `cmd/server/main.go` 中集成初始化代码
2. 更新 `config.yaml` 配置文件
3. 启动 Prometheus 和 Node Exporter
4. 测试所有 API 端点

### 短期工作 (1-2 周)
1. 编写单元测试
2. 编写集成测试
3. 在前端集成 Prometheus 分析面板
4. 配置告警规则

### 中期工作 (1 个月)
1. 优化查询性能
2. 实现缓存机制
3. 添加更多预测模型
4. 集成成本分析

### 长期工作 (2-3 个月)
1. 支持多个 Prometheus 实例
2. 集成告警规则自动生成
3. 支持自定义预测模型
4. 完整的容量规划报告生成

## 总结

本方案为 AI-Ops 项目提供了完整的 Prometheus 集成，包括：

1. **故障复盘时光机** - 查询历史指标进行故障分析
2. **容量规划 What-if 分析** - 基于趋势预测进行容量规划
3. **工具系统集成** - 通过 AI 对话调用分析工具
4. **API 接口** - RESTful API 支持直接调用
5. **完整文档** - 详细的集成和使用文档

所有代码文件已创建完成，可以直接集成到项目中。
