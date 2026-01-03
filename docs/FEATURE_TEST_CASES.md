# 新增功能测试用例

## 概述

本文档包含以下4个新功能的详细测试用例：
1. 执行结果的"语义化提炼"
2. "联动式"动作推荐
3. "会话记忆"与"关联分析"
4. UI 细节微调

---

## 1. 语义化提炼功能测试

### 1.1 后端 API 测试

#### 测试接口: `POST /api/analysis/semantic`

**测试用例 1.1.1: 磁盘检查结果分析**
```bash
curl -X POST http://localhost:8080/api/analysis/semantic \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "check_disk",
    "tool_result": "Filesystem      Size  Used Avail Use% Mounted on\n/dev/sda1       100G   77G   23G  77% /\n/dev/sdb1       500G  450G   50G  90% /data",
    "host_id": "server-001"
  }'
```

**预期响应:**
```json
{
  "code": 0,
  "data": {
    "summary": "/data 分区使用率达 90%，建议清理",
    "risk_level": "warning",
    "trend": "过去24小时增长约5%",
    "recommendation": "建议清理 /var/log 或扩容磁盘",
    "key_metrics": [
      {"name": "根分区", "value": "77%", "status": "normal", "threshold": "85%"},
      {"name": "数据分区", "value": "90%", "status": "warning", "threshold": "85%"}
    ]
  }
}
```

**测试用例 1.1.2: 内存检查结果分析**
```bash
curl -X POST http://localhost:8080/api/analysis/semantic \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "check_memory",
    "tool_result": "              total        used        free      shared  buff/cache   available\nMem:          16Gi       14Gi       500Mi       256Mi       1.5Gi       1.2Gi",
    "host_id": "server-001"
  }'
```

**预期响应:**
```json
{
  "code": 0,
  "data": {
    "summary": "内存使用率较高，可用内存仅 1.2GB",
    "risk_level": "warning",
    "recommendation": "建议检查内存占用进程或增加内存"
  }
}
```

**测试用例 1.1.3: CPU 检查结果分析（危急）**
```bash
curl -X POST http://localhost:8080/api/analysis/semantic \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "check_cpu",
    "tool_result": "CPU Usage: 99.5%\nLoad Average: 8.5, 7.2, 6.8",
    "host_id": "server-001"
  }'
```

**预期响应:**
```json
{
  "code": 0,
  "data": {
    "summary": "CPU 使用率过高，需立即处理",
    "risk_level": "critical",
    "recommendation": "建议检查高负载进程"
  }
}
```

### 1.2 前端组件测试

#### 测试组件: `EnhancedToolCallCard.vue`

**测试场景 1.2.1: AI 深度解读区块显示**

1. 打开聊天页面
2. 选择一台主机
3. 发送消息: "检查磁盘使用情况"
4. 等待工具执行完成

**预期结果:**
- 工具卡片下方显示 "AI 深度解读" 区块
- 显示加载动画 "正在分析..."
- 分析完成后显示:
  - 风险等级标签（绿色/黄色/红色）
  - 摘要信息
  - 趋势分析（如有）
  - 建议操作（如有）
  - 关键指标列表

**测试场景 1.2.2: 风险等级颜色验证**

| 风险等级 | 标签颜色 | 图标 |
|---------|---------|------|
| normal | 绿色 (success) | CircleCheck |
| warning | 黄色 (warning) | WarningFilled |
| critical | 红色 (danger) | CircleClose |

### 1.3 前后端交互测试

**测试流程:**
```
1. 用户发送 "检查 server-001 的磁盘"
   ↓
2. 后端执行 check_disk 工具
   ↓
3. 前端收到工具执行结果 (status: success)
   ↓
4. EnhancedToolCallCard 组件触发 fetchSemanticInsight()
   ↓
5. 前端调用 POST /api/analysis/semantic
   ↓
6. 后端调用 LLM 分析结果
   ↓
7. 返回 SemanticInsight 数据
   ↓
8. 前端渲染 AI 深度解读区块
```

