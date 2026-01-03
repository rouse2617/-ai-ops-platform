# AI-Ops 运维平台前端设计文档

## 1. 架构设计

### 1.1 技术栈
- **框架**: Vue 3.4+ (Composition API)
- **语言**: TypeScript 5.0+
- **状态管理**: Pinia 2.1+
- **UI 组件库**: Element Plus 2.4+
- **构建工具**: Vite 5.0+
- **图表库**: ECharts 5.4+ (Prometheus 监控)
- **样式方案**: SCSS + CSS Variables

### 1.2 目录结构
```
web/src/
├── components/
│   ├── chat/                    # 聊天相关组件
│   │   ├── DangerConfirmDialog.vue      # P0: 高危操作确认弹窗
│   │   ├── HostTagBadge.vue             # P0: 执行结果主机标签
│   │   ├── RightPanelSync.vue           # P1: 右侧面板联动
│   │   ├── CollapsibleContent.vue       # P1: 长内容折叠
│   │   ├── ThinkingChain.vue            # P1: 思考链展示
│   │   ├── SlotFillingPanel.vue         # P2: Slot Filling
│   │   ├── TrendWarningBubble.vue       # 高级: 趋势预警气泡
│   │   ├── HealthReportCard.vue         # 高级: 健康报告卡片
│   │   └── RiskWarningDialog.vue        # 高级: 风险警告弹窗
│   ├── prometheus/              # Prometheus 监控组件
│   │   ├── DynamicChart.vue             # 动态图表
│   │   ├── PromQLEditor.vue             # PromQL 查询界面
│   │   └── MetricSelector.vue           # 指标选择器
│   └── common/                  # 通用组件
│       └── MobileAdaptive.vue           # P2: 移动端适配容器
├── stores/
│   ├── danger.ts                # 高危操作状态管理
│   ├── panel.ts                 # 面板联动状态管理
│   ├── prometheus.ts            # Prometheus 状态管理
│   └── mobile.ts                # 移动端状态管理
├── composables/
│   ├── useDangerConfirm.ts      # 高危操作确认 Hook
│   ├── usePanelSync.ts          # 面板联动 Hook
│   ├── useCollapse.ts           # 折叠功能 Hook
│   ├── useSlotFilling.ts        # Slot Filling Hook
│   └── usePrometheus.ts         # Prometheus Hook
└── types/
    ├── danger.ts                # 高危操作类型定义
    ├── panel.ts                 # 面板类型定义
    └── prometheus.ts            # Prometheus 类型定义
```

### 1.3 响应式布局策略
```typescript
// 断点定义
const BREAKPOINTS = {
  mobile: 768,      // < 768px: 单栏布局
  tablet: 1200,     // 768-1200px: 两栏布局
  desktop: 1200     // > 1200px: 三栏布局
}

// 布局模式
type LayoutMode = 'mobile' | 'tablet' | 'desktop'
```

---

## 2. P0 功能实现

### 2.1 高危操作确认弹窗

#### 组件设计: DangerConfirmDialog.vue

**Props**
```typescript
interface DangerConfirmProps {
  visible: boolean              // 弹窗显示状态
  operation: DangerOperation    // 高危操作信息
  hostCount: number             // 影响主机数量
  estimatedImpact: string       // 预估影响
}

interface DangerOperation {
  id: string
  command: string               // 执行命令
  type: 'delete' | 'restart' | 'stop' | 'modify'
  riskLevel: 'high' | 'critical'
  affectedHosts: string[]
  reason?: string               // 风险原因
}
```

**Events**
```typescript
interface DangerConfirmEvents {
  'update:visible': (value: boolean) => void
  'confirm': (operation: DangerOperation) => void
  'cancel': () => void
}
```

**Slots**
```typescript
interface DangerConfirmSlots {
  header: () => VNode           // 自定义头部
  footer: () => VNode           // 自定义底部
  impact: () => VNode           // 自定义影响说明
}
```

#### 状态管理: stores/danger.ts

```typescript
export interface DangerState {
  pendingOperations: DangerOperation[]
  confirmHistory: ConfirmRecord[]
  autoConfirmDisabled: boolean
}

export const useDangerStore = defineStore('danger', {
  state: (): DangerState => ({
    pendingOperations: [],
    confirmHistory: [],
    autoConfirmDisabled: true
  }),

  actions: {
    async requestConfirm(operation: DangerOperation): Promise<boolean>
    recordConfirm(operation: DangerOperation, confirmed: boolean): void
    checkDangerLevel(command: string): DangerLevel
  }
})
```

### 2.2 执行结果主机标签

