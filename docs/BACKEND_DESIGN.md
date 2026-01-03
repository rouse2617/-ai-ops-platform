# AI-Ops 运维平台后端设计文档

## 1. 架构设计

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer (Gin)                       │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │  Chat    │  Host    │ Operation│ Monitor  │Prometheus│  │
│  │ Handler  │ Handler  │ Handler  │ Handler  │ Handler  │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │  Chat    │  Host    │Operation │ Monitor  │Prometheus│  │
│  │ Service  │ Service  │ Service  │ Service  │ Service  │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Core Components                           │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │ Danger   │ Session  │  Agent   │ Anomaly  │  Trend   │  │
│  │Interceptor│ Manager │ Executor │ Detector │ Analyzer │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                        │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │   SSH    │   DB     │  Cache   │  LLM     │Prometheus│  │
│  │   Pool   │  (GORM)  │ (Memory) │ Client   │ Client   │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 模块职责

- **API Layer**: 处理 HTTP 请求、参数验证、响应封装
- **Service Layer**: 业务逻辑编排、事务管理
- **Core Components**: 核心功能实现（拦截、检测、分析）
- **Infrastructure Layer**: 基础设施（数据库、缓存、外部服务）

## 2. 核心功能设计

### 2.1 P0: 高危操作确认机制

#### 设计思路
1. 命令拦截：在 SSH 执行前检测高危命令
2. 用户确认：通过 WebSocket 实时推送确认请求
3. 超时处理：30秒内未确认则拒绝执行
4. 审计日志：记录所有拦截和确认操作

#### 核心组件
- `DangerInterceptor`: 高危命令拦截器
- `ConfirmationManager`: 确认请求管理器
- `AuditLogger`: 审计日志记录器

### 2.2 P1: 上下文与多主机管理

#### 设计思路
1. Session-Host 绑定：会话创建时关联主机列表
2. 上下文保持：会话内保持主机上下文
3. 批量执行：支持多主机并发操作
4. 结果聚合：统一展示多主机执行结果

#### 核心组件
- `SessionManager`: 会话管理器
- `HostContextManager`: 主机上下文管理器
- `BatchExecutor`: 批量执行器

### 2.3 高级功能

#### 趋势预警
- 基于历史数据的趋势分析
- 机器学习预测资源使用
- 提前预警潜在问题

#### 晨间体检
- 定时健康检查任务
- 生成每日健康报告
- 异常自动告警

#### 影子观察者
- 静默监控模式
- 异常行为检测
- 智能告警降噪

### 2.4 Prometheus 集成

#### 功能
- 指标采集和查询
- 自定义告警规则
- 可视化数据展示

## 3. 数据库模型设计

### 3.1 新增表

#### operation_confirmations (操作确认表)
```sql
CREATE TABLE operation_confirmations (
    id VARCHAR(36) PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL,
    host_id VARCHAR(36) NOT NULL,
    command TEXT NOT NULL,
    risk_level VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    requested_at TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP,
    confirmed_by VARCHAR(100),
    timeout_at TIMESTAMP NOT NULL,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session (session_id),
    INDEX idx_status (status),
    INDEX idx_timeout (timeout_at)
);
```

#### health_checks (健康检查表)
```sql
CREATE TABLE health_checks (
    id VARCHAR(36) PRIMARY KEY,
    host_id VARCHAR(36) NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    metrics JSON,
    issues JSON,
    suggestions JSON,
    checked_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_host (host_id),
    INDEX idx_checked_at (checked_at)
);
```

#### trend_predictions (趋势预测表)
```sql
CREATE TABLE trend_predictions (
    id VARCHAR(36) PRIMARY KEY,
    host_id VARCHAR(36) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    current_value FLOAT NOT NULL,
    predicted_value FLOAT NOT NULL,
    prediction_time TIMESTAMP NOT NULL,
    confidence FLOAT NOT NULL,
    alert_level VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_host (host_id),
    INDEX idx_prediction_time (prediction_time)
);
```

#### prometheus_alerts (Prometheus 告警表)
```sql
CREATE TABLE prometheus_alerts (
    id VARCHAR(36) PRIMARY KEY,
    alert_name VARCHAR(100) NOT NULL,
    host_id VARCHAR(36),
    labels JSON,
    annotations JSON,
    status VARCHAR(20) NOT NULL,
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_host (host_id),
    INDEX idx_status (status),
    INDEX idx_starts_at (starts_at)
);
```

### 3.2 表结构变更

