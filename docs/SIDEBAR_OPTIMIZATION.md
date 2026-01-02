# AI-Pro 侧边栏和全局状态优化方案

## 概述

本优化方案为 AI-Pro 项目添加了三个核心功能：
1. **会话健康状态评分** - 实时追踪会话质量
2. **MCP 连接指示器** - 显示 Agent 与后端服务的通信状态
3. **全局任务进度条** - 非阻塞式进度显示

## 架构设计

### 1. 全局状态管理 (system.ts)

新增 `useSystemStore` 用于管理系统级状态：

```typescript
// C:\Users\hrp\Downloads\ai-pro\web\src\stores\system.ts
- MCPConnection: MCP 连接状态追踪
- TaskProgress: 任务进度管理
- SessionHealth: 会话健康评分
```

**核心功能：**
- 实时连接状态监控（connected/disconnected/error）
- 平均延迟计算
- 会话健康评分算法（基于响应时间、错误率、工具成功率）

### 2. 组件架构

#### MCPStatusIndicator.vue
```
位置: C:\Users\hrp\Downloads\ai-pro\web\src\components\common\MCPStatusIndicator.vue
功能: 侧边栏底部显示 MCP 连接状态
特性:
  - 状态指示灯（绿/黄/红/灰）
  - 脉冲动画效果
  - 延迟显示
  - 悬浮提示详细信息
```

#### SessionHealthBadge.vue
```
位置: C:\Users\hrp\Downloads\ai-pro\web\src\components\common\SessionHealthBadge.vue
功能: 会话列表中显示健康评分
特性:
  - 0-100 分评分系统
  - 颜色编码（优秀/良好/警告/差）
  - 详细指标提示
```

#### GlobalTaskProgress.vue
```
位置: C:\Users\hrp\Downloads\ai-pro\web\src\components\common\GlobalTaskProgress.vue
功能: 全局任务进度显示
特性:
  - 右上角浮动显示
  - 多任务并行显示
  - 进度条动画
  - 自动清理已完成任务
  - 不阻塞用户操作
```

## 集成步骤

### 步骤 1: 在 App.vue 中注册全局组件

```vue
<!-- C:\Users\hrp\Downloads\ai-pro\web\src\App.vue -->
<template>
  <div id="app">
    <!-- 全局任务进度条 -->
    <GlobalTaskProgress />

    <!-- 其他内容 -->
    <router-view />
  </div>
</template>

<script setup lang="ts">
import GlobalTaskProgress from '@/components/common/GlobalTaskProgress.vue'
</script>
```

### 步骤 2: 侧边栏已自动集成

侧边栏组件已更新，包含 MCP 状态指示器：
```
文件: C:\Users\hrp\Downloads\ai-pro\web\src\components\common\AppSidebar.vue
位置: sidebar-footer 区域
```

### 步骤 3: 历史面板已自动集成

历史面板已更新，包含会话健康徽章：
```
文件: C:\Users\hrp\Downloads\ai-pro\web\src\components\chat\HistoryPanelOverlay.vue
位置: session-item 内部
```

## 使用方法

### 1. 更新 MCP 连接状态

```typescript
import { useSystemStore } from '@/stores/system'

const systemStore = useSystemStore()

// 更新连接状态
systemStore.updateMCPConnection({
  id: 'mcp-main',
  name: 'Main MCP Server',
  status: 'connected',
  latency: 45,
  lastPing: Date.now()
})
```

### 2. 添加任务进度

```typescript
import { useSystemStore } from '@/stores/system'

const systemStore = useSystemStore()

// 添加新任务
const taskId = `scan-${Date.now()}`
systemStore.addTaskProgress({
  id: taskId,
  type: 'scan',
  title: '深度安全扫描',
  progress: 0,
  status: 'running',
  startTime: Date.now(),
  estimatedTime: 60000 // 60秒
})

// 更新进度
systemStore.updateTaskProgress(taskId, {
  progress: 50
})

// 完成任务
systemStore.updateTaskProgress(taskId, {
  progress: 100,
  status: 'completed'
})
```

### 3. 会话健康状态自动追踪

会话健康状态已集成到 `useChatStore`，会在每次消息完成时自动更新：

```typescript
// 自动追踪指标：
- 响应时间
- 错误率
- 工具调用成功率

// 评分算法：
- 响应时间 < 1s: 满分
- 响应时间 1-3s: -10分
- 响应时间 3-5s: -20分
- 响应时间 > 5s: -30分
- 错误率: 每1%扣0.4分
- 工具失败率: 每1%扣0.3分
```