#### 组件设计: HostTagBadge.vue

**Props**
```typescript
interface HostTagProps {
  hostId: string
  status: 'success' | 'error' | 'warning' | 'running'
  result?: ExecutionResult
  showDetails?: boolean
}

interface ExecutionResult {
  exitCode: number
  output: string
  error?: string
  duration: number
  timestamp: number
}
```

**Events**
```typescript
interface HostTagEvents {
  'click': (hostId: string) => void
  'show-detail': (result: ExecutionResult) => void
}
```

---

## 3. P1 功能实现

### 3.1 右侧面板联动

#### 组件设计: RightPanelSync.vue

**Props**
```typescript
interface PanelSyncProps {
  activeHost?: string           // 当前激活主机
  syncMode: 'auto' | 'manual'   // 联动模式
  panels: PanelConfig[]         // 面板配置
}

interface PanelConfig {
  id: string
  type: 'metrics' | 'logs' | 'events'
  title: string
  refreshInterval?: number
}
```

**状态管理: stores/panel.ts**

```typescript
export const usePanelStore = defineStore('panel', {
  state: () => ({
    activeHost: null as string | null,
    syncEnabled: true,
    panelData: new Map<string, PanelData>(),
    lastSyncTime: 0
  }),

  actions: {
    syncToHost(hostId: string): void
    updatePanelData(panelId: string, data: PanelData): void
    enableAutoSync(): void
  }
})
```

### 3.2 长内容折叠

#### 组件设计: CollapsibleContent.vue

**Props**
```typescript
interface CollapsibleProps {
  content: string
  maxLines?: number             // 默认显示行数 (默认: 10)
  maxHeight?: number            // 最大高度 (px)
  expandable?: boolean          // 是否可展��� (默认: true)
  syntax?: string               // 语法高亮类型
}
```

**Events**
```typescript
interface CollapsibleEvents {
  'expand': () => void
  'collapse': () => void
}
```

### 3.3 思考链展示

#### 组件设计: ThinkingChain.vue

**Props**
```typescript
interface ThinkingChainProps {
  steps: ThinkingStep[]
  currentStep?: number
  showTimeline?: boolean
}

interface ThinkingStep {
  id: string
  step: number
  status: 'calling_llm' | 'executing_tools' | 'analyzing_results' | 'completed'
  content: string
  timestamp: number
  duration?: number
  toolCalls?: ToolCallInfo[]
}
```

---

## 4. P2 功能实现

### 4.1 移动端适配

#### 组件设计: MobileAdaptive.vue

**特性**
- 响应式布局自动切换
- 触摸手势支持
- 虚拟键盘适配
- 底部导航栏

**状态管理: stores/mobile.ts**

```typescript
export const useMobileStore = defineStore('mobile', {
  state: () => ({
    isMobile: false,
    orientation: 'portrait' as 'portrait' | 'landscape',
    viewportHeight: window.innerHeight,
    keyboardVisible: false
  }),

  actions: {
    detectDevice(): void
    handleOrientationChange(): void
    adjustForKeyboard(visible: boolean): void
  }
})
```

### 4.2 Slot Filling

#### 组件设计: SlotFillingPanel.vue

**Props**
```typescript
interface SlotFillingProps {
  template: CommandTemplate
  values: Record<string, string>
  validation?: ValidationRules
}

interface CommandTemplate {
  id: string
  command: string
  slots: SlotDefinition[]
}

interface SlotDefinition {
  name: string
  type: 'text' | 'select' | 'number' | 'host'
  required: boolean
  options?: string[]
  placeholder?: string
  validator?: (value: string) => boolean
}
```

---

## 5. 高级功能实现

### 5.1 趋势预警气泡

#### 组件设计: TrendWarningBubble.vue

**Props**
```typescript
interface TrendWarningProps {
  metric: string
  currentValue: number
  threshold: number
  trend: 'up' | 'down' | 'stable'
  severity: 'info' | 'warning' | 'critical'
  position: { x: number; y: number }
}
```

### 5.2 健康报告卡片

#### 组件设计: HealthReportCard.vue

**Props**
```typescript
interface HealthReportProps {
  hostId: string
  score: number                 // 0-100
  metrics: HealthMetric[]
  lastUpdate: number
  recommendations?: string[]
}

interface HealthMetric {
  name: string
  value: number
  status: 'good' | 'warning' | 'critical'
  trend: number                 // 变化百分比
}
```

### 5.3 风险警告弹窗

#### 组件设计: RiskWarningDialog.vue