---

## 2. 联动式动作推荐功能测试

### 2.1 后端 API 测试

#### 测试接口: `POST /api/analysis/suggest-actions`

**测试用例 2.1.1: 磁盘满时推荐清理操作**
```bash
curl -X POST http://localhost:8080/api/analysis/suggest-actions \
  -H "Content-Type: application/json" \
  -d '{
    "tool_results": [
      {
        "tool_name": "check_disk",
        "result": "Use%: 95%",
        "host_id": "server-001"
      }
    ]
  }'
```

**预期响应:**
```json
{
  "code": 0,
  "data": [
    {
      "id": "clear_cache",
      "label": "清理缓存",
      "icon": "Delete",
      "command": "清理 /tmp 和 /var/cache 目录",
      "highlight": true,
      "ai_recommended": true,
      "confidence": 0.85,
      "risk_level": "low",
      "description": "清理临时文件可释放磁盘空间"
    },
    {
      "id": "clear_logs",
      "label": "清理日志",
      "icon": "Document",
      "highlight": true,
      "ai_recommended": true,
      "confidence": 0.80,
      "risk_level": "low"
    },
    {
      "id": "view_logs",
      "label": "查看日志",
      "icon": "Document",
      "highlight": false,
      "ai_recommended": false,
      "risk_level": "low"
    }
  ]
}
```

**测试用例 2.1.2: CPU 高时推荐查看进程**
```bash
curl -X POST http://localhost:8080/api/analysis/suggest-actions \
  -H "Content-Type: application/json" \
  -d '{
    "tool_results": [
      {
        "tool_name": "check_cpu",
        "result": "CPU: 99%",
        "host_id": "server-001"
      }
    ]
  }'
```

**预期响应包含:**
```json
{
  "id": "check_process",
  "label": "查看进程",
  "ai_recommended": true,
  "confidence": 0.90
}
```

**测试用例 2.1.3: 无工具结果时返回默认操作**
```bash
curl -X POST http://localhost:8080/api/analysis/suggest-actions \
  -H "Content-Type: application/json" \
  -d '{
    "tool_results": []
  }'
```

**预期响应:** 返回默认操作列表（查看日志、检查状态、重启服务、扩容）

### 2.2 前端组件测试

#### 测试组件: `ContextualActionsPanel.vue`

**测试场景 2.2.1: AI 推荐操作高亮显示**

1. 执行磁盘检查命令
2. 观察右侧智能操作面板

**预期结果:**
- "AI 推荐" 分组显示在顶部
- 推荐操作带有紫色边框和渐变背景
- 显示置信度标签（如 "85% 置信度"）
- 高亮操作有脉冲动画效果

**测试场景 2.2.2: 操作点击确认**

1. 点击 "重启服务" 按钮（中风险）
2. 观察确认弹窗

**预期结果:**
- 弹出确认对话框
- 显示操作名称和描述
- 显示风险等级警告
- 确认后发送聊天消息

**测试场景 2.2.3: 低风险操作直接执行**

1. 点击 "查看日志" 按钮（低风险）

**预期结果:**
- 不弹出确认框
- 直接发送聊天消息

### 2.3 前后端交互测试

**测试流程:**
```
1. 用户执行工具命令
   ↓
2. ChatWindow 组件更新 latestToolResults
   ↓
3. ContextualActionsPanel 监听到变化
   ↓
4. 调用 fetchSuggestedActions()
   ↓
5. POST /api/analysis/suggest-actions
   ↓
6. 后端分析工具结果，生成推荐操作
   ↓
7. 前端渲染操作按钮
   ↓
8. 用户点击操作 → 发送聊天消息
```

---

## 3. 会话记忆与关联分析功能测试

### 3.1 后端 API 测试

#### 测试接口: `POST /api/analysis/correlation`

