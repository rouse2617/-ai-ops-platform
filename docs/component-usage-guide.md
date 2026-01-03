# AI-Ops 前端组件使用指南

## 快速开始

本文档展示如何在 AI-Ops 运维平台中使用新开发的前端组件。

---

## 1. P0 功能：高危操作确认

### 使用 DangerConfirmDialog 组件

```vue
<template>
  <div>
    <el-button @click="executeCommand">执行删除命令</el-button>

    <DangerConfirmDialog
      v-model:visible="showConfirm"
      :operation="dangerOperation"
      :host-count="selectedHosts.length"
      estimated-impact="将删除所有数据，无法恢复"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DangerConfirmDialog from '@/components/chat/DangerConfirmDialog.vue'
import { useDangerConfirm } from '@/composables/useDangerConfirm'
import type { DangerOperation } from '@/stores/danger'

const { requestConfirm } = useDangerConfirm()

const showConfirm = ref(false)
const dangerOperation = ref<DangerOperation>()
const selectedHosts = ref(['host-1', 'host-2', 'host-3'])

const executeCommand = async () => {
  const command = 'rm -rf /data/*'

  // 使用 composable 检查命令
  const confirmed = await requestConfirm(command, selectedHosts.value)

  if (confirmed) {
    // 执行命令
    console.log('执行命令:', command)
  }
}

const handleConfirm = (operation: DangerOperation) => {
  console.log('用户确认执行:', operation)
  // 执行实际操作
}

const handleCancel = () => {
  console.log('用户取消操作')
}
</script>
```

### 使用 HostTagBadge 组件

```vue
<template>
  <div class="execution-results">
    <HostTagBadge
      v-for="result in executionResults"
      :key="result.hostId"
      :host-id="result.hostId"
      :status="result.status"
      :result="result.result"
      @click="handleHostClick"
      @show-detail="handleShowDetail"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import HostTagBadge from '@/components/chat/HostTagBadge.vue'
import type { ExecutionResult } from '@/components/chat/HostTagBadge.vue'

const executionResults = ref([
  {
    hostId: 'web-server-01',
    status: 'success' as const,
    result: {
      exitCode: 0,
      output: 'Command executed successfully',
      duration: 1234,
      timestamp: Date.now()
    }
  },
  {
    hostId: 'web-server-02',
    status: 'error' as const,
    result: {
      exitCode: 1,
      output: '',
      error: 'Connection timeout',
      duration: 5000,
      timestamp: Date.now()
    }
  }
])

const handleHostClick = (hostId: string) => {
  console.log('点击主机:', hostId)
}

const handleShowDetail = (result: ExecutionResult) => {
  console.log('查看详情:', result)
}
</script>
```

---

## 2. P1 功能：思考链展示

### 使用 ThinkingChain 组件

```vue
<template>
  <div class="chat-message">
    <ThinkingChain
      :steps="thinkingSteps"
      :current-step="currentStep"
      :show-timeline="true"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import ThinkingChain from '@/components/chat/ThinkingChain.vue'
import { useChatStore } from '@/stores/chat'

const chatStore = useChatStore()

const thinkingSteps = computed(() => chatStore.thinkingSteps)
const currentStep = computed(() => chatStore.currentThinkingStatus?.step)
</script>
```

### 使用 CollapsibleContent 组件

```vue
<template>
  <CollapsibleContent
    :content="longOutput"
    :max-lines="10"
    syntax="bash"
    @expand="handleExpand"
    @collapse="handleCollapse"
  />
</template>

<script setup lang="ts">
import CollapsibleContent from '@/components/chat/CollapsibleContent.vue'

const longOutput = `
[2024-01-03 10:00:00] Starting service...
[2024-01-03 10:00:01] Loading configuration...
[2024-01-03 10:00:02] Connecting to database...
[2024-01-03 10:00:03] Service started successfully
`.repeat(10)

const handleExpand = () => {
  console.log('内容已展开')
}

const handleCollapse = () => {
  console.log('内容已收起')
}
</script>
```

### 使用右侧面板联动

```vue
<template>
  <div class="ops-console">
    <div class="main-content">
      <!-- 主机列表 -->
      <HostList @select="handleHostSelect" />
    </div>

    <aside class="right-panel">
      <!-- 面板会自动同步到选中的主机 -->
      <MonitorPanel />
      <LogPanel />
      <EventPanel />
    </aside>
  </div>
</template>

<script setup lang="ts">
import { usePanelSync } from '@/composables/usePanelSync'

const { syncToHost, syncEnabled, toggleSync } = usePanelSync()

const handleHostSelect = (hostId: string) => {
  // 自动同步右侧面板到选中主机
  syncToHost(hostId)
}
</script>
```

---

## 3. Prometheus 监控

### 使用 DynamicChart 组件