**Props**
```typescript
interface RiskWarningProps {
  visible: boolean
  risks: RiskItem[]
  autoShow?: boolean
}

interface RiskItem {
  id: string
  type: 'security' | 'performance' | 'availability'
  severity: 'low' | 'medium' | 'high' | 'critical'
  title: string
  description: string
  affectedHosts: string[]
  recommendation: string
  timestamp: number
}
```

---

## 6. Prometheus 监控实现

### 6.1 动态图表

#### 组件设计: DynamicChart.vue

**Props**
```typescript
interface DynamicChartProps {
  query: string                 // PromQL 查询
  timeRange: TimeRange
  refreshInterval?: number      // 刷新间隔 (ms)
  chartType: 'line' | 'area' | 'bar'
  height?: number
}

interface TimeRange {
  start: number
  end: number
  step?: string                 // 查询步长
}
```

**状态管理: stores/prometheus.ts**

```typescript
export const usePrometheusStore = defineStore('prometheus', {
  state: () => ({
    queries: new Map<string, QueryResult>(),
    timeRange: { start: Date.now() - 3600000, end: Date.now() },
    refreshInterval: 30000,
    connected: false
  }),

  actions: {
    async executeQuery(query: string): Promise<QueryResult>
    updateTimeRange(range: TimeRange): void
    setRefreshInterval(interval: number): void
  }
})
```

### 6.2 PromQL 查询界面

#### 组件设计: PromQLEditor.vue

**Props**
```typescript
interface PromQLEditorProps {
  modelValue: string
  suggestions?: string[]
  validation?: boolean
  height?: number
}
```

**Features**
- 语法高亮
- 自动补全
- 查询历史
- 语法验证
- 快捷查询模板

---

## 7. 样式设计

### 7.1 设计令牌 (CSS Variables)

```scss
:root {
  // 颜色系统
  --color-primary: #409eff;
  --color-success: #67c23a;
  --color-warning: #e6a23c;
  --color-danger: #f56c6c;
  --color-error: #f56c6c;
  --color-info: #909399;

  // 危险等级颜色
  --color-risk-low: #67c23a;
  --color-risk-medium: #e6a23c;
  --color-risk-high: #f56c6c;
  --color-risk-critical: #c71585;

  // 状态颜色
  --color-status-success: #67c23a;
  --color-status-error: #f56c6c;
  --color-status-warning: #e6a23c;
  --color-status-running: #409eff;

  // 间距
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;

  // 圆角
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;

  // 阴影
  --shadow-sm: 0 2px 4px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 8px rgba(0, 0, 0, 0.12);
  --shadow-lg: 0 8px 16px rgba(0, 0, 0, 0.15);

  // 动画
  --transition-fast: 150ms ease;
  --transition-base: 300ms ease;
  --transition-slow: 500ms ease;
}
```

### 7.2 组件样式规范

```scss
// 高危操作弹窗
.danger-confirm-dialog {
  .el-dialog__header {
    background: linear-gradient(135deg, #f56c6c 0%, #c71585 100%);
    color: white;
  }

  .danger-level-critical {
    border: 2px solid var(--color-risk-critical);
    animation: pulse 2s infinite;
  }

  .affected-hosts {
    max-height: 200px;
    overflow-y: auto;
  }
}

// 主机标签
.host-tag-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: 12px;
  transition: all var(--transition-fast);

  &.status-success {
    background: rgba(103, 194, 58, 0.1);
    color: var(--color-success);
    border: 1px solid var(--color-success);
  }

  &.status-error {
    background: rgba(245, 108, 108, 0.1);
    color: var(--color-danger);
    border: 1px solid var(--color-danger);
  }

  &:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-sm);
  }
}

// 思考链
.thinking-chain {
  .step-item {
    position: relative;
    padding-left: 32px;

    &::before {
      content: '';
      position: absolute;
      left: 12px;
      top: 24px;
      bottom: -8px;
      width: 2px;
      background: var(--color-primary);
    }

    &:last-child::before {
      display: none;
    }
  }

  .step-icon {
    position: absolute;
    left: 0;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--color-primary);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

// 折叠内容
.collapsible-content {
  position: relative;

  &.collapsed {
    max-height: 300px;
    overflow: hidden;

    &::after {
      content: '';
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 60px;
      background: linear-gradient(transparent, white);
    }
  }

  .expand-button {
    margin-top: var(--spacing-sm);
    width: 100%;
  }
}

// Prometheus 图表
.prometheus-chart {
  .chart-container {
    position: relative;
    background: white;
    border-radius: var(--radius-md);
    padding: var(--spacing-md);
    box-shadow: var(--shadow-sm);
  }

  .chart-loading {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.9);
  }
}
```