**测试用例 3.1.1: 磁盘问题历史关联**
```bash
curl -X POST http://localhost:8080/api/analysis/correlation \
  -H "Content-Type: application/json" \
  -d '{
    "host_id": "server-001",
    "current_issue": "磁盘使用率达到 95%",
    "issue_type": "disk"
  }'
```

**预期响应:**
```json
{
  "code": 0,
  "data": {
    "host_id": "server-001",
    "current_issue": "磁盘使用率达到 95%",
    "similar_incident": "该主机在历史上出现过类似的disk问题",
    "occurred_at": "3天前",
    "resolution": "上次通过清理 /var/log 和 /tmp 目录解决，释放了约 20GB 空间",
    "confidence": 0.75
  }
}
```

**测试用例 3.1.2: 内存问题历史关联**
```bash
curl -X POST http://localhost:8080/api/analysis/correlation \
  -H "Content-Type: application/json" \
  -d '{
    "host_id": "server-002",
    "current_issue": "内存不足，OOM Killer 触发",
    "issue_type": "memory"
  }'
```

**预期响应包含:**
```json
{
  "resolution": "上次通过重启服务并调整 JVM 参数解决"
}
```

**测试用例 3.1.3: 无匹配历史**
```bash
curl -X POST http://localhost:8080/api/analysis/correlation \
  -H "Content-Type: application/json" \
  -d '{
    "host_id": "new-server",
    "current_issue": "未知错误",
    "issue_type": ""
  }'
```

**预期响应:**
```json
{
  "code": 0,
  "data": {
    "host_id": "new-server",
    "current_issue": "未知错误",
    "confidence": 0
  }
}
```

### 3.2 前端组件测试

#### 测试组件: `HistoricalCorrelationCard.vue`

**测试场景 3.2.1: 历史关联卡片显示**

1. 执行磁盘检查，结果显示使用率 > 90%
2. 等待 AI 深度解读完成（risk_level = warning）

**预期结果:**
- 在工具卡片下方显示黄色历史关联卡片
- 显示 "历史关联分析" 标题
- 显示匹配度百分比
- 显示相似问题描述
- 显示历史解决方案
- 显示 "尝试此方案" 按钮

**测试场景 3.2.2: 点击应用解决方案**

1. 点击 "尝试此方案" 按钮

**预期结果:**
- 触发 `applySolution` 事件
- 父组件接收解决方案文本

**测试场景 3.2.3: 无历史关联时不显示**

1. 执行检查，结果正常（risk_level = normal）

**预期结果:**
- 不显示历史关联卡片

### 3.3 前后端交互测试

**测试流程:**
```
1. 工具执行完成，AI 解读显示 warning/critical
   ↓
2. EnhancedToolCallCard 计算 showHistoricalCorrelation = true
   ↓
3. HistoricalCorrelationCard 组件挂载
   ↓
4. 调用 fetchCorrelation()
   ↓
5. POST /api/analysis/correlation
   ↓
6. 后端匹配历史问题模式
   ↓
7. 返回关联分析结果
   ↓
8. 前端渲染历史关联卡片
```

---

## 4. UI 细节微调测试

### 4.1 右侧面板样式测试

**测试场景 4.1.1: 背景色渐变**

1. 打开聊天页面
2. 观察右侧面板

**预期结果:**
- 背景色为渐变: `#FAFBFC → #F5F7FA`
- 与中间白色聊天区有明显视觉区分
- 左侧有柔和阴影: `box-shadow: -2px 0 8px rgba(0, 0, 0, 0.03)`

**测试场景 4.1.2: 面板头部样式**

**预期结果:**
- 头部背景为紫色渐变: `#667eea → #764ba2`
- 标题文字为白色
- 与内容区形成层次感

### 4.2 工具卡片样式测试

**测试场景 4.2.1: 卡片阴影效果**

1. 执行任意工具命令
2. 观察工具卡片

**预期结果:**
- 默认阴影: `0 2px 8px rgba(0, 0, 0, 0.06)`
- 边框: `1px solid #e8e8e8`
- 圆角: `12px`

