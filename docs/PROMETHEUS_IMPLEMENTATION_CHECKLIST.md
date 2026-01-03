# Prometheus 集成实现清单

## 已完成的文件

### 1. 核心数据模型
- **文件**: `internal/prometheus/types.go`
- **内容**:
  - AlertLevel, MetricSample, RangeSeries 等基础类型
  - BaselineMetrics, AnomalyResult 异常检测类型
  - SemanticTranslation, NLQuery, PromQLQuery 语义翻译类型
  - Playbook, PlaybookStep, PlaybookExecution 剧本类型
  - DashboardConfig, PanelConfig, Dashboard 看板类型
  - PrometheusConfig, QueryRange 配置类型

### 2. Prometheus 客户端
- **文件**: `internal/prometheus/client.go`
- **功能**:
  - Query() - 即时查询
  - QueryRange() - 范围查询
  - LabelValues() - 获取标签值
  - Series() - 获取系列
  - Rules() - 获取告警规则
  - Targets() - 获取目标
  - Health() - 健康检查

### 3. 语义翻译引擎 (Prometheus-1)
- **文件**: `internal/prometheus/semantic.go`
- **功能**:
  - TranslateMetric() - 指标转自然语言
  - GenerateAlertDescription() - 生成告警描述
  - GenerateRecommendations() - 生成建议
  - 支持 CPU、内存、磁盘、网络等指标翻译

### 4. PromQL 生成器 (Prometheus-2)
- **文件**: `internal/prometheus/promql_generator.go`
- **功能**:
  - GenerateQuery() - 自然语言转 PromQL
  - OptimizeQuery() - 优化 PromQL
  - ExplainQuery() - 解释 PromQL
  - 支持 CPU、内存、磁盘、负载、网络等查询

### 5. 动态阈值检测 (Prometheus-3)
- **文件**: `internal/prometheus/threshold_detector.go`
- **功能**:
  - LearnBaseline() - 学习基线
  - DetectAnomaly() - 检测异常
  - UpdateThreshold() - 更新阈值
  - GetBaseline() - 获取基线
  - 使用 3-sigma 规则进行异常检测

### 6. 自愈剧本引擎 (Prometheus-4)
- **文件**: `internal/prometheus/playbook_engine.go`
- **功能**:
  - CreatePlaybook() - 创建剧本
  - ExecutePlaybook() - 执行剧本
  - GetPlaybook() - 获取剧本
  - ListPlaybooks() - 列表剧本
  - GetExecution() - 获取执行记录
  - DefaultPlaybooks() - 默认剧本模板

### 7. 动态看板生成器 (Prometheus-5)
- **文件**: `internal/prometheus/dashboard_generator.go`
- **功能**:
  - GenerateDashboard() - 生成看板
  - UpdateDashboard() - 更新看板
  - GetDashboard() - 获取看板
  - RecommendPanels() - 推荐面板
  - ListDashboards() - 列表看板
  - DeleteDashboard() - 删除看板
  - GetDashboardTemplate() - 获取模板
  - ListTemplates() - 列表模板

### 8. API 处理器
- **文件**: `internal/api/handler/prometheus.go`
- **新增端点**:
  - POST /api/prometheus/translate - 翻译指标
  - POST /api/prometheus/generate-promql - 生成 PromQL
  - POST /api/prometheus/detect-anomaly - 检测异常
  - POST /api/prometheus/playbooks - 创建剧本
  - POST /api/prometheus/playbooks/:id/execute - 执行剧本
  - GET /api/prometheus/playbooks - 列表剧本
  - GET /api/prometheus/playbooks/:id - 获取剧本
  - POST /api/prometheus/dashboards - 生成看板
  - GET /api/prometheus/dashboards - 列表看板
  - GET /api/prometheus/dashboards/:id - 获取看板
  - GET /api/prometheus/dashboards/recommend-panels - 推荐面板

### 9. 设计文档
- **文件**: `docs/PROMETHEUS_INTEGRATION_DESIGN.md`
- **内容**: 完整的架构设计、接口定义、实现方案

## 需要修改的文件

### 1. 路由配置
- **文件**: `internal/api/router.go`
- **修改**: 添加 Prometheus 路由组

```go
// 添加 Prometheus 路由
promHandler, err := handler.NewPrometheusHandler(cfg.Prometheus.URL)
if err != nil {
    logger.Fatal("初始化 Prometheus 处理器失败", zap.Error(err))
}

promGroup := r.Group("/api/prometheus")
{
    promGroup.POST("/translate", promHandler.TranslateMetric)
    promGroup.POST("/generate-promql", promHandler.GeneratePromQL)
    promGroup.POST("/detect-anomaly", promHandler.DetectAnomaly)
    promGroup.POST("/playbooks", promHandler.CreatePlaybook)
    promGroup.POST("/playbooks/:id/execute", promHandler.ExecutePlaybook)
    promGroup.GET("/playbooks", promHandler.ListPlaybooks)
    promGroup.GET("/playbooks/:id", promHandler.GetPlaybook)
    promGroup.POST("/dashboards", promHandler.GenerateDashboard)
    promGroup.GET("/dashboards", promHandler.ListDashboards)
    promGroup.GET("/dashboards/:id", promHandler.GetDashboard)
    promGroup.GET("/dashboards/recommend-panels", promHandler.RecommendPanels)
}
```

### 2. 配置文件
- **文件**: `config.yaml`
- **添加**:

```yaml
prometheus:
  url: "http://localhost:9090"
  timeout: 30s
  username: ""
  password: ""
  tls_verify: false

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

### 3. go.mod 依赖
- **文件**: `go.mod`
- **需要添加**:

```
github.com/prometheus/client_golang v1.17.0
github.com/prometheus/common v0.44.0
```

## 前端组件需要创建

### 1. PromQL 查询组件
- **文件**: `web/src/components/prometheus/PromQLQuery.vue`
- **功能**: 自然语言转 PromQL 查询

### 2. 指标查看器
- **文件**: `web/src/components/prometheus/MetricViewer.vue`
- **功能**: 显示指标数据和语义翻译

### 3. 动态看板
- **文件**: `web/src/components/prometheus/DynamicDashboard.vue`
- **功能**: 千人千面的个性化看板

### 4. 剧本执行器
- **文件**: `web/src/components/prometheus/PlaybookExecutor.vue`
- **功能**: 执行自愈剧本

### 5. 前端 API 调用
- **文件**: `web/src/api/prometheus.ts`
- **功能**: 封装 Prometheus API 调用

## 测试用例

### 单元测试
- `internal/prometheus/semantic_test.go`
- `internal/prometheus/promql_generator_test.go`
- `internal/prometheus/threshold_detector_test.go`
- `internal/prometheus/playbook_engine_test.go`
- `internal/prometheus/dashboard_generator_test.go`

### 集成测试
- `internal/api/handler/prometheus_test.go`

## 部署检查清单

- [ ] 配置 Prometheus 服务器地址
- [ ] 配置告警规则和阈值
- [ ] 初始化基线数据
- [ ] 创建默认剧本
- [ ] 创建默认看板模板
- [ ] 配置 LLM 集成
- [ ] 测试 API 端点
- [ ] 部署前端组件
- [ ] 配置监控告警
- [ ] 文档更新