### 7.3 响应式样式

```scss
// 移动端适配
@media (max-width: 768px) {
  .danger-confirm-dialog {
    .el-dialog {
      width: 90% !important;
      margin: 0 auto;
    }
  }

  .host-tag-badge {
    font-size: 11px;
    padding: 4px 8px;
  }

  .thinking-chain {
    .step-item {
      padding-left: 24px;
      font-size: 14px;
    }
  }

  .prometheus-chart {
    .chart-container {
      padding: var(--spacing-sm);
    }
  }
}

// 平板适配
@media (min-width: 768px) and (max-width: 1200px) {
  .right-panel-sync {
    width: 320px;
  }
}
```

---

## 8. 性能优化策略

### 8.1 组件懒加载

```typescript
// router/index.ts
const routes = [
  {
    path: '/prometheus',
    component: () => import('@/views/Prometheus.vue')
  }
]

// 组件内懒加载
const DynamicChart = defineAsyncComponent(() =>
  import('@/components/prometheus/DynamicChart.vue')
)
```

### 8.2 虚拟滚动

```typescript
// 长列表使用虚拟滚动
import { useVirtualList } from '@vueuse/core'

const { list, containerProps, wrapperProps } = useVirtualList(
  hosts,
  { itemHeight: 48 }
)
```

### 8.3 防抖节流

```typescript
// composables/useDebounce.ts
export function useDebouncedRef<T>(value: T, delay = 300) {
  return refDebounced(value, delay)
}

// 使用示例
const searchQuery = useDebouncedRef('', 500)
```

---

## 9. 测试策略

### 9.1 单元测试

```typescript
// components/__tests__/DangerConfirmDialog.spec.ts
import { mount } from '@vue/test-utils'
import DangerConfirmDialog from '../DangerConfirmDialog.vue'

describe('DangerConfirmDialog', () => {
  it('should show critical warning for high-risk operations', () => {
    const wrapper = mount(DangerConfirmDialog, {
      props: {
        visible: true,
        operation: {
          type: 'delete',
          riskLevel: 'critical'
        }
      }
    })

    expect(wrapper.find('.danger-level-critical').exists()).toBe(true)
  })
})
```

### 9.2 E2E 测试

```typescript
// e2e/danger-confirm.spec.ts
import { test, expect } from '@playwright/test'

test('danger confirm workflow', async ({ page }) => {
  await page.goto('/chat')
  await page.fill('input[placeholder="输入命令"]', 'rm -rf /')
  await page.click('button:has-text("发送")')

  await expect(page.locator('.danger-confirm-dialog')).toBeVisible()
  await page.click('button:has-text("确认")')
})
```

---

## 10. 部署清单

### 10.1 构建优化

```typescript
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'echarts': ['echarts'],
          'vendor': ['vue', 'pinia', 'vue-router']
        }
      }
    },
    chunkSizeWarningLimit: 1000
  }
})
```

### 10.2 环境变量

```bash
# .env.production
VITE_API_BASE_URL=https://api.aiops.example.com
VITE_PROMETHEUS_URL=https://prometheus.example.com
VITE_WS_URL=wss://api.aiops.example.com/ws
```

---

## 11. 开发规范

### 11.1 命名规范

- 组件: PascalCase (DangerConfirmDialog.vue)
- 文件: kebab-case (danger-confirm.ts)
- 变量/函数: camelCase (handleConfirm)
- 常量: UPPER_SNAKE_CASE (MAX_RETRY_COUNT)
- CSS 类: kebab-case (danger-confirm-dialog)

### 11.2 代码审查清单

- [ ] TypeScript 类型完整
- [ ] Props 验证完整
- [ ] 事件命名规范
- [ ] 响应式设计实现
- [ ] 加载/错误状态处理
- [ ] 可访问性支持
- [ ] 性能优化考虑
- [ ] 单元测试覆盖

---

## 12. 总结

本设计文档提供了 AI-Ops 运维平台前端的完整实现方案，涵盖：

1. **P0 功能**: 高危操作确认、主机标签
2. **P1 功能**: 面板联动、内容折叠、思考链
3. **P2 功能**: 移动端适配、Slot Filling
4. **高级功能**: 趋势预警、健康报告、风险警告
5. **Prometheus**: 动态图表、PromQL 编辑器

所有组件遵循 Vue 3 Composition API 最佳实践，使用 TypeScript 确保类型安全，通过 Pinia 管理状态，Element Plus 提供 UI 基础，ECharts 实现数据可视化。