**测试场景 4.2.2: 卡片 Hover 效果**

1. 鼠标悬停在工具卡片上

**预期结果:**
- 阴影增强: `0 6px 20px rgba(0, 0, 0, 0.1)`
- 轻微上移: `transform: translateY(-1px)`
- 过渡动画平滑

**测试场景 4.2.3: 状态颜色阴影**

| 状态 | 左边框颜色 | 阴影颜色 |
|------|-----------|---------|
| running | #f59e0b | rgba(245, 158, 11, 0.15) |
| success | #10b981 | rgba(16, 185, 129, 0.1) |
| error | #ef4444 | rgba(239, 68, 68, 0.1) |

### 4.3 监控区域卡片测试

**测试场景 4.3.1: 主机监控卡片样式**

1. 选择主机
2. 观察右侧监控区域

**预期结果:**
- 白色背景: `#fff`
- 圆角: `12px`
- 阴影: `0 2px 8px rgba(0, 0, 0, 0.06)`
- 内边距: `var(--spacing-4)`

**测试场景 4.3.2: 监控卡片 Hover 效果**

1. 鼠标悬停在监控卡片上

**预期结果:**
- 阴影增强: `0 4px 16px rgba(0, 0, 0, 0.1)`

---

## 5. 集成测试场景

### 5.1 完整工作流测试

**场景: 磁盘告警处理流程**

```
步骤 1: 选择主机 server-001
步骤 2: 发送 "检查磁盘使用情况"
步骤 3: 等待工具执行完成
步骤 4: 验证 AI 深度解读显示 (risk_level: warning)
步骤 5: 验证历史关联卡片显示
步骤 6: 验证右侧面板显示 "清理缓存" AI 推荐操作
步骤 7: 点击 "清理缓存" 操作
步骤 8: 验证聊天消息发送
步骤 9: 验证清理命令执行
```

### 5.2 错误处理测试

**测试场景 5.2.1: API 请求失败**

1. 断开网络连接
2. 执行工具命令

**预期结果:**
- AI 深度解读显示失败提示或使用降级方案
- 智能操作面板显示默认操作
- 历史关联卡片不显示

**测试场景 5.2.2: 空结果处理**

1. 执行返回空结果的命令

**预期结果:**
- 不显示 AI 深度解读
- 智能操作面板显示默认操作

---

## 6. 性能测试

### 6.1 响应时间测试

| 接口 | 目标响应时间 |
|------|-------------|
| /api/analysis/semantic | < 3s (含 LLM 调用) |
| /api/analysis/suggest-actions | < 500ms |
| /api/analysis/correlation | < 500ms |

### 6.2 并发测试

```bash
# 使用 ab 进行并发测试
ab -n 100 -c 10 -p request.json -T application/json \
  http://localhost:8080/api/analysis/suggest-actions
```

**预期结果:**
- 99% 请求在 1s 内完成
- 无请求失败

---

## 7. 浏览器兼容性测试

| 浏览器 | 版本 | 测试项 |
|--------|------|--------|
| Chrome | 120+ | 所有功能 |
| Firefox | 120+ | 所有功能 |
| Safari | 17+ | 所有功能 |
| Edge | 120+ | 所有功能 |

---

## 附录: 测试数据准备

### 模拟磁盘检查结果
```
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1       100G   95G    5G  95% /
/dev/sdb1       500G  450G   50G  90% /data
tmpfs           8.0G  256M  7.8G   4% /run
```

### 模拟内存检查结果
```
              total        used        free      shared  buff/cache   available
Mem:          16Gi       14Gi       500Mi       256Mi       1.5Gi       1.2Gi
Swap:          4Gi       2.0Gi       2.0Gi
```

### 模拟 CPU 检查结果
```
CPU Usage: 99.5%
Load Average: 8.5, 7.2, 6.8
Top Processes:
  PID   %CPU  COMMAND
  1234  45.2  java
  5678  32.1  python
  9012  15.3  nginx
```