```vue
<template>
  <div class="monitoring-dashboard">
    <DynamicChart
      query="rate(node_cpu_seconds_total{mode='user'}[5m])"
      title="CPU 使用率"
      :refresh-interval="30000"
      chart-type="area"
      :height="300"
    />

    <DynamicChart
      query="node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100"
      title="内存使用率"
      :refresh-interval="30000"
      chart-type="line"
      :height="300"
    />
  </div>
</template>

<script setup lang="ts">
import DynamicChart from '@/components/prometheus/DynamicChart.vue'
</script>
```

### 使用 PromQLEditor 组件

```vue
<template>
  <div class="query-panel">
    <PromQLEditor
      v-model="query"
      :suggestions="metricSuggestions"
      :validation="true"
      :height="150"
      @execute="handleExecute"
    />

    <div v-if="queryResult" class="query-result">
      <pre>{{ JSON.stringify(queryResult, null, 2) }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PromQLEditor from '@/components/prometheus/PromQLEditor.vue'
import { usePrometheus } from '@/composables/usePrometheus'

const { executeQuery, getMetrics } = usePrometheus()

const query = ref('rate(http_requests_total[5m])')
const queryResult = ref()
const metricSuggestions = ref<string[]>([])

onMounted(async () => {
  metricSuggestions.value = await getMetrics()
})

const handleExecute = async (q: string) => {
  queryResult.value = await executeQuery(q)
}
</script>
```

---

## 4. 高级功能

### 使用 HealthReportCard 组件

```vue
<template>
  <div class="health-dashboard">
    <HealthReportCard
      v-for="host in hosts"
      :key="host.id"
      :host-id="host.id"
      :score="host.healthScore"
      :metrics="host.metrics"
      :last-update="host.lastUpdate"
      :recommendations="host.recommendations"
      @refresh="handleRefresh(host.id)"
      @view-details="handleViewDetails(host.id)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import HealthReportCard from '@/components/chat/HealthReportCard.vue'
import type { HealthMetric } from '@/components/chat/HealthReportCard.vue'

const hosts = ref([
  {
    id: 'web-server-01',
    healthScore: 85,
    lastUpdate: Date.now(),
    metrics: [
      { name: 'CPU', value: 45.2, status: 'good' as const, trend: -2.3 },
      { name: '内存', value: 68.5, status: 'warning' as const, trend: 5.1 },
      { name: '磁盘', value: 82.3, status: 'warning' as const, trend: 1.2 },
      { name: '网络', value: 23.4, status: 'good' as const, trend: -0.5 }
    ],
    recommendations: [
      '建议清理临时文件释放磁盘空间',
      '内存使用率持续上升，建议检查内存泄漏'
    ]
  }
])

const handleRefresh = async (hostId: string) => {
  console.log('刷新主机健康数据:', hostId)
  // 调用 API 刷新数据
}

const handleViewDetails = (hostId: string) => {
  console.log('查看主机详情:', hostId)
}
</script>
```

### 使用 TrendWarningBubble 组件

```vue
<template>
  <div class="chart-container" @click="handleChartClick">
    <canvas ref="chartRef" />

    <TrendWarningBubble
      v-if="showWarning"
      metric="CPU 使用率"
      :current-value="cpuValue"
      :threshold="80"
      trend="up"
      severity="warning"
      :position="warningPosition"
      :change-percent="15.5"
      @view-details="handleViewDetails"
      @dismiss="showWarning = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import TrendWarningBubble from '@/components/chat/TrendWarningBubble.vue'

const showWarning = ref(false)
const cpuValue = ref(85.3)
const warningPosition = ref({ x: 200, y: 150 })

const handleChartClick = (e: MouseEvent) => {
  // 在点击位置显示警告气泡
  warningPosition.value = { x: e.clientX, y: e.clientY }
  showWarning.value = true
}

const handleViewDetails = (metric: string) => {
  console.log('查看指标详情:', metric)
}
</script>
```

### 使用 RiskWarningDialog 组件

```vue
<template>
  <div>
    <RiskWarningDialog
      v-model:visible="showRiskDialog"
      :risks="detectedRisks"
      :auto-show="true"
      @view-details="handleViewRiskDetails"
      @resolve="handleResolveRisk"
      @resolve-all="handleResolveAllRisks"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import RiskWarningDialog from '@/components/chat/RiskWarningDialog.vue'
import type { RiskItem } from '@/components/chat/RiskWarningDialog.vue'

const showRiskDialog = ref(false)
const detectedRisks = ref<RiskItem[]>([
  {
    id: 'risk-1',
    type: 'security',
    severity: 'critical',
    title: '检测到未授权访问尝试',
    description: '在过去 1 小时内检测到 50 次失败的 SSH 登录尝试',
    affectedHosts: ['web-server-01', 'web-server-02'],
    recommendation: '建议立即启用 fail2ban 并检查防火墙规则',
    timestamp: Date.now()
  },
  {
    id: 'risk-2',
    type: 'performance',
    severity: 'high',
    title: '数据库连接池耗尽',
    description: '数据库连接数达到上限，新请求被拒绝',
    affectedHosts: ['db-server-01'],
    recommendation: '增加连接池大小或优化慢查询',
    timestamp: Date.now()
  }
])

onMounted(() => {
  // 检测到风险时自动显示
  if (detectedRisks.value.length > 0) {
    showRiskDialog.value = true
  }
})

const handleViewRiskDetails = (risk: RiskItem) => {
  console.log('查看风险详情:', risk)
}

const handleResolveRisk = (riskId: string) => {
  console.log('标记风险已处理:', riskId)
  detectedRisks.value = detectedRisks.value.filter(r => r.id !== riskId)
}

const handleResolveAllRisks = () => {
  console.log('标记所有风险已处理')
  detectedRisks.value = []
}
</script>
```

