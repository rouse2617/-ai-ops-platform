# AI-Ops 运维平台 - API 使用文档

## 目录
1. [高危操作确认 API](#高危操作确认-api)
2. [健康检查 API](#健康检查-api)
3. [趋势分析 API](#趋势分析-api)
4. [Prometheus 集成 API](#prometheus-集成-api)
5. [完整示例](#完整示例)

---

## 高危操作确认 API

### 1. 获取确认请求详情

```http
GET /api/operations/confirmations/:id
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "conf-123",
    "session_id": "sess-456",
    "host_id": "host-789",
    "command": "rm -rf /tmp/test",
    "risk_level": "high",
    "status": "pending",
    "requested_at": "2025-01-03T10:00:00Z",
    "timeout_at": "2025-01-03T10:00:30Z",
    "reason": "删除操作"
  }
}
```

### 2. 批准操作

```http
POST /api/operations/confirmations/:id/approve
Content-Type: application/json

{
  "confirmed_by": "admin"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "操作已批准",
  "data": null
}
```

### 3. 拒绝操作

```http
POST /api/operations/confirmations/:id/reject
Content-Type: application/json

{
  "confirmed_by": "admin",
  "reason": "风险过高"
}
```

### 4. 获取待确认列表

```http
GET /api/operations/confirmations/pending?session_id=sess-456
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "confirmations": [
      {
        "id": "conf-123",
        "command": "rm -rf /tmp/test",
        "risk_level": "high",
        "status": "pending",
        "requested_at": "2025-01-03T10:00:00Z"
      }
    ]
  }
}
```

---

## 健康检查 API

### 1. 检查单个主机

```http
POST /api/health/check
Content-Type: application/json

{
  "host_id": "host-789"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "host_id": "host-789",
    "host_name": "web-server-01",
    "status": "warning",
    "metrics": {
      "cpu_usage": 85.5,
      "memory_usage": 72.3,
      "disk_usage": 68.9,
      "load_average": {
        "1min": 2.5,
        "5min": 2.1,
        "15min": 1.8
      }
    },
    "issues": [
      "CPU 使用率过高: 85.5%"
    ],
    "suggestions": [
      "检查 CPU 占用最高的进程"
    ],
    "checked_at": "2025-01-03T10:00:00Z"
  }
}
```

### 2. 每日健康体检

```http
POST /api/health/daily-report
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "results": {
      "host-789": {
        "host_name": "web-server-01",
        "status": "healthy",
        "metrics": {...}
      },
      "host-790": {
        "host_name": "db-server-01",
        "status": "warning",
        "metrics": {...}
      }
    },
    "summary": {
      "total": 10,
      "healthy": 7,
      "warning": 2,
      "critical": 1
    }
  }
}
```

### 3. 获取健康检查历史

```http
GET /api/health/reports?host_id=host-789
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "history": [
      {
        "id": "check-001",
        "host_id": "host-789",
        "status": "healthy",
        "metrics": {...},
        "checked_at": "2025-01-03T09:00:00Z"
      }
    ]
  }
}
```

---

## 趋势分析 API

### 1. 分析趋势

```http
POST /api/trends/analyze
Content-Type: application/json

{
  "host_id": "host-789"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "predictions": [
      {
        "host_id": "host-789",
        "host_name": "web-server-01",
        "metric_type": "cpu",
        "current_value": 75.5,
        "predicted_value": 82.3,
        "trend": "increasing",
        "confidence": 0.85,
        "alert_level": "medium",
        "recommendation": "CPU 使用率呈上升趋势，建议监控",
        "prediction_time": "2025-01-03T11:00:00Z"
      },
      {
        "metric_type": "memory",
        "current_value": 68.2,
        "predicted_value": 70.1,
        "trend": "stable",
        "confidence": 0.92
      }
    ]
  }
}
```

### 2. 获取预测历史

```http
GET /api/trends/predictions?host_id=host-789
```

### 3. 获取告警

```http
GET /api/trends/alerts?host_id=host-789
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "alerts": [
      {
        "id": "pred-001",
        "host_id": "host-789",
        "metric_type": "cpu",
        "predicted_value": 92.5,
        "alert_level": "critical",
        "prediction_time": "2025-01-03T12:00:00Z"
      }
    ]
  }
}
```

---

## Prometheus 集成 API

### 1. 执行 PromQL 查询

```http
POST /api/prometheus/query
Content-Type: application/json

{
  "query": "node_cpu_seconds_total"
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "results": [
      {
        "metric": {
          "instance": "localhost:9100",
          "job": "node",
          "mode": "idle"
        },
        "value": 12345.67,
        "time": "2025-01-03T10:00:00Z"
      }
    ]
  }
}
```

### 2. 范围查询

```http
POST /api/prometheus/query_range
Content-Type: application/json

{
  "query": "rate(node_cpu_seconds_total[5m])",
  "start": 1704268800,
  "end": 1704272400,
  "step": 60
}
```

### 3. 获取监控目标

```http
GET /api/prometheus/targets
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "targets": [
      {
        "labels": {
          "instance": "localhost:9100",
          "job": "node"
        },
        "health": "up",
        "last_scrape": "2025-01-03T10:00:00Z",
        "scrape_url": "http://localhost:9100/metrics"
      }
    ]
  }
}
```

### 4. 获取告警

```http
GET /api/prometheus/alerts
```

### 5. 获取主机指标

```http
GET /api/prometheus/metrics?host_id=localhost:9100
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "metrics": {
      "cpu_usage": 45.2,
      "memory_usage": 62.8,
      "disk_usage": 58.3
    }
  }
}
```

---

## 完整示例

### 场景 1: 执行高危命令流程

```javascript
// 1. 用户发起聊天请求（包含高危命令）
const chatResponse = await fetch('/api/chat', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    session_id: 'sess-456',
    message: '删除 /tmp 目录下的所有文件',
    host_ids: ['host-789']
  })
});

// 2. 后端检测到高危命令，返回确认请求
// 响应中包含 confirmation_id

// 3. 前端轮询或通过 WebSocket 获取待确认列表
const pendingResponse = await fetch('/api/operations/confirmations/pending?session_id=sess-456');
const { confirmations } = await pendingResponse.json();

// 4. 用户确认操作
const confirmResponse = await fetch(`/api/operations/confirmations/${confirmations[0].id}/approve`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    confirmed_by: 'admin'
  })
});

// 5. 命令继续执行
```

### 场景 2: 每日健康体检

```javascript
// 定时任务：每天早上 8 点执行
async function dailyHealthCheck() {
  const response = await fetch('/api/health/daily-report', {
    method: 'POST'
  });

  const { results, summary } = await response.json();

  // 发送通知
  if (summary.critical > 0) {
    sendAlert(`发现 ${summary.critical} 台主机处于严重状态`);
  }

  // 生成报告
  generateReport(results);
}
```

### 场景 3: 趋势预警

```javascript
// 定时任务：每小时执行一次
async function trendAnalysis() {
  const hosts = await getHostList();

  for (const host of hosts) {
    const response = await fetch('/api/trends/analyze', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ host_id: host.id })
    });

    const { predictions } = await response.json();

    // 检查告警
    const alerts = predictions.filter(p => p.alert_level);
    if (alerts.length > 0) {
      sendTrendAlert(host, alerts);
    }
  }
}
```

### 场景 4: Prometheus 监控集成

```javascript
// 获取主机实时指标
async function getHostMetrics(hostId) {
  const response = await fetch(`/api/prometheus/metrics?host_id=${hostId}`);
  const { metrics } = await response.json();

  return metrics;
}

// 查询历史数据
async function getHistoricalData(query, startTime, endTime) {
  const response = await fetch('/api/prometheus/query_range', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query,
      start: Math.floor(startTime / 1000),
      end: Math.floor(endTime / 1000),
      step: 60
    })
  });

  const { results } = await response.json();
  return results;
}
```

---

## 错误处理

所有 API 遵循统一的错误响应格式：

```json
{
  "code": 400,
  "message": "参数错误: host_id 不能为空",
  "data": null
}
```

### 常见错误码

- `400`: 参数错误
- `401`: 未授权
- `403`: 禁止访问
- `404`: 资源不存在
- `500`: 服务器内部错误

---

## 认证

所有 API 请求需要在 Header 中携带 JWT Token：

```http
Authorization: Bearer <your-jwt-token>
```

---

## 速率限制

- 普通 API: 100 请求/分钟
- 查询 API: 1000 请求/分钟
- 批量操作: 10 请求/分钟

---

## WebSocket 实时通知

连接地址：`ws://your-domain/ws`

### 消息格式

```json
{
  "type": "confirmation_request",
  "data": {
    "id": "conf-123",
    "command": "rm -rf /tmp/test",
    "risk_level": "high"
  }
}
```

### 消息类型

- `confirmation_request`: 确认请求
- `health_alert`: 健康告警
- `trend_alert`: 趋势告警
- `prometheus_alert`: Prometheus 告警