#### sessions 表增强
```sql
ALTER TABLE sessions ADD COLUMN context JSON;
ALTER TABLE sessions ADD COLUMN last_activity TIMESTAMP;
ALTER TABLE sessions ADD COLUMN host_ids JSON;
```

## 4. API 端点设计

### 4.1 高危操作确认

```
POST   /api/operations/confirm-request
GET    /api/operations/confirmations/:id
POST   /api/operations/confirmations/:id/approve
POST   /api/operations/confirmations/:id/reject
GET    /api/operations/confirmations
```

### 4.2 会话与主机管理

```
POST   /api/sessions
GET    /api/sessions
GET    /api/sessions/:id
PUT    /api/sessions/:id/hosts
DELETE /api/sessions/:id
GET    /api/sessions/:id/context
```

### 4.3 健康检查

```
POST   /api/health/check
GET    /api/health/reports
GET    /api/health/reports/:id
POST   /api/health/schedule
GET    /api/health/daily-report
```

### 4.4 趋势预警

```
GET    /api/trends/predictions
POST   /api/trends/analyze
GET    /api/trends/alerts
POST   /api/trends/alerts/:id/acknowledge
```

### 4.5 Prometheus 集成

```
GET    /api/prometheus/metrics
POST   /api/prometheus/query
GET    /api/prometheus/alerts
POST   /api/prometheus/alerts/rules
GET    /api/prometheus/targets
```

## 5. 技术选型

### 5.1 核心框架
- **Web 框架**: Gin v1.9+
- **ORM**: GORM v1.25+
- **数据库**: SQLite/MySQL/PostgreSQL
- **缓存**: 内存缓存 (sync.Map)
- **WebSocket**: gorilla/websocket

### 5.2 监控与可观测性
- **指标采集**: Prometheus Client
- **日志**: zap
- **追踪**: OpenTelemetry (可选)

### 5.3 安全
- **认证**: JWT
- **加密**: crypto/aes
- **审计**: 自定义审计日志

## 6. 性能优化

### 6.1 并发控制
- 使用 Worker Pool 限制并发数
- 合理使用 goroutine，避免泄漏
- 使用 context 管理超时和取消

### 6.2 缓存策略
- 主机信息缓存 (5分钟)
- 会话上下文缓存 (30分钟)
- 指标数据缓存 (1分钟)

### 6.3 数据库优化
- 合理使用索引
- 批量操作优化
- 连接池配置

## 7. 安全设计

### 7.1 命令白名单
- 默认只允许只读命令
- 高危命令需要确认
- 完全禁止的命令列表

### 7.2 审计日志
- 记录所有命令执行
- 记录确认操作
- 记录异常访问

### 7.3 权限控制
- 基于角色的访问控制 (RBAC)
- 主机级别权限
- 操作级别权限

## 8. 测试策略

### 8.1 单元测试
- 覆盖率目标: 80%+
- Table-driven 测试
- Mock 外部依赖

### 8.2 集成测试
- API 端到端测试
- 数据库集成测试
- SSH 连接测试

### 8.3 性能测试
- 并发压力测试
- 响应时间测试
- 资源使用测试

## 9. 部署方案

### 9.1 容器化
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ai-ops cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ai-ops .
CMD ["./ai-ops"]
```

### 9.2 配置管理
- 环境变量配置
- 配置文件支持
- 动态配置更新

## 10. 监控与告警

### 10.1 系统指标
- API 请求量和延迟
- SSH 连接池状态
- 数据库连接状态
- 内存和 CPU 使用

### 10.2 业务指标
- 命令执行成功率
- 高危操作拦截率
- 告警触发频率
- 用户活跃度

## 11. 扩展性设计

### 11.1 插件化
- 自定义检测规则
- 自定义告警通道
- 自定义命令工具

### 11.2 分布式支持
- 多实例部署
- 负载均衡
- 会话共享

## 12. 开发计划

### Phase 1: 核心功能 (Week 1-2)
- [ ] 高危操作拦截机制
- [ ] Session-Host 关联
- [ ] 基础 API 实现

### Phase 2: 高级功能 (Week 3-4)
- [ ] 趋势预警
- [ ] 健康检查
- [ ] Prometheus 集成

### Phase 3: 优化与测试 (Week 5-6)
- [ ] 性能优化
- [ ] 测试覆盖
- [ ] 文档完善

## 13. 风险与挑战

### 13.1 技术风险
- SSH 连接稳定性
- 并发控制复杂度
- 实时通信可靠性

### 13.2 业务风险
- 误拦截正常操作
- 告警风暴
- 性能瓶颈

### 13.3 缓解措施
- 完善的测试覆盖
- 灰度发布策略
- 监控和告警机制