---

## 5. 完整集成示例

### 在 ChatWindow 中集成所有功能

```vue
<template>
  <div class="chat-window-enhanced">
    <!-- 原有的聊天窗口 -->
    <ChatWindow />

    <!-- P0: 高危操作确认 -->
    <DangerConfirmDialog
      v-model:visible="dangerStore.hasPendingOperations"
      :operation="dangerStore.pendingOperations[0]"
      :host-count="selectedHosts.length"
      estimated-impact="此操作将影响多台主机"
      @confirm="handleDangerConfirm"
      @cancel="handleDangerCancel"
    />

    <!-- 高级: 风险警告 -->
    <RiskWarningDialog
      v-model:visible="showRisks"
      :risks="systemRisks"
      @view-details="handleViewRiskDetails"
      @resolve="handleResolveRisk"
    />

    <!-- 趋势预警气泡 -->
    <TrendWarningBubble
      v-for="warning in activeWarnings"
      :key="warning.id"
      v-bind="warning"
      @dismiss="dismissWarning(warning.id)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import ChatWindow from '@/components/chat/ChatWindow.vue'
import DangerConfirmDialog from '@/components/chat/DangerConfirmDialog.vue'
import RiskWarningDialog from '@/components/chat/RiskWarningDialog.vue'
import TrendWarningBubble from '@/components/chat/TrendWarningBubble.vue'
import { useDangerStore } from '@/stores/danger'
import { useChatStore } from '@/stores/chat'

const dangerStore = useDangerStore()
const chatStore = useChatStore()

const showRisks = ref(false)
const systemRisks = ref([])
const activeWarnings = ref([])
const selectedHosts = computed(() => chatStore.selectedHostIds)

const handleDangerConfirm = (operation) => {
  dangerStore.recordConfirm(operation, true)
  // 执行操作
}

const handleDangerCancel = () => {
  const operation = dangerStore.pendingOperations[0]
  if (operation) {
    dangerStore.recordConfirm(operation, false)
  }
}
</script>
```

---

## 6. 性能优化建议

### 虚拟滚动长列表

```vue
<template>
  <div class="host-list">
    <RecycleScroller
      :items="hosts"
      :item-size="48"
      key-field="id"
      v-slot="{ item }"
    >
      <HostTagBadge
        :host-id="item.id"
        :status="item.status"
        :result="item.result"
      />
    </RecycleScroller>
  </div>
</template>

<script setup lang="ts">
import { RecycleScroller } from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
</script>
```

### 组件懒加载

```typescript
// router/index.ts
const routes = [
  {
    path: '/prometheus',
    component: () => import('@/views/PrometheusMonitor.vue')
  },
  {
    path: '/health',
    component: () => import('@/views/HealthDashboard.vue')
  }
]
```

---

## 7. 类型定义参考

所有组件的完整类型定义位于：

- `C:\Users\hrp\Downloads\ai-pro\web\src\stores\danger.ts`
- `C:\Users\hrp\Downloads\ai-pro\web\src\stores\panel.ts`
- `C:\Users\hrp\Downloads\ai-pro\web\src\stores\prometheus.ts`

---

## 8. 样式定制

所有组件支持通过 CSS Variables 定制样式：

```scss
:root {
  --color-danger: #f56c6c;
  --color-warning: #e6a23c;
  --color-success: #67c23a;
  --radius-md: 8px;
  --shadow-md: 0 4px 8px rgba(0, 0, 0, 0.12);
}
```

---

## 总结

本指南涵盖了所有新开发组件的使用方法。所有组件都：

- ✅ 完整的 TypeScript 类型支持
- ✅ 响应式设计（移动端适配）
- ✅ 加载和错误状态处理
- ✅ 可访问性支持
- ✅ 性能优化
- ✅ 可测试

完整的组件代码位于：
- `C:\Users\hrp\Downloads\ai-pro\web\src\components\chat\`
- `C:\Users\hrp\Downloads\ai-pro\web\src\components\prometheus\`
- `C:\Users\hrp\Downloads\ai-pro\web\src\composables\`
- `C:\Users\hrp\Downloads\ai-pro\web\src\stores\`