## 后端集成建议

### 1. WebSocket 连接状态推送

```go
// 建议在后端添加 WebSocket 端点
// 路径: /api/system/status

type MCPStatusMessage struct {
    Type      string  `json:"type"`      // "mcp_status"
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Status    string  `json:"status"`    // connected/disconnected/error
    Latency   int     `json:"latency"`   // ms
    Timestamp int64   `json:"timestamp"`
}
```

前端监听示例：
```typescript
const ws = new WebSocket('ws://localhost:8080/api/system/status')

ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  if (data.type === 'mcp_status') {
    systemStore.updateMCPConnection({
      id: data.id,
      name: data.name,
      status: data.status,
      latency: data.latency,
      lastPing: data.timestamp
    })
  }
}
```

### 2. 任务进度推送

```go
type TaskProgressMessage struct {
    Type          string `json:"type"`           // "task_progress"
    ID            string `json:"id"`
    TaskType      string `json:"task_type"`      // scan/analysis/execution
    Title         string `json:"title"`
    Progress      int    `json:"progress"`       // 0-100
    Status        string `json:"status"`         // running/completed/error
    EstimatedTime int64  `json:"estimated_time"` // ms
}
```

## 样式定制

所有组件都使用 CSS 变量，可以通过主题系统定制：

```css
/* 在全局样式中覆盖 */
:root {
  --bg-elevated: #ffffff;
  --border-color: #e5e7eb;
  --text-primary: #1f2937;
  --text-secondary: #6b7280;
}

/* 暗色主题 */
.dark {
  --bg-elevated: #1e293b;
  --border-color: rgba(255, 255, 255, 0.1);
  --text-primary: #f1f5f9;
  --text-secondary: #94a3b8;
}
```

## 性能优化

1. **会话健康状态计算**
   - 使用滑动窗口保留最近 100 条响应时间
   - 避免内存泄漏

2. **任务进度自动清理**
   - 完成任务 5 秒后自动移除
   - 最多同时显示 5 个任务

3. **MCP 状态更新节流**
   - 建议后端每 2 秒推送一次状态
   - 前端使用防抖处理频繁更新

## 测试建议

### 单元测试
```typescript
// 测试会话健康评分算法
describe('SessionHealth', () => {
  it('should calculate correct score', () => {
    const score = systemStore.calculateSessionScore('test', {
      avgResponseTime: 1500,
      errorCount: 1,
      totalMessages: 10,
      toolSuccessCount: 8,
      toolTotalCount: 10
    })
    expect(score).toBeGreaterThan(60)
  })
})
```

### 集成测试
```typescript
// 测试 MCP 状态更新
it('should update MCP connection status', () => {
  systemStore.updateMCPConnection({
    id: 'test',
    name: 'Test',
    status: 'connected',
    latency: 50,
    lastPing: Date.now()
  })
  expect(systemStore.overallConnectionStatus).toBe('connected')
})
```

## 文件清单

### 新增文件
1. `C:\Users\hrp\Downloads\ai-pro\web\src\stores\system.ts` - 系统状态管理
2. `C:\Users\hrp\Downloads\ai-pro\web\src\components\common\MCPStatusIndicator.vue` - MCP 状态指示器
3. `C:\Users\hrp\Downloads\ai-pro\web\src\components\common\GlobalTaskProgress.vue` - 全局任务进度
4. `C:\Users\hrp\Downloads\ai-pro\web\src\components\common\SessionHealthBadge.vue` - 会话健康徽章

### 修改文件
1. `C:\Users\hrp\Downloads\ai-pro\web\src\components\common\AppSidebar.vue` - 集成 MCP 指示器
2. `C:\Users\hrp\Downloads\ai-pro\web\src\components\chat\HistoryPanelOverlay.vue` - 集成健康徽章
3. `C:\Users\hrp\Downloads\ai-pro\web\src\stores\chat.ts` - 添加健康状态追踪

## 下一步

1. 在 `App.vue` 中添加 `<GlobalTaskProgress />` 组件
2. 实现后端 WebSocket 推送
3. 添加用户偏好设置（是否显示进度条等）
4. 添加单元测试和集成测试
5. 性能监控和优化

## 总结

本优化方案采用最小化代码实现，核心特点：

- **轻量级**: 每个组件 < 150 行代码
- **非侵入式**: 不影响现有功能
- **高性能**: 使用 computed 和防抖优化
- **可扩展**: 易于添加新的状态指标
- **类型安全**: 完整的 TypeScript 支持
